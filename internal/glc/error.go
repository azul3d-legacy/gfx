// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glc

import (
	"errors"
	"fmt"
)

var (
	InvalidEnum                 = errors.New("GL_INVALID_ENUM")
	InvalidValue                = errors.New("GL_INVALID_VALUE")
	InvalidOperation            = errors.New("GL_INVALID_OPERATION")
	InvalidFramebufferOperation = errors.New("GL_INVALID_FRAMEBUFFER_OPERATION")
	OutOfMemory                 = errors.New("GL_OUT_OF_MEMORY")
	StackUnderflow              = errors.New("GL_STACK_UNDERFLOW")
	StackOverflow               = errors.New("GL_STACK_OVERFLOW")
)

func (c *Context) GetError() error {
	code := c.gl.GetError()
	switch code {
	case c.NO_ERROR:
		return nil
	case c.INVALID_ENUM:
		return InvalidEnum
	case c.INVALID_VALUE:
		return InvalidValue
	case c.INVALID_OPERATION:
		return InvalidOperation
	case c.INVALID_FRAMEBUFFER_OPERATION:
		return InvalidFramebufferOperation
	case c.OUT_OF_MEMORY:
		return OutOfMemory
	case c.STACK_UNDERFLOW:
		return StackUnderflow
	case c.STACK_OVERFLOW:
		return StackOverflow
	default:
		return fmt.Errorf("Unknown GL Error (0x%X)\n", code)
	}
}
