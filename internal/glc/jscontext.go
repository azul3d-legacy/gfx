// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build js

package glc

import (
	"azul3d.org/gfx.v2-dev"
	"github.com/gopherjs/webgl"
)

type glFuncs struct {
	*webgl.Context
	BlendColor        func(r, g, b, a float32)
	StencilOpSeparate func(face, fail, zfail, zpass int)

	GetScissorBox       func() (x, y, width, height int)
	GetColorWriteMask   func() (r, g, b, a bool)
	GetParameterColor   func(p int) gfx.Color
	GetParameterBool    func(p int) bool
	GetParameterInt     func(p int) int
	GetParameterFloat64 func(p int) float64
}

type Context struct {
	gl *glFuncs
	*webgl.Context

	// TODO(slimsag): add to webgl bindings.
	ALWAYS int

	// WebGL doesn't have these errors, they are faked here for GetError.
	STACK_OVERFLOW  int
	STACK_UNDERFLOW int

	MULTISAMPLE     int
	CLAMP_TO_BORDER int
}

func NewContext(ctx *webgl.Context) *Context {
	return &Context{
		gl: &glFuncs{
			Context: ctx,
			BlendColor: func(r, g, b, a float32) {
				ctx.BlendColor(float64(r), float64(g), float64(b), float64(a))
			},
			StencilOpSeparate: func(face, fail, zfail, zpass int) {
				// TODO(slimsag): add to webgl bindings.
				ctx.Call("stencilOpSeparate", face, fail, zfail, zpass)
			},
			GetScissorBox: func() (x, y, width, height int) {
				sb := ctx.GetParameter(ctx.SCISSOR_BOX).Interface().([]int)
				return sb[0], sb[1], sb[2], sb[3]
			},
			GetColorWriteMask: func() (r, g, b, a bool) {
				cwm := ctx.GetParameter(ctx.COLOR_WRITEMASK)
				r = cwm.Index(0).Bool()
				g = cwm.Index(1).Bool()
				b = cwm.Index(2).Bool()
				a = cwm.Index(3).Bool()
				return
			},
			GetParameterColor: func(p int) gfx.Color {
				f := ctx.GetParameter(p).Interface().([]float32)
				return gfx.Color{R: f[0], G: f[1], B: f[2], A: f[3]}
			},
			GetParameterBool: func(p int) bool {
				return ctx.GetParameter(p).Bool()
			},
			GetParameterInt: func(p int) int {
				return ctx.GetParameter(p).Int()
			},
			GetParameterFloat64: func(p int) float64 {
				return ctx.GetParameter(p).Float()
			},
		},
		Context: ctx,
		ALWAYS:  519,

		// TODO(slimsag): Find out if this is valid WebGL ? See gles2context.go
		MULTISAMPLE: 0x809D,

		STACK_OVERFLOW:  -1024,
		STACK_UNDERFLOW: -1025,

		// WebGL does not support BorderColor (CLAMP_TO_BORDER), per the gfx
		// package spec we choose just Clamp (CLAMP_TO_EDGE) instead.
		CLAMP_TO_BORDER: ctx.CLAMP_TO_EDGE,
	}
}
