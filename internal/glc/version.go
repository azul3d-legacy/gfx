// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glc

import "azul3d.org/gfx.v2-dev/internal/glutil"

func (c *Context) Version() (major, minor, release int, vendor string) {
	s := c.gl.GetParameterString(c.VERSION)
	return glutil.ParseVersionString(s)
}

func (c *Context) ShadingLanguageVersion() (major, minor, release int, vendor string) {
	s := c.gl.GetParameterString(c.SHADING_LANGUAGE_VERSION)
	return glutil.ParseVersionString(s)
}
