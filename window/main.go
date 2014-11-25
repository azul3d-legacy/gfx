// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package window

import "runtime"

// The communicative main loop pattern used by this package is outlined lightly
// in this blog post:
//
// http://slimsag.blogspot.com/2014/11/go-solves-busy-waiting.html
//
// For details about the main thread and LockOSThread, see:
//
// https://groups.google.com/forum/#!topic/golang-nuts/6Rh4e0wvhn4

func init() {
	// Lock main.main onto the main OS thread. This allows each function in the
	// MainLoop to be executed on specifically the main OS thread, which is
	// required by e.g. Cocoa on OS X (which GLFW uses internally).
	runtime.LockOSThread()
}

// MainLoopChan is a channel of functions over which each window created
// through this package will request for functions to be executed.
//
// When any window is closed, it will send nil over this channel. The main loop
// should handle this case by checking the number of open windows (see the
// Num function). If no windows are left open, the main loop should exit.
var MainLoopChan = make(chan func())

// MainLoop enters the main loop, executing the main loop functions received
// from MainLoopChan until no windows are left open.
//
// This function must be called only from the program's main function (other
// work should be done in other goroutines):
//
//  func main() {
//      window.MainLoop()
//  }
//
// By implementing MainLoop yourself, you can run other functions on the main
// OS thread (for instance Cocoa API's on OS X).
func MainLoop() {
	for {
		select {
		case f := <-MainLoopChan:
			// If the function is nil then a window has closed. We should check
			// if the number of open windows is zero, and if so, the main loop
			// can end.
			if f == nil && Num(0) == 0 {
				return
			}

			// If the function is non-nil, execute it.
			if f != nil {
				f()
			}
		}
	}
}
