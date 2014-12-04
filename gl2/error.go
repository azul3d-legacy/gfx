// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"errors"
	"fmt"

	"azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"
)

func getError() error {
	var err error
	switch gl.GetError() {
	case gl.NO_ERROR:
		return nil
	case gl.INVALID_ENUM:
		return errors.New("GL_INVALID_ENUM")
	case gl.INVALID_VALUE:
		return errors.New("GL_INVALID_VALUE")
	case gl.INVALID_OPERATION:
		return errors.New("GL_INVALID_OPERATION")
	case gl.INVALID_FRAMEBUFFER_OPERATION:
		return errors.New("GL_INVALID_FRAMEBUFFER_OPERATION")
	case gl.OUT_OF_MEMORY:
		return errors.New("GL_OUT_OF_MEMORY")
	case gl.STACK_UNDERFLOW:
		return errors.New("GL_STACK_UNDERFLOW")
	case gl.STACK_OVERFLOW:
		return errors.New("GL_STACK_OVERFLOW")
	default:
		return errors.New(fmt.Sprintf("Unknown Error: %v\n", err))
	}
}
