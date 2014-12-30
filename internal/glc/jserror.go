// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build js

package glc

import (
	"fmt"

	"azul3d.org/gfx.v2-dev/internal/glc"
	gl "github.com/gopherjs/webgl"
)

func getError(c *gl.Context) error {
	code := c.GetError()
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
	default:
		return fmt.Errorf("Unknown GL Error (0x%X)\n", code)
	}
}
