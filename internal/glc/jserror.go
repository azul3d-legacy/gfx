// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build js

package glc

import "fmt"

func (c *jsContext) GetError() error {
	code := c.gl.GetError()
	switch code {
	case c.gl.NO_ERROR:
		return nil
	case c.gl.INVALID_ENUM:
		return InvalidEnum
	case c.gl.INVALID_VALUE:
		return InvalidValue
	case c.gl.INVALID_OPERATION:
		return InvalidOperation
	case c.gl.INVALID_FRAMEBUFFER_OPERATION:
		return InvalidFramebufferOperation
	case c.gl.OUT_OF_MEMORY:
		return OutOfMemory
	default:
		return fmt.Errorf("Unknown GL Error (0x%X)\n", code)
	}
}
