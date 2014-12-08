// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build 386 amd64

package window

import (
	"runtime"

	glfw "azul3d.org/native/glfw.v3"
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
	}
)

// assetLoader is the goroutine responsible for running the asset renderer.
//
// TODO(slimsag): it should exit when no more windows are open
func assetLoader() {
	renderExec := asset.glfwRenderer.RenderExec()

	// OpenGL function calls must occur in the same thread.
	runtime.LockOSThread()

	// Make the window's context the current one.
	asset.Window.MakeContextCurrent()

	for {
		select {
		case fn := <-renderExec:
			fn()
		}
	}
}

// doInit initializes GLFW and the hidden asset window/renderer, if not already
// initialized.
func doInit() error {
	// Initialize GLFW, if needed.
	if !glfwInit {
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

		// Create the asset renderer.
		asset.Window.MakeContextCurrent()
		asset.glfwRenderer, err = glfwNewRenderer()
		if err != nil {
			return err
		}
		glfw.DetachCurrentContext()

		go assetLoader()

		glfwInit = true
	}
	return nil
}
