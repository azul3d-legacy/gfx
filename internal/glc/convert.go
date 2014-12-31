// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glc

import "azul3d.org/gfx.v2-dev"

func (c *Context) ConvertPrimitive(p gfx.Primitive) int {
	switch p {
	case gfx.Triangles:
		return c.TRIANGLES
	case gfx.Points:
		return c.POINTS
	case gfx.Lines:
		return c.LINES
	default:
		panic("failed to convert")
	}
}

func (c *Context) UnconvertFaceCull(fc int) gfx.FaceCullMode {
	switch fc {
	case c.FRONT:
		return gfx.FrontFaceCulling
	case c.BACK:
		return gfx.BackFaceCulling
	case c.FRONT_AND_BACK:
		return gfx.NoFaceCulling
	default:
		panic("failed to convert")
	}
}

func (c *Context) ConvertStencilOp(o gfx.StencilOp) int {
	switch o {
	case gfx.SKeep:
		return c.KEEP
	case gfx.SZero:
		return c.ZERO
	case gfx.SReplace:
		return c.REPLACE
	case gfx.SIncr:
		return c.INCR
	case gfx.SIncrWrap:
		return c.INCR_WRAP
	case gfx.SDecr:
		return c.DECR
	case gfx.SDecrWrap:
		return c.DECR_WRAP
	case gfx.SInvert:
		return c.INVERT
	default:
		panic("failed to convert")
	}
}

func (c *Context) UnconvertStencilOp(o int) gfx.StencilOp {
	switch o {
	case c.KEEP:
		return gfx.SKeep
	case c.ZERO:
		return gfx.SZero
	case c.REPLACE:
		return gfx.SReplace
	case c.INCR:
		return gfx.SIncr
	case c.INCR_WRAP:
		return gfx.SIncrWrap
	case c.DECR:
		return gfx.SDecr
	case c.DECR_WRAP:
		return gfx.SDecrWrap
	case c.INVERT:
		return gfx.SInvert
	default:
		panic("failed to convert")
	}
}

func (c *Context) ConvertCmp(cmp gfx.Cmp) int {
	switch cmp {
	case gfx.Always:
		return c.ALWAYS
	case gfx.Never:
		return c.NEVER
	case gfx.Less:
		return c.LESS
	case gfx.LessOrEqual:
		return c.LEQUAL
	case gfx.Greater:
		return c.GREATER
	case gfx.GreaterOrEqual:
		return c.GEQUAL
	case gfx.Equal:
		return c.EQUAL
	case gfx.NotEqual:
		return c.NOTEQUAL
	default:
		panic("failed to convert")
	}
}

func (c *Context) UnconvertCmp(cmp int) gfx.Cmp {
	switch cmp {
	case c.ALWAYS:
		return gfx.Always
	case c.NEVER:
		return gfx.Never
	case c.LESS:
		return gfx.Less
	case c.LEQUAL:
		return gfx.LessOrEqual
	case c.GREATER:
		return gfx.Greater
	case c.GEQUAL:
		return gfx.GreaterOrEqual
	case c.EQUAL:
		return gfx.Equal
	case c.NOTEQUAL:
		return gfx.NotEqual
	default:
		panic("failed to convert")
	}
}

func (c *Context) ConvertBlendOp(o gfx.BlendOp) int {
	switch o {
	case gfx.BZero:
		return c.ZERO
	case gfx.BOne:
		return c.ONE
	case gfx.BSrcColor:
		return c.SRC_COLOR
	case gfx.BOneMinusSrcColor:
		return c.ONE_MINUS_SRC_COLOR
	case gfx.BDstColor:
		return c.DST_COLOR
	case gfx.BOneMinusDstColor:
		return c.ONE_MINUS_DST_COLOR
	case gfx.BSrcAlpha:
		return c.SRC_ALPHA
	case gfx.BOneMinusSrcAlpha:
		return c.ONE_MINUS_SRC_ALPHA
	case gfx.BDstAlpha:
		return c.DST_ALPHA
	case gfx.BOneMinusDstAlpha:
		return c.ONE_MINUS_DST_ALPHA
	case gfx.BConstantColor:
		return c.CONSTANT_COLOR
	case gfx.BOneMinusConstantColor:
		return c.ONE_MINUS_CONSTANT_COLOR
	case gfx.BConstantAlpha:
		return c.CONSTANT_ALPHA
	case gfx.BOneMinusConstantAlpha:
		return c.ONE_MINUS_CONSTANT_ALPHA
	case gfx.BSrcAlphaSaturate:
		return c.SRC_ALPHA_SATURATE
	default:
		panic("failed to convert")
	}
}

func (c *Context) UnconvertBlendOp(o int) gfx.BlendOp {
	switch o {
	case c.ZERO:
		return gfx.BZero
	case c.ONE:
		return gfx.BOne
	case c.SRC_COLOR:
		return gfx.BSrcColor
	case c.ONE_MINUS_SRC_COLOR:
		return gfx.BOneMinusSrcColor
	case c.DST_COLOR:
		return gfx.BDstColor
	case c.ONE_MINUS_DST_COLOR:
		return gfx.BOneMinusDstColor
	case c.SRC_ALPHA:
		return gfx.BSrcAlpha
	case c.ONE_MINUS_SRC_ALPHA:
		return gfx.BOneMinusSrcAlpha
	case c.DST_ALPHA:
		return gfx.BDstAlpha
	case c.ONE_MINUS_DST_ALPHA:
		return gfx.BOneMinusDstAlpha
	case c.CONSTANT_COLOR:
		return gfx.BConstantColor
	case c.ONE_MINUS_CONSTANT_COLOR:
		return gfx.BOneMinusConstantColor
	case c.CONSTANT_ALPHA:
		return gfx.BConstantAlpha
	case c.ONE_MINUS_CONSTANT_ALPHA:
		return gfx.BOneMinusConstantAlpha
	case c.SRC_ALPHA_SATURATE:
		return gfx.BSrcAlphaSaturate
	default:
		panic("failed to convert")
	}
}

func (c *Context) ConvertBlendEq(eq gfx.BlendEq) int {
	switch eq {
	case gfx.BAdd:
		return c.FUNC_ADD
	case gfx.BSub:
		return c.FUNC_SUBTRACT
	case gfx.BReverseSub:
		return c.FUNC_REVERSE_SUBTRACT
	default:
		panic("failed to convert")
	}
}

func (c *Context) UnconvertBlendEq(eq int) gfx.BlendEq {
	switch eq {
	case c.FUNC_ADD:
		return gfx.BAdd
	case c.FUNC_SUBTRACT:
		return gfx.BSub
	case c.FUNC_REVERSE_SUBTRACT:
		return gfx.BReverseSub
	default:
		panic("failed to convert")
	}
}
