// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build 386,!gles2 amd64,!gles2

package window

import (
	"azul3d.org/gfx.v2-unstable/gl2"
	"azul3d.org/native/glfw.v5"
)

const (
	glfwClientAPI           = glfw.OpenGLAPI
	glfwContextVersionMajor = 2
	glfwContextVersionMinor = 0
)

var share = gl2.Share

func glfwNewDevice(opts ...gl2.Option) (glfwDevice, error) {
	return gl2.New(opts...)
}
