// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"fmt"

	"azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"
	"azul3d.org/gfx.v2-dev/internal/glutil"
)

func getError() error {
	code := gl.GetError()
	switch code {
	case gl.NO_ERROR:
		return nil
	case gl.INVALID_ENUM:
		return glutil.InvalidEnum
	case gl.INVALID_VALUE:
		return glutil.InvalidValue
	case gl.INVALID_OPERATION:
		return glutil.InvalidOperation
	case gl.INVALID_FRAMEBUFFER_OPERATION:
		return glutil.InvalidFramebufferOperation
	case gl.OUT_OF_MEMORY:
		return glutil.OutOfMemory
	case gl.STACK_UNDERFLOW:
		return glutil.StackUnderflow
	case gl.STACK_OVERFLOW:
		return glutil.StackOverflow
	default:
		return fmt.Errorf("Unknown GL Error (0x%X)\n", code)
	}
}
