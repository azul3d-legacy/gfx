// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build 386 amd64

package window

import (
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"azul3d.org/gfx.v2-dev"
	"azul3d.org/gfx.v2-dev/internal/tag"
	"azul3d.org/gfx.v2-dev/internal/util"
	"azul3d.org/keyboard.v2-dev"
	"azul3d.org/mouse.v2-dev"
	"azul3d.org/native/glfw.v4"
)

// intBool returns 0 or 1 depending on b.
func intBool(b bool) int {
	if b {
		return 1
	}
	return 0
}

// logError simply logs the error.
func logError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "window: %v\n", err)
	}
}

type glfwDevice interface {
	gfx.Device
	Exec() chan func() bool
	UpdateBounds(bounds image.Rectangle)
	SetDebugOutput(w io.Writer)
	Destroy()
}

// glfwWindow implements the Window interface using a GLFW backend.
type glfwWindow struct {
	// The below variables are read-only after initialization of this struct,
	// and thus do not use the RWMutex.
	*notifier
	mouse                                              *mouse.Watcher
	keyboard                                           *keyboard.Watcher
	extWGLEXTSwapControlTear, extGLXEXTSwapControlTear bool
	exit, rebuild                                      chan struct{}

	// The below variables are read-write after initialization of this struct,
	// and as such must only be modified under the RWMutex.
	sync.RWMutex
	swapper                  *util.Swapper
	props, last              *Props
	device                   glfwDevice
	window                   *glfw.Window
	monitor                  *glfw.Monitor
	beforeFullscreen         [2]int // Window size before fullscreen.
	lastCursorX, lastCursorY float64
	closed                   bool
}

// Props implements the Window interface.
func (w *glfwWindow) Props() *Props {
	w.RLock()
	props := w.props
	w.RUnlock()
	return props
}

// Request implements the Window interface.
func (w *glfwWindow) Request(p *Props) {
	MainLoopChan <- func() {
		w.Lock()
		w.useProps(p, false)
		w.Unlock()
	}
}

// Keyboard implements the Window interface.
func (w *glfwWindow) Keyboard() *keyboard.Watcher {
	return w.keyboard
}

// Mouse implements the Window interface.
func (w *glfwWindow) Mouse() *mouse.Watcher {
	return w.mouse
}

// SetClipboard implements the Clipboard interface.
func (w *glfwWindow) SetClipboard(clipboard string) {
	MainLoopChan <- func() {
		w.Lock()
		logError(w.window.SetClipboardString(clipboard))
		w.window.SetClipboardString(clipboard)
		w.Unlock()
	}
}

// Clipboard implements the Clipboard interface.
func (w *glfwWindow) Clipboard() string {
	w.RLock()
	var (
		str string
		err error
	)
	w.waitFor(func() {
		str, err = w.window.GetClipboardString()
	})
	w.RUnlock()
	logError(err)
	return str
}

// Close implements the Window interface.
func (w *glfwWindow) Close() {
	// Protect against double-closes.
	w.Lock()
	if w.closed {
		w.Unlock()
		return
	}
	w.closed = true
	w.Unlock()

	// Signal to the window of it's closing.
	w.exit <- struct{}{}
}

// waitFor runs f on the main thread and waits for the function to complete.
func (w *glfwWindow) waitFor(f func()) {
	done := make(chan bool, 1)
	MainLoopChan <- func() {
		f()
		done <- true
	}
	<-done
}

// updateTitle updates the window title and accounts for "{FPS}" strings.
//
// It may only be called on the main thread, and under the presence of the
// window's read lock.
func (w *glfwWindow) updateTitle() {
	fps := fmt.Sprintf("%dFPS", int(math.Ceil(w.device.Clock().FrameRate())))
	title := strings.Replace(w.props.Title(), "{FPS}", fps, 1)
	logError(w.window.SetTitle(title))
}

