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
	"azul3d.org/gfx.v2-dev/internal/gfxdebug"
	"azul3d.org/keyboard.v1"
	"azul3d.org/mouse.v1"
	"azul3d.org/native/glfw.v4"
)

// TODO(slimsag): rebuild window when fullscreen/precision changes.

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
	exit                                               chan struct{}

	// The below variables are read-write after initialization of this struct,
	// and as such must only be modified under the RWMutex.
	sync.RWMutex
	props, last              *Props
	device                   glfwDevice
	window                   *glfw.Window
	monitor                  *glfw.Monitor
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
		w.useProps(p, false)
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
// It may only be called on the main thread, and it utilizes the window's write
// lock on it's own.
func (w *glfwWindow) useProps(p *Props, force bool) {
	w.Lock()
	defer w.Unlock()
	w.props = p

	// Runs f without the currently held lock. Because some functions cause an
	// event to be generated, calling the event callback and causing a deadlock.
	withoutLock := func(f func()) {
		w.Unlock()
		f()
		w.Lock()
	}
	win := w.window

	// Set each property, only if it differs from the last known value for that
	// property.

	w.updateTitle()

	// Window Size.
	width, height := w.props.Size()
	lastWidth, lastHeight := w.last.Size()
	if force || width != lastWidth || height != lastHeight {
		w.last.SetSize(width, height)
		withoutLock(func() {
			logError(win.SetSize(width, height))
		})
	}

	// Window Position.
	x, y := w.props.Pos()
	lastX, lastY := w.last.Pos()
	if force || x != lastX || y != lastY {
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
	// TODO(slimsag): consider these when rebuilding the window for Fullscreen
	// or Precision switches.
	//
	//  Focused
	//  Resizable
	//  Decorated
	//  AlwaysOnTop (via GLFW_FLOATING)

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
		w.props.SetPos(x, y)
		w.RUnlock()
		w.sendEvent(Moved{X: x, Y: y, T: time.Now()}, MovedEvents)
	})

	// Resized event.
	w.window.SetSizeCallback(func(gw *glfw.Window, width, height int) {
		// Store the size state.
		w.RLock()
		w.last.SetSize(width, height)
		w.props.SetSize(width, height)
		w.RUnlock()
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

	// keyboard.TypedEvent
	w.window.SetCharacterCallback(func(gw *glfw.Window, r rune) {
		w.sendEvent(keyboard.TypedEvent{Rune: r, T: time.Now()}, KeyboardTypedEvents)
	})

	// keyboard.StateEvent
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
		w.sendEvent(keyboard.StateEvent{
			T:     time.Now(),
			Key:   convertKey(key),
			State: convertKeyAction(action),
			Raw:   uint64(scancode),
		}, KeyboardStateEvents)
	})

	// mouse.Event
	w.window.SetMouseButtonCallback(func(gw *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		// Convert GLFW event.
		b := convertMouseButton(button)
		s := convertMouseAction(action)

		// Update mouse watcher.
		w.mouse.SetState(b, s)

		// Send the event.
		w.sendEvent(mouse.Event{
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

	for {
		select {
		case <-w.exit:
			// Destroy the device.
			w.device.Destroy()

			// Release the context.
			logError(glfw.DetachCurrentContext())

			// Destroy the window on the main thread.
			MainLoopChan <- func() {
				logError(w.window.Destroy())
			}

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

func doNew(p *Props) (Window, gfx.Device, error) {
	var (
		targetMonitor, monitor *glfw.Monitor
		err                    error
	)

	if err := doInit(); err != nil {
		return nil, nil, err
	}

	// Specify the primary monitor if we want fullscreen, store the monitor
	// regardless for centering the window.
	monitor, err = glfw.GetPrimaryMonitor()
	if err != nil {
		return nil, nil, err
	}
	if p.Fullscreen() {
		targetMonitor = monitor
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
		glfw.OpenGLDebugContext:  intBool(gfxdebug.Flag),
		glfw.ContextVersionMajor: glfwContextVersionMajor,
		glfw.ContextVersionMinor: glfwContextVersionMinor,
		glfw.ClientAPI:           glfwClientAPI,
	}
	for hint, value := range hints {
		err = glfw.WindowHint(hint, value)
		if err != nil {
			return nil, nil, err
		}
	}

	// Create the window.
	width, height := p.Size()
	window, err := glfw.CreateWindow(width, height, p.Title(), targetMonitor, asset.Window)
	if err != nil {
		return nil, nil, err
	}

	// OpenGL context must be active.
	err = window.MakeContextCurrent()
	if err != nil {
		return nil, nil, err
	}

	// Create the device.
	r, err := glfwNewDevice(keepState(), share(asset.glfwDevice))
	if err != nil {
		return nil, nil, err
	}

	// Write device debug output (shader errors, etc) to stdout.
	r.SetDebugOutput(os.Stderr)

	// Initialize window.
	w := &glfwWindow{
		notifier: &notifier{},
		props:    p,
		last:     NewProps(),
		mouse:    mouse.NewWatcher(),
		keyboard: keyboard.NewWatcher(),
		device:   r,
		window:   window,
		monitor:  monitor,
		exit:     make(chan struct{}, 1),
	}

	// Test for adaptive vsync extensions.
	w.extWGLEXTSwapControlTear, err = glfw.ExtensionSupported("WGL_EXT_swap_control_tear")
	if err != nil {
		return nil, nil, err
	}
	w.extGLXEXTSwapControlTear, err = glfw.ExtensionSupported("GLX_EXT_swap_control_tear")
	if err != nil {
		return nil, nil, err
	}

	// Setup callbacks and the window.
	w.initCallbacks()
	w.useProps(p, true)

	// Done with OpenGL things on this window, for now.
	err = glfw.DetachCurrentContext()
	if err != nil {
		return nil, nil, err
	}

	// Spawn the goroutine responsible for running the window.
	go w.run()
	return w, r, nil
}
