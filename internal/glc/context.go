// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glc

import "azul3d.org/gfx.v2-dev"

type Context interface {
	GetError() error

	ConvertPrimitive(p gfx.Primitive) int
	UnconvertFaceCull(facecull int) gfx.FaceCullMode

	ConvertStencilOp(o gfx.StencilOp) int
	UnconvertStencilOp(o int) gfx.StencilOp

	ConvertCmp(cmp gfx.Cmp) int
	UnconvertCmp(cmp int) gfx.Cmp

	ConvertBlendOp(o gfx.BlendOp) int
	UnconvertBlendOp(o int) gfx.BlendOp

	ConvertBlendEq(eq gfx.BlendEq) int
	UnconvertBlendEq(eq int) gfx.BlendEq
}