// useProps sets the GLFW window to reflect the given window properties. It
// detects the properties that have not changed since the last call to
// useProps and, if force == false, omits them for efficiency.
//
// It may only be called on the main thread, and under the presence of the
// window's write lock.
func (w *glfwWindow) useProps(p *Props, force bool) {
	w.props = p

	// Runs f without the currently held lock. Because some functions cause an
	// event to be generated, calling the event callback and causing a deadlock.
	withoutLock := func(f func()) {
		w.Unlock()
		f()
		w.Lock()
	}
	win := w.window

	// GLFW doesn't yet support switching at runtime between fullscreen and
	// windowed mode. We employ a more traditional workaround here which is
	// destroying and rebuilding the window and it's associated device (i.e.
	// OpenGL context). It is hence important that this be the first operation.
	//
	// We can do this without losing time, as assets are stored in the shared
	// asset context -- not in this window's context.
	fullscreen := w.props.Fullscreen()
	lastFullscreen := w.last.Fullscreen()
	if fullscreen != lastFullscreen {
		w.last.SetFullscreen(fullscreen)

		// If we're not switching to fullscreen, restore the window size from
		// before we entered fullscreen.
		if !fullscreen {
			w.props.SetSize(w.beforeFullscreen[0], w.beforeFullscreen[1])
		}

		// Signal to the window goroutine that we need a window rebuild now, it
		// will call useProps on it's own to initialize the new window.
		w.rebuild <- struct{}{}
		return
	}

	// Set each property, only if it differs from the last known value for that
	// property.

	w.updateTitle()

	// Window Size.
	width, height := w.props.Size()
	lastWidth, lastHeight := w.last.Size()
	if force || width != lastWidth || height != lastHeight {
		// If we're not switching to fullscreen, save the window size as it was
		// for restoration after we've exited fullscreen later.
		if !fullscreen {
			w.beforeFullscreen = [2]int{width, height}
		}

		w.last.SetSize(width, height)
		withoutLock(func() {
			logError(win.SetSize(width, height))
		})
	}

	// Window Position.
	x, y := w.props.Pos()
	lastX, lastY := w.last.Pos()
	if (force || x != lastX || y != lastY) && !fullscreen {
		w.last.SetPos(x, y)
		if x == -1 && y == -1 {
			vm, err := w.monitor.GetVideoMode()
			logError(err)
			if err == nil {
				x = (vm.Width / 2) - (width / 2)
				y = (vm.Height / 2) - (height / 2)
			}
		}
		withoutLock(func() {
			logError(win.SetPosition(x, y))
		})
	}

	// Cursor Position.
	cursorX, cursorY := w.props.CursorPos()
	lastCursorX, lastCursorY := w.last.CursorPos()
	if force || cursorX != lastCursorX || cursorY != lastCursorY {
		w.last.SetCursorPos(cursorX, cursorY)
		if cursorX != -1 && cursorY != -1 {
			withoutLock(func() {
				logError(win.SetCursorPosition(cursorX, cursorY))
			})
		}
	}

	// Window Visibility.
	visible := w.props.Visible()
	if force || w.last.Visible() != visible {
		w.last.SetVisible(visible)
		withoutLock(func() {
			if visible {
				logError(win.Show())
			} else {
				logError(win.Hide())
			}
		})
	}

	// Window Minimized.
	minimized := w.props.Minimized()
	if force || w.last.Minimized() != minimized {
		w.last.SetMinimized(minimized)
		withoutLock(func() {
			if minimized {
				logError(win.Iconify())
			} else {
				logError(win.Restore())
			}
		})
	}

	// Vertical sync mode.
	vsync := w.props.VSync()
	if force || w.last.VSync() != vsync {
		w.last.SetVSync(vsync)

		// Determine the swap interval and set it.
		var swapInterval int
		if vsync {
			// We want vsync on, we will use adaptive vsync if we have it, if
			// not we will use standard vsync.
			if w.extWGLEXTSwapControlTear || w.extGLXEXTSwapControlTear {
				// We can use adaptive vsync via a swap interval of -1.
				swapInterval = -1
			} else {
				// No adaptive vsync, use standard then.
				swapInterval = 1
			}
		}
		logError(glfw.SwapInterval(swapInterval))
	}

	// The following cannot be changed via GLFW post window creation -- and
	// they are not deemed significant enough to warrant rebuilding the window.
	//
	//  Focused
	//  Resizable
	//  Decorated
	//  AlwaysOnTop (via GLFW_FLOATING)
	//

	// Cursor Mode.
	grabbed := w.props.CursorGrabbed()
	if force || w.last.CursorGrabbed() != grabbed {
		w.last.SetCursorGrabbed(grabbed)

		// Reset both last cursor values to the callback can identify the
		// large/fake delta.
		w.lastCursorX = math.Inf(-1)
		w.lastCursorY = math.Inf(-1)

		// Set input mode.
		withoutLock(func() {
			if grabbed {
				logError(w.window.SetInputMode(glfw.Cursor, glfw.CursorDisabled))
			} else {
				logError(w.window.SetInputMode(glfw.Cursor, glfw.CursorNormal))
			}
		})
	}
}

