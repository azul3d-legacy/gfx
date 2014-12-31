// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build js

package glc

import "github.com/gopherjs/webgl"

type Context struct {
	gl *webgl.Context
	*webgl.Context

	// TODO(slimsag): add to webgl bindings.
	ALWAYS int

	// WebGL doesn't have these errors, they are faked here for GetError.
	STACK_OVERFLOW int
	STACK_UNDERFLOW int
}

func NewContext(ctx *webgl.Context) *Context {
	return &Context{
		gl: ctx,
		Context: ctx,
		ALWAYS: 519,

		STACK_OVERFLOW: -1024,
		STACK_UNDERFLOW: -1025,
	}
}

