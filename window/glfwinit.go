// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build 386 amd64

package window

import (
	"os"
	"runtime"
	"time"

	"azul3d.org/native/glfw.v5"
)

var (
	// Whether or not GLFW has been initialized yet (only modified on the main
	// thread).
	glfwInit bool

	asset struct {
		// A hidden window which is used for it's context to own OpenGL assets
		// shared between multiple windows.
		*glfw.Window

		// The device of the hidden window, again just used to store assets.
		glfwDevice

		// Channel for executing functions without the context active.
		withoutContext chan func()

		// Signals shutdown to the assetLoader goroutine.
		exit chan struct{}
	}

	// Signals shutdown to the event poller goroutine.
	pollerExit chan struct{}
)

// assetLoader is the goroutine responsible for running the asset device.
func assetLoader() {
	exec := asset.glfwDevice.Exec()

	// OpenGL function calls must occur in the same thread.
	runtime.LockOSThread()

	// Make the window's context the current one.
	asset.Window.MakeContextCurrent()

	for {
		select {
		case <-asset.withoutContext:
			// Drop the context, signal back.
			glfw.DetachCurrentContext()
			asset.withoutContext <- nil

			// Grab the context and continue.
			<-asset.withoutContext
			asset.Window.MakeContextCurrent()

		case <-asset.exit:
			// Destroy the device while the window is alive and OpenGL context
			// is active.
			asset.glfwDevice.Destroy()

			// Release context, before destroying window.
			glfw.DetachCurrentContext()

			// Destroy window and unlock the thread.
			asset.Window.Destroy()
			runtime.UnlockOSThread()

			// Exit asset loader goroutine.
			return

		case fn := <-exec:
			fn()
		}
	}
}

// pollEvents submits a function to the main loop to poll for GLFW events at
// 120hz.
func pollEvents() {
	// Poll for events at 120hz.
	ticker := time.NewTicker(time.Second / 120)
	defer ticker.Stop()

	for {
		<-ticker.C

		// Consider exiting now.
		select {
		case <-pollerExit:
			return
		default:
		}

		// Poll for events in the main loop, or exit.
		select {
		case <-pollerExit:
			return
		case MainLoopChan <- glfw.PollEvents:
		}
	}
}

// doInit initializes GLFW and the hidden asset window/device, if not already
// initialized.
func doInit() error {
	if glfwInit {
		// Already initialized.
		return nil
	}

	// Initialize GLFW now.
	err := glfw.Init()
	if err != nil {
		return err
	}

	// Create the hidden asset window.
	glfw.WindowHint(glfw.Visible, 0)
	asset.Window, err = glfw.CreateWindow(128, 128, "assets", nil, nil)
	if err != nil {
		return err
	}

	// Create the asset device.
	asset.exit = make(chan struct{})
	asset.withoutContext = make(chan func())
	asset.Window.MakeContextCurrent()
	asset.glfwDevice, err = glfwNewDevice()
	if err != nil {
		return err
	}

	// Write device debug output (shader errors, etc) to stderr.
	asset.glfwDevice.SetDebugOutput(os.Stderr)
	glfw.DetachCurrentContext()

	go assetLoader()

	// Spawn the event poller.
	pollerExit = make(chan struct{}, 1)
	go pollEvents()

	glfwInit = true
	return nil
}

// doExit de-initializes GLFW and the hidden asset/window device, only if it
// is initialized.
func doExit() error {
	if !glfwInit {
		// Not even initialized.
		return nil
	}

	// Exit the assetLoader goroutine.
	asset.exit <- struct{}{}

	// Exit the event poller.
	pollerExit <- struct{}{}

	// Terminate GLFW now.
	glfw.Terminate()
	glfwInit = false
	return nil
}
