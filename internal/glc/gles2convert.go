// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build arm gles2

package glc

import (
	"azul3d.org/gfx.v2-dev"
	gl "azul3d.org/gfx.v2-dev/internal/gles2/2.0/gles2"
)

func (c *gles2Context) ConvertPrimitive(p gfx.Primitive) int {
	switch p {
	case gfx.Triangles:
		return gl.TRIANGLES
	case gfx.Points:
		return gl.POINTS
	case gfx.Lines:
		return gl.LINES
	default:
		panic("failed to convert")
	}
}

func (c *gles2Context) UnconvertFaceCull(fc int) gfx.FaceCullMode {
	switch fc {
	case gl.FRONT:
		return gfx.FrontFaceCulling
	case gl.BACK:
		return gfx.BackFaceCulling
	case gl.FRONT_AND_BACK:
		return gfx.NoFaceCulling
	default:
		panic("failed to convert")
	}
}

func (c *gles2Context) ConvertStencilOp(o gfx.StencilOp) int {
	switch o {
	case gfx.SKeep:
		return gl.KEEP
	case gfx.SZero:
		return gl.ZERO
	case gfx.SReplace:
		return gl.REPLACE
	case gfx.SIncr:
		return gl.INCR
	case gfx.SIncrWrap:
		return gl.INCR_WRAP
	case gfx.SDecr:
		return gl.DECR
	case gfx.SDecrWrap:
		return gl.DECR_WRAP
	case gfx.SInvert:
		return gl.INVERT
	default:
		panic("failed to convert")
	}
}

func (c *gles2Context) UnconvertStencilOp(o int) gfx.StencilOp {
	switch o {
	case gl.KEEP:
		return gfx.SKeep
	case gl.ZERO:
		return gfx.SZero
	case gl.REPLACE:
		return gfx.SReplace
	case gl.INCR:
		return gfx.SIncr
	case gl.INCR_WRAP:
		return gfx.SIncrWrap
	case gl.DECR:
		return gfx.SDecr
	case gl.DECR_WRAP:
		return gfx.SDecrWrap
	case gl.INVERT:
		return gfx.SInvert
	default:
		panic("failed to convert")
	}
}

func (c *gles2Context) ConvertCmp(cmp gfx.Cmp) int {
	switch cmp {
	case gfx.Always:
		return gl.ALWAYS
	case gfx.Never:
		return gl.NEVER
	case gfx.Less:
		return gl.LESS
	case gfx.LessOrEqual:
		return gl.LEQUAL
	case gfx.Greater:
		return gl.GREATER
	case gfx.GreaterOrEqual:
		return gl.GEQUAL
	case gfx.Equal:
		return gl.EQUAL
	case gfx.NotEqual:
		return gl.NOTEQUAL
	default:
		panic("failed to convert")
	}
}

func (c *gles2Context) UnconvertCmp(cmp int) gfx.Cmp {
	switch cmp {
	case gl.ALWAYS:
		return gfx.Always
	case gl.NEVER:
		return gfx.Never
	case gl.LESS:
		return gfx.Less
	case gl.LEQUAL:
		return gfx.LessOrEqual
	case gl.GREATER:
		return gfx.Greater
	case gl.GEQUAL:
		return gfx.GreaterOrEqual
	case gl.EQUAL:
		return gfx.Equal
	case gl.NOTEQUAL:
		return gfx.NotEqual
	default:
		panic("failed to convert")
	}
}

func (c *gles2Context) ConvertBlendOp(o gfx.BlendOp) int {
	switch o {
	case gfx.BZero:
		return gl.ZERO
	case gfx.BOne:
		return gl.ONE
	case gfx.BSrcColor:
		return gl.SRC_COLOR
	case gfx.BOneMinusSrcColor:
		return gl.ONE_MINUS_SRC_COLOR
	case gfx.BDstColor:
		return gl.DST_COLOR
	case gfx.BOneMinusDstColor:
		return gl.ONE_MINUS_DST_COLOR
	case gfx.BSrcAlpha:
		return gl.SRC_ALPHA
	case gfx.BOneMinusSrcAlpha:
		return gl.ONE_MINUS_SRC_ALPHA
	case gfx.BDstAlpha:
		return gl.DST_ALPHA
	case gfx.BOneMinusDstAlpha:
		return gl.ONE_MINUS_DST_ALPHA
	case gfx.BConstantColor:
		return gl.CONSTANT_COLOR
	case gfx.BOneMinusConstantColor:
		return gl.ONE_MINUS_CONSTANT_COLOR
	case gfx.BConstantAlpha:
		return gl.CONSTANT_ALPHA
	case gfx.BOneMinusConstantAlpha:
		return gl.ONE_MINUS_CONSTANT_ALPHA
	case gfx.BSrcAlphaSaturate:
		return gl.SRC_ALPHA_SATURATE
	default:
		panic("failed to convert")
	}
}

func (c *gles2Context) UnconvertBlendOp(o int) gfx.BlendOp {
	switch o {
	case gl.ZERO:
		return gfx.BZero
	case gl.ONE:
		return gfx.BOne
	case gl.SRC_COLOR:
		return gfx.BSrcColor
	case gl.ONE_MINUS_SRC_COLOR:
		return gfx.BOneMinusSrcColor
	case gl.DST_COLOR:
		return gfx.BDstColor
	case gl.ONE_MINUS_DST_COLOR:
		return gfx.BOneMinusDstColor
	case gl.SRC_ALPHA:
		return gfx.BSrcAlpha
	case gl.ONE_MINUS_SRC_ALPHA:
		return gfx.BOneMinusSrcAlpha
	case gl.DST_ALPHA:
		return gfx.BDstAlpha
	case gl.ONE_MINUS_DST_ALPHA:
		return gfx.BOneMinusDstAlpha
	case gl.CONSTANT_COLOR:
		return gfx.BConstantColor
	case gl.ONE_MINUS_CONSTANT_COLOR:
		return gfx.BOneMinusConstantColor
	case gl.CONSTANT_ALPHA:
		return gfx.BConstantAlpha
	case gl.ONE_MINUS_CONSTANT_ALPHA:
		return gfx.BOneMinusConstantAlpha
	case gl.SRC_ALPHA_SATURATE:
		return gfx.BSrcAlphaSaturate
	default:
		panic("failed to convert")
	}
}

func (c *gles2Context) ConvertBlendEq(eq gfx.BlendEq) int {
	switch eq {
	case gfx.BAdd:
		return gl.FUNC_ADD
	case gfx.BSub:
		return gl.FUNC_SUBTRACT
	case gfx.BReverseSub:
		return gl.FUNC_REVERSE_SUBTRACT
	default:
		panic("failed to convert")
	}
}

func (c *gles2Context) UnconvertBlendEq(eq int) gfx.BlendEq {
	switch eq {
	case gl.FUNC_ADD:
		return gfx.BAdd
	case gl.FUNC_SUBTRACT:
		return gfx.BSub
	case gl.FUNC_REVERSE_SUBTRACT:
		return gfx.BReverseSub
	default:
		panic("failed to convert")
	}
}
