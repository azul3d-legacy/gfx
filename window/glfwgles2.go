// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build 386,gles2 amd64,gles2

package window

import (
	"azul3d.org/gfx.v2-unstable/gles2"
	glfw "azul3d.org/native/glfw.v4"
)

const (
	glfwClientAPI           = glfw.OpenGLESAPI
	glfwContextVersionMajor = 2
	glfwContextVersionMinor = 0
)

var share = gles2.Share

func glfwNewRenderer(opts ...gles2.Option) (glfwRenderer, error) {
	return gles2.New(opts...)
}
