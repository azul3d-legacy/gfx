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
		// shared between render windows.
		*glfw.Window

		// The renderer of the hidden window, again just used to store assets.
		glfwRenderer

		// Signals shutdown to the assetLoader goroutine.
		exit chan error
	}
)

// assetLoader is the goroutine responsible for running the asset renderer.
func assetLoader() {
	renderExec := asset.glfwRenderer.RenderExec()

	// OpenGL function calls must occur in the same thread.
	runtime.LockOSThread()

	// Make the window's context the current one.
	asset.Window.MakeContextCurrent()

	for {
		select {
		case <-asset.exit:
			// Destroy renderer, while window and context are active.
			asset.glfwRenderer.Destroy()

			// Release context, before destroying window.
			glfw.DetachCurrentContext()

			// Destroy window and unlock the thread.
			err := asset.Window.Destroy()
			runtime.UnlockOSThread()

			// Signal completion.
			asset.exit <- err
			return

		case fn := <-renderExec:
			fn()
		}
	}
}

// doInit initializes GLFW and the hidden asset window/renderer, if not already
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

	// Create the asset renderer.
	asset.exit = make(chan error)
	asset.Window.MakeContextCurrent()
	asset.glfwRenderer, err = glfwNewRenderer()
	if err != nil {
		return err
	}
	glfw.DetachCurrentContext()

	go assetLoader()

	glfwInit = true
	return nil
}

// doExit de-initializes GLFW and the hidden asset/window renderer, only if it
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
