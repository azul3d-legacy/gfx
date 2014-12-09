// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build 386 amd64

package window

import (
	"runtime"

	"azul3d.org/native/glfw.v4"
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

		// Signals shutdown to the assetLoader goroutine.
		exit chan error
	}
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
		case <-asset.exit:
			// Destroy the device while the window is alive and OpenGL context
			// is active.
			asset.glfwDevice.Destroy()

			// Release context, before destroying window.
			glfw.DetachCurrentContext()

			// Destroy window and unlock the thread.
			err := asset.Window.Destroy()
			runtime.UnlockOSThread()

			// Signal completion.
			asset.exit <- err
			return

		case fn := <-exec:
			fn()
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
	err = glfw.WindowHint(glfw.Visible, 0)
	if err != nil {
		return err
	}
	asset.Window, err = glfw.CreateWindow(128, 128, "assets", nil, nil)
	if err != nil {
		return err
	}

	// Create the asset device.
	asset.exit = make(chan error)
	asset.Window.MakeContextCurrent()
	asset.glfwDevice, err = glfwNewDevice()
	if err != nil {
		return err
	}
	glfw.DetachCurrentContext()

	go assetLoader()

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
	asset.exit <- nil
	err := <-asset.exit
	if err != nil {
		return err
	}

	// Terminate GLFW now.
	err = glfw.Terminate()
	glfwInit = false
	return err
}
