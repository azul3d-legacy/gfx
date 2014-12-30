// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build js

package glc

import "github.com/gopherjs/webgl"

type jsContext struct {
	gl *webgl.Context
}

func NewContext(ctx *webgl.Context) Context {
	return &jsContext{
		gl: ctx,
	}
}