// initCallbacks sets a callback handler for each GLFW window event.
//
// It may only be called on the main thread, and under the presence of the
// window's read lock.
func (w *glfwWindow) initCallbacks() {
	// Close event.
	w.window.SetCloseCallback(func(gw *glfw.Window) {
		// If they want us to close the window, then close the window.
		if w.Props().ShouldClose() {
			w.Close()

			// Return so we don't give people the idea that they can rely on
			// Close event below to cleanup things.
			return
		}
		w.sendEvent(Close{T: time.Now()}, CloseEvents)
	})

	// Damaged event.
	w.window.SetRefreshCallback(func(gw *glfw.Window) {
		w.sendEvent(Damaged{T: time.Now()}, DamagedEvents)
	})

	// Minimized and Restored events.
	w.window.SetIconifyCallback(func(gw *glfw.Window, iconify bool) {
		// Store the minimized/restored state.
		w.RLock()
		w.last.SetMinimized(iconify)
		w.props.SetMinimized(iconify)
		w.RUnlock()

		// Send the proper event.
		if iconify {
			w.sendEvent(Minimized{T: time.Now()}, MinimizedEvents)
			return
		}
		w.sendEvent(Restored{T: time.Now()}, RestoredEvents)
	})

	// FocusChanged event.
	w.window.SetFocusCallback(func(gw *glfw.Window, focused bool) {
		// Store the focused state.
		w.RLock()
		w.last.SetFocused(focused)
		w.props.SetFocused(focused)
		w.RUnlock()

		// Send the proper event.
		if focused {
			w.sendEvent(GainedFocus{T: time.Now()}, GainedFocusEvents)
			return
		}
		w.sendEvent(LostFocus{T: time.Now()}, LostFocusEvents)
	})

	// Moved event.
	w.window.SetPositionCallback(func(gw *glfw.Window, x, y int) {
		// Store the position state.
		w.RLock()
		w.last.SetPos(x, y)
		if w.last.Fullscreen() {
			// If we're in fullscreen, we don't expose the window position.
			w.RUnlock()
			return
		}
		w.props.SetPos(x, y)
		w.RUnlock()
		w.sendEvent(Moved{X: x, Y: y, T: time.Now()}, MovedEvents)
	})

	// Resized event.
	w.window.SetSizeCallback(func(gw *glfw.Window, width, height int) {
		// Store the size state.
		w.Lock()
		if !w.last.Fullscreen() {
			// If we're not currently in fullscreen, save the window size as it
			// was for restoration after we've exited fullscreen later.
			w.beforeFullscreen = [2]int{width, height}
		}
		w.last.SetSize(width, height)
		w.props.SetSize(width, height)
		w.Unlock()
		w.sendEvent(Resized{
			Width:  width,
			Height: height,
			T:      time.Now(),
		}, ResizedEvents)
	})

	// FramebufferResized event.
	w.window.SetFramebufferSizeCallback(func(gw *glfw.Window, width, height int) {
		// Store the framebuffer size state.
		w.RLock()
		w.last.SetFramebufferSize(width, height)
		w.props.SetFramebufferSize(width, height)
		w.RUnlock()

		// Update device's bounds.
		w.device.UpdateBounds(image.Rect(0, 0, width, height))

		// Send the event.
		w.sendEvent(FramebufferResized{
			Width:  width,
			Height: height,
			T:      time.Now(),
		}, FramebufferResizedEvents)
	})

	// Dropped event.
	w.window.SetDropCallback(func(gw *glfw.Window, items []string) {
		w.sendEvent(ItemsDropped{Items: items, T: time.Now()}, ItemsDroppedEvents)
	})

	// CursorMoved event.
	w.window.SetCursorPositionCallback(func(gw *glfw.Window, x, y float64) {
		// Store the cursor position state.
		w.RLock()
		grabbed := w.props.CursorGrabbed()
		if grabbed {
			// Store/swap last cursor values. Note: It's safe to modify
			// lastCursorX/Y with just w.RLock because they are only modified
			// in this callback on the main thread.
			lastX := w.lastCursorX
			lastY := w.lastCursorY
			w.lastCursorX = x
			w.lastCursorY = y

			// First cursor position callback since grab occured, avoid the
			// large/fake delta.
			if lastX == math.Inf(-1) && lastY == math.Inf(-1) {
				w.RUnlock()
				return
			}

			// Calculate cursor delta.
			x = x - lastX
			y = y - lastY
		} else {
			// Store cursor position.
			w.last.SetCursorPos(x, y)
			w.props.SetCursorPos(x, y)
		}
		w.RUnlock()

		// Send proper event.
		w.sendEvent(CursorMoved{
			X:     x,
			Y:     y,
			Delta: grabbed,
			T:     time.Now(),
		}, CursorMovedEvents)
	})

	// CursorEnter and CursorExit events.
	w.window.SetCursorEnterCallback(func(gw *glfw.Window, enter bool) {
		// TODO(slimsag): expose *within window* state, but not via Props.
		if enter {
			w.sendEvent(CursorEnter{T: time.Now()}, CursorEnterEvents)
			return
		}
		w.sendEvent(CursorExit{T: time.Now()}, CursorExitEvents)
	})

	// keyboard.Typed
	w.window.SetCharacterCallback(func(gw *glfw.Window, r rune) {
		w.sendEvent(keyboard.Typed{S: string(r), T: time.Now()}, KeyboardTypedEvents)
	})

	// keyboard.ButtonEvent
	w.window.SetKeyCallback(func(gw *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Repeat {
			return
		}

		// Convert GLFW event.
		k := convertKey(key)
		s := convertKeyAction(action)
		r := uint64(scancode)

		// Update keyboard watcher.
		w.keyboard.SetState(k, s)
		w.keyboard.SetRawState(r, s)

		// Send the event.
		w.sendEvent(keyboard.ButtonEvent{
			T:     time.Now(),
			Key:   k,
			State: s,
			Raw:   r,
		}, KeyboardButtonEvents)
	})

	// mouse.ButtonEvent
	w.window.SetMouseButtonCallback(func(gw *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		// Convert GLFW event.
		b := convertMouseButton(button)
		s := convertMouseAction(action)

		// Update mouse watcher.
		w.mouse.SetState(b, s)

		// Send the event.
		w.sendEvent(mouse.ButtonEvent{
			T:      time.Now(),
			Button: b,
			State:  s,
		}, MouseEvents)
	})

	// mouse.Scrolled event.
	w.window.SetScrollCallback(func(gw *glfw.Window, x, y float64) {
		w.sendEvent(mouse.Scrolled{
			T: time.Now(),
			X: x,
			Y: y,
		}, MouseScrolledEvents)
	})
}

