// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glc

// Feature enables or disables the given capability depending on the given
// boolean.
func (c *Context) Feature(capability int, enabled bool) {
	if enabled {
		c.gl.Enable(capability)
		return
	}
	c.gl.Disable(capability)
}
