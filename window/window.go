// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package window

import (
	"errors"
	"log"
	"math"
	"sync"

	"azul3d.org/gfx.v2-dev"
	"azul3d.org/keyboard.v2-dev"
	"azul3d.org/mouse.v2-dev"
)

// EventMask is a bitmask of event types. They can be combined, for instance:
//
//  mask := GenericEvents
//  mask |= MouseEvents
//  mask |= KeyboardEvents
//
// would select generic, mouse, and keyboard events.
type EventMask uint32

const (
	// NoEvents is a event mask matching no events at all.
	NoEvents EventMask = 0

	// CloseEvents is a event mask matching window.Close events.
	CloseEvents EventMask = 1 << 0

	// DamagedEvents is a event mask matching window.Damaged events.
	DamagedEvents EventMask = 1 << 1

	// CursorMovedEvents is a event mask matching window.CursorMoved events.
	CursorMovedEvents EventMask = 1 << 2

	// CursorEnterEvents is a event mask matching window.CursorEnter events.
	CursorEnterEvents EventMask = 1 << 3

	// CursorExitEvents is a event mask matching window.CursorExit events.
	CursorExitEvents EventMask = 1 << 4

	// MinimizedEvents is a event mask matching window.Minimized events.
	MinimizedEvents EventMask = 1 << 5

	// RestoredEvents is a event mask matching window.Restored events.
	RestoredEvents EventMask = 1 << 6

	// GainedFocusEvents is a event mask matching window.GainedFocus events.
	GainedFocusEvents EventMask = 1 << 7

	// LostFocusEvents is a event mask matching window.LostFocus events.
	LostFocusEvents EventMask = 1 << 8

	// MovedEvents is a event mask matching window.Moved events.
	MovedEvents EventMask = 1 << 9

	// ResizedEvents is a event mask matching window.Resized events.
	ResizedEvents EventMask = 1 << 10

	// FramebufferResizedEvents is a event mask matching
	// window.FramebufferResized events.
	FramebufferResizedEvents EventMask = 1 << 11

	// ItemsDroppedEvents is a event mask matching window.ItemsDropped events.
	ItemsDroppedEvents EventMask = 1 << 12

	// MouseEvents is a event mask matching mouse.Event events.
	MouseEvents EventMask = 1 << 13

	// MouseScrolledEvents is a event mask matching mouse.Scrolled events.
	MouseScrolledEvents EventMask = 1 << 14

	// KeyboardTypedEvents is a event mask matching keyboard.TypedEvent events.
	KeyboardTypedEvents EventMask = 1 << 15

	// KeyboardStateEvents is a event mask matching keyboard.StateEvent events.
	KeyboardStateEvents EventMask = 1 << 16

	// AllEvents is a event mask matching all possible events.
	AllEvents = EventMask(math.MaxUint32)
)

// Clipboard is the interface describing a system's clipboard. Grab a clipboard
// from a window (some platforms don't support clipboard access):
//
//  clip, ok := win.(window.Clipboard)
//  if ok {
//      // We have clipboard access.
//      clip.SetClipboard("Hello World!")
//  }
//
type Clipboard interface {
	// SetClipboard sets the clipboard string.
	SetClipboard(clipboard string)

	// Clipboard returns the clipboard string.
	Clipboard() string
}

// Window represents a single window that graphics can be drawn to. The window
// is safe for use concurrently from multiple goroutines.
type Window interface {
	// Props returns the window's properties.
	Props() *Props

	// Request makes a request to use a new set of properties, p. It is then
	// recommended to make changes to the window using something like:
	//
	//  props := window.Props()
	//  props.SetTitle("Hello World!")
	//  props.SetSize(640, 480)
	//  window.Request(props)
	//
	// Interpretation of the given properties is left strictly up to the
	// platform dependent implementation (for instance, on Android you cannot
	// set the window's size, so instead a request for this is simply ignored.
	Request(p *Props)

	// Keyboard returns a keyboard watcher for the window. It can be used to
	// tell if certain keyboard buttons are currently held down, for instance:
	//
	//  if w.Keyboard().Down(keyboard.W) {
	//      fmt.Println("The W key is currently held down")
	//  }
	//
	Keyboard() *keyboard.Watcher

	// Mouse returns a mouse watcher for the window. It can be used to tell if
	// certain mouse buttons are currently held down, for instance:
	//
	//  if w.Mouse().Down(mouse.Left) {
	//      fmt.Println("The left mouse button is currently held down")
	//  }
	//
	Mouse() *mouse.Watcher

	// Notify causes the window to relay window events to ch based on the event
	// mask.
	//
	// The special event mask NoEvents causes the window to stop relaying any
	// events to the given channel. You should always perform this action when
	// you are done using the event channel.
	//
	// The window will not block sending events to ch: the caller must ensure
	// that ch has a sufficient amount of buffer space to keep up with the
	// event rate.
	//
	// If you only expect to receive a single event like Close then a buffer
	// size of one is acceptable.
	//
	// You are allowed to make multiple calls to this method with the same
	// channel, if you do then the same event will be sent over the channel
	// multiple times. When you no longer want the channel to receive events
	// then call this function again with NoEvents:
	//
	//  w.Notify(ch, NoEvents)
	//
	// Multiple calls to Events with different channels works as you would
	// expect, each channel receives a copy of the events independent of other
	// ones.
	//
	// Warning: Many people use high-precision mice, some which can reach well
	// above 1000hz, so for cursor events a generous buffer size (256, etc) is
	// highly recommended.
	//
	// Warning: Depending on the operating system, window manager, etc, some of
	// the events may never be sent or may only be sent sporadically, so plan
	// for this.
	Notify(ch chan<- Event, m EventMask)

	// Close closes the window, it must be called or else the main loop (and
	// inheritely, the application) will not exit.
	Close()
}