// run is the goroutine responsible for manging this window.
func (w *glfwWindow) run() {
	// A ticker for updating the window title with the new FPS each second.
	updateFPS := time.NewTicker(1 * time.Second)
	defer updateFPS.Stop()

	exec := w.device.Exec()

	// OpenGL function calls must occur in the same thread.
	runtime.LockOSThread()

	// Make the window's context the current one.
	w.window.MakeContextCurrent()

	cleanup := func() {
		// Destroy the device.
		w.device.Destroy()

		// Release the context.
		logError(glfw.DetachCurrentContext())

		// Destroy the window on the main thread.
		MainLoopChan <- func() {
			logError(w.window.Destroy())
		}
	}

	for {
		select {
		case <-w.exit:
			cleanup()

			// Decrement the number of open windows by one.
			windowCount := Num(-1)

			// Signal that a window has closed to the main loop.
			MainLoopChan <- nil

			// Unlock the thread.
			runtime.UnlockOSThread()

			if windowCount == 0 {
				// No more windows are open, so de-initialize.
				MainLoopChan <- func() {
					logError(doExit())
				}
			}
			return

		case <-w.rebuild:
			// We need to rebuild the window and it's context. Signal to the
			// swapper that it should yield when it can.
			w.swapper.Yield <- struct{}{}

			// Execute functions on the existing window until the swapper
			// yields for us.
		sr:
			for {
				select {
				case fn := <-exec:
					// Execute the device's render function.
					if renderedFrame := fn(); renderedFrame {
						// Swap OpenGL buffers.
						logError(w.window.SwapBuffers())
					}

				case <-w.swapper.Swap:
					// The swapper has yielded for us. Cleanup the device,
					// window, and OpenGL context.
					w.Lock()
					cleanup()

					// Rebuild the window in the main thread.
					w.waitFor(func() {
						logError(w.build())
					})

					// Make the new window's context the active one.
					w.window.MakeContextCurrent()

					// Rebind the exec variable that we use, unlock the window.
					exec = w.device.Exec()
					w.Unlock()

					// Perform the swap of the underlying device and break exit
					// the rebuild loop.
					w.swapper.Swap <- w.device
					break sr
				}
			}

		case <-updateFPS.C:
			// Update title with FPS.
			MainLoopChan <- func() {
				w.Lock()
				w.updateTitle()
				w.Unlock()
			}

		case fn := <-exec:
			// Execute the device's render function.
			if renderedFrame := fn(); renderedFrame {
				// Swap OpenGL buffers.
				logError(w.window.SwapBuffers())
			}
		}
	}
}

