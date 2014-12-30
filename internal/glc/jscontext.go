// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build js

package glc

import gl "github.com/gopherjs/webgl"

func NewContext(c *gl.Context) *Context {
	return &Context{
		GetError: func() error { return getError(c) },
	}
}
