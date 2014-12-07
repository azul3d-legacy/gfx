// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build 386,!gles2 amd64,!gles2

package window

import (
	"azul3d.org/gfx.v2-dev/gl2"
	glfw "azul3d.org/native/glfw.v3"
)

const (
	glfwClientAPI           = glfw.OpenGLAPI
	glfwContextVersionMajor = 2
	glfwContextVersionMinor = 0
)

var (
	keepState = gl2.KeepState
	share     = gl2.Share
)

func glfwNewRenderer(opts ...gl2.Option) (glfwRenderer, error) {
	return gl2.New(opts...)
}