// build builds the underlying GLFW window. It is used both at window init time
// (see doNew) and when rebuilding the window for fullscreen switching (which
// GLFW doesn't yet support itself).
//
// It may only be called on the main thread, and under the presence of the
// window's write lock.
func (w *glfwWindow) build() error {
	var (
		dstMonitor          *glfw.Monitor
		err                 error
		p                   = w.props
		dstWidth, dstHeight = p.Size()
	)

	// Specify the primary monitor if we want fullscreen, store the monitor
	// regardless for centering the window.
	w.monitor, err = glfw.GetPrimaryMonitor()
	if err != nil {
		return err
	}
	if p.Fullscreen() {
		dstMonitor = w.monitor
		w.beforeFullscreen = [2]int{dstWidth, dstHeight}

		// TODO(slimsag): publish a way to get valid video modes instead of
		// assuming the monitor's one.
		vm, err := w.monitor.GetVideoMode()
		if err != nil {
			return err
		}
		dstWidth, dstHeight = vm.Width, vm.Height
		w.props.SetSize(dstWidth, dstHeight)
		w.last.SetSize(dstWidth, dstHeight)
	} else {
		w.beforeFullscreen = [2]int{dstWidth, dstHeight}
	}

	// Hint standard properties (note visibility is always false, we show the
	// window later after moving it).
	prec := p.Precision()
	hints := map[glfw.Hint]int{
		glfw.Visible: 0,
		// TODO(slimsag): once GLFW 3.1 is released we can use these hints:
		//glfw.Focused: intBool(p.Focused()),
		//glfw.Iconified: intBool(p.Minimized()),
		glfw.Resizable:           intBool(p.Resizable()),
		glfw.Decorated:           intBool(p.Decorated()),
		glfw.AutoIconify:         1,
		glfw.Floating:            intBool(p.AlwaysOnTop()),
		glfw.RedBits:             int(prec.RedBits),
		glfw.GreenBits:           int(prec.GreenBits),
		glfw.BlueBits:            int(prec.BlueBits),
		glfw.AlphaBits:           int(prec.AlphaBits),
		glfw.DepthBits:           int(prec.DepthBits),
		glfw.StencilBits:         int(prec.StencilBits),
		glfw.Samples:             prec.Samples,
		glfw.SRGBCapable:         1,
		glfw.OpenGLDebugContext:  intBool(tag.Gfxdebug),
		glfw.ContextVersionMajor: glfwContextVersionMajor,
		glfw.ContextVersionMinor: glfwContextVersionMinor,
		glfw.ClientAPI:           glfwClientAPI,
	}
	for hint, value := range hints {
		err = glfw.WindowHint(hint, value)
		if err != nil {
			return err
		}
	}

	// Create the window.
	asset.withoutContext <- nil // Ask to disable the asset context.
	<-asset.withoutContext      // Wait for disable to complete.
	w.window, err = glfw.CreateWindow(dstWidth, dstHeight, p.Title(), dstMonitor, asset.Window)
	asset.withoutContext <- nil // Give back the asset context.
	if err != nil {
		return err
	}

	// OpenGL context must be active.
	err = w.window.MakeContextCurrent()
	if err != nil {
		return err
	}

	// Create the device.
	d, err := glfwNewDevice(keepState(), share(asset.glfwDevice))
	if err != nil {
		return err
	}
	w.device = d

	// Write device debug output (shader errors, etc) to stderr.
	d.SetDebugOutput(os.Stderr)

	// Test for adaptive vsync extensions.
	w.extWGLEXTSwapControlTear, err = glfw.ExtensionSupported("WGL_EXT_swap_control_tear")
	if err != nil {
		return err
	}
	w.extGLXEXTSwapControlTear, err = glfw.ExtensionSupported("GLX_EXT_swap_control_tear")
	if err != nil {
		return err
	}

	// Setup callbacks and the window.
	w.initCallbacks()
	w.useProps(p, true)

	// Done with OpenGL things on this window, for now.
	return glfw.DetachCurrentContext()
}

func doNew(p *Props) (Window, gfx.Device, error) {
	// Initialize the hidden asset window if needed.
	if err := doInit(); err != nil {
		return nil, nil, err
	}

	// Initialize window.
	w := &glfwWindow{
		notifier: &notifier{},
		props:    p,
		last:     NewProps(),
		mouse:    mouse.NewWatcher(),
		keyboard: keyboard.NewWatcher(),
		exit:     make(chan struct{}, 1),
		rebuild:  make(chan struct{}),
	}

	// Build the actual GLFW window.
	w.Lock()
	if err := w.build(); err != nil {
		return nil, nil, err
	}
	w.Unlock()

	w.swapper = util.NewSwapper(w.device)

	// Spawn the goroutine responsible for running the window.
	go w.run()

	return w, w.swapper, nil
}
