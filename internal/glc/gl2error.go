// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build 386,!gles2 amd64,!gles2

package glc

import (
	"fmt"

	"azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"
)

func (c *gl2Context) GetError() error {
	code := gl.GetError()
	switch code {
	case gl.NO_ERROR:
		return nil
	case gl.INVALID_ENUM:
		return InvalidEnum
	case gl.INVALID_VALUE:
		return InvalidValue
	case gl.INVALID_OPERATION:
		return InvalidOperation
	case gl.INVALID_FRAMEBUFFER_OPERATION:
		return InvalidFramebufferOperation
	case gl.OUT_OF_MEMORY:
		return OutOfMemory
	case gl.STACK_UNDERFLOW:
		return StackUnderflow
	case gl.STACK_OVERFLOW:
		return StackOverflow
	default:
		return fmt.Errorf("Unknown GL Error (0x%X)\n", code)
	}
}