// numWindows maintains the count of open windows.
var numWindows struct {
	sync.Mutex
	N int
}

// Num tells the number of windows that are currently open, after adding the
// given integer to the count.
//
// Query how many windows are currently open:
//
//  open := window.Num(0)
//  fmt.Println(open) // e.g. "1"
//
// Add two windows to the current count (1) and fetch the new value:
//
//  open := window.Num(2)
//  fmt.Println(open) // e.g. "3"
//
// Implementors of the Window interface are the only ones that should modify
// the window count.
//
// This function is safe for access from multiple goroutines concurrently.
func Num(n int) int {
	numWindows.Lock()
	numWindows.N += n
	n = numWindows.N
	numWindows.Unlock()
	return n
}

// ErrSingleWindow is returned by New if you attempt to create multiple windows
// on a system that does not support it.
var ErrSingleWindow = errors.New("only a single window is allowed")

// New creates a new window, and is safe to call from any goroutine.
//
// If you just want to create a single window, use the simpler Run function
// instead.
//
// If the properties, p, are nil then DefaultProps is used instead.
//
// Interpretation of the properties is left strictly up to the platform
// dependent implementation (for instance, on Android you cannot set a window's
// size, so it is ignored).
//
// If you attempt to create multiple windows and the system does not allow it,
// then ErrSingleWindow will be returned (e.g. on mobile platforms where there
// is no concept of multiple windows).
//
// If any error is returned, the window could not be created, and the returned
// window and device are nil.
//
// New requests several operations be run on the main loop internally, because
// of this it cannot be run on the main thread itself. That is, MainLoop must
// be running for New to complete.
//
// The following code works fine, because New is run in a seperate goroutine:
//
//  func main() {
//      go func() {
//          // New runs in a seperate goroutine, after MainLoop has started.
//          w, d, err := window.New(nil)
//          ... use w, d, handle err ...
//      }()
//      window.MainLoop()
//  }
//
// The following code does not work, a deadlock occurs because MainLoop is
// called after New, and New cannot complete unless MainLoop is running.
//
//  func main() {
//      // Won't ever complete: the main loop isn't running yet!
//      w, d, err := window.New(nil)
//      ... use w, d, handle err ...
//      window.MainLoop()
//  }
//
func New(p *Props) (w Window, d gfx.Device, err error) {
	if p == nil {
		p = DefaultProps
	}

	// Run doNew on the main loop.
	done := make(chan struct{}, 1)
	MainLoopChan <- func() {
		// Create a new window via the platform-specific backend.
		w, d, err = doNew(p)
		done <- struct{}{}
	}
	<-done

	// Return if any error occured.
	if err != nil {
		return nil, nil, err
	}

	// No error occured, increment the number of open windows and return.
	Num(1)
	return w, d, err
}

// Run opens a window with the given properties and runs the given graphics
// loop in a separate goroutine.
//
// This function must be called only from the program's main function (other
// work should be done in other goroutines):
//
//  func main() {
//      window.Run(gfxLoop, nil)
//  }
//
// For more documentation about the behavior of Run, see the New function.
func Run(gfxLoop func(w Window, d gfx.Device), p *Props) {
	if gfxLoop == nil {
		panic("window: nil graphics loop function!")
	}

	// Create a new window in a seperate goroutine, because the New function
	// requires that the main loop be running before it can complete.
	go func() {
		// Create the window with the given properties.
		w, d, err := New(p)
		if err != nil {
			log.Fatal(err)
		}

		// If the gfxLoop panics, we should still close the window (or else the
		// user's screen resolution won't be restored, for example).
		defer func() {
			if r := recover(); r != nil {
				w.Close()
				panic(r)
			}
		}()

		// Enter the graphics loop.
		gfxLoop(w, d)
	}()

	// Enter the main loop now.
	MainLoop()
}
