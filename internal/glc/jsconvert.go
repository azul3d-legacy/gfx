// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build js

package glc

import (
	"azul3d.org/gfx.v2-dev"
)

// TODO(slimsag): add to webgl bindings.
const ALWAYS = 519

func (c *jsContext) ConvertPrimitive(p gfx.Primitive) int {
	switch p {
	case gfx.Triangles:
		return c.gl.TRIANGLES
	case gfx.Points:
		return c.gl.POINTS
	case gfx.Lines:
		return c.gl.LINES
	default:
		panic("failed to convert")
	}
}

func (c *jsContext) UnconvertFaceCull(fc int) gfx.FaceCullMode {
	switch fc {
	case c.gl.FRONT:
		return gfx.FrontFaceCulling
	case c.gl.BACK:
		return gfx.BackFaceCulling
	case c.gl.FRONT_AND_BACK:
		return gfx.NoFaceCulling
	default:
		panic("failed to convert")
	}
}

func (c *jsContext) ConvertStencilOp(o gfx.StencilOp) int {
	switch o {
	case gfx.SKeep:
		return c.gl.KEEP
	case gfx.SZero:
		return c.gl.ZERO
	case gfx.SReplace:
		return c.gl.REPLACE
	case gfx.SIncr:
		return c.gl.INCR
	case gfx.SIncrWrap:
		return c.gl.INCR_WRAP
	case gfx.SDecr:
		return c.gl.DECR
	case gfx.SDecrWrap:
		return c.gl.DECR_WRAP
	case gfx.SInvert:
		return c.gl.INVERT
	default:
		panic("failed to convert")
	}
}

func (c *jsContext) UnconvertStencilOp(o int) gfx.StencilOp {
	switch o {
	case c.gl.KEEP:
		return gfx.SKeep
	case c.gl.ZERO:
		return gfx.SZero
	case c.gl.REPLACE:
		return gfx.SReplace
	case c.gl.INCR:
		return gfx.SIncr
	case c.gl.INCR_WRAP:
		return gfx.SIncrWrap
	case c.gl.DECR:
		return gfx.SDecr
	case c.gl.DECR_WRAP:
		return gfx.SDecrWrap
	case c.gl.INVERT:
		return gfx.SInvert
	default:
		panic("failed to convert")
	}
}

func (c *jsContext) ConvertCmp(cmp gfx.Cmp) int {
	switch cmp {
	case gfx.Always:
		return ALWAYS
	case gfx.Never:
		return c.gl.NEVER
	case gfx.Less:
		return c.gl.LESS
	case gfx.LessOrEqual:
		return c.gl.LEQUAL
	case gfx.Greater:
		return c.gl.GREATER
	case gfx.GreaterOrEqual:
		return c.gl.GEQUAL
	case gfx.Equal:
		return c.gl.EQUAL
	case gfx.NotEqual:
		return c.gl.NOTEQUAL
	default:
		panic("failed to convert")
	}
}

func (c *jsContext) UnconvertCmp(cmp int) gfx.Cmp {
	switch cmp {
	case ALWAYS:
		return gfx.Always
	case c.gl.NEVER:
		return gfx.Never
	case c.gl.LESS:
		return gfx.Less
	case c.gl.LEQUAL:
		return gfx.LessOrEqual
	case c.gl.GREATER:
		return gfx.Greater
	case c.gl.GEQUAL:
		return gfx.GreaterOrEqual
	case c.gl.EQUAL:
		return gfx.Equal
	case c.gl.NOTEQUAL:
		return gfx.NotEqual
	default:
		panic("failed to convert")
	}
}

func (c *jsContext) ConvertBlendOp(o gfx.BlendOp) int {
	switch o {
	case gfx.BZero:
		return c.gl.ZERO
	case gfx.BOne:
		return c.gl.ONE
	case gfx.BSrcColor:
		return c.gl.SRC_COLOR
	case gfx.BOneMinusSrcColor:
		return c.gl.ONE_MINUS_SRC_COLOR
	case gfx.BDstColor:
		return c.gl.DST_COLOR
	case gfx.BOneMinusDstColor:
		return c.gl.ONE_MINUS_DST_COLOR
	case gfx.BSrcAlpha:
		return c.gl.SRC_ALPHA
	case gfx.BOneMinusSrcAlpha:
		return c.gl.ONE_MINUS_SRC_ALPHA
	case gfx.BDstAlpha:
		return c.gl.DST_ALPHA
	case gfx.BOneMinusDstAlpha:
		return c.gl.ONE_MINUS_DST_ALPHA
	case gfx.BConstantColor:
		return c.gl.CONSTANT_COLOR
	case gfx.BOneMinusConstantColor:
		return c.gl.ONE_MINUS_CONSTANT_COLOR
	case gfx.BConstantAlpha:
		return c.gl.CONSTANT_ALPHA
	case gfx.BOneMinusConstantAlpha:
		return c.gl.ONE_MINUS_CONSTANT_ALPHA
	case gfx.BSrcAlphaSaturate:
		return c.gl.SRC_ALPHA_SATURATE
	default:
		panic("failed to convert")
	}
}

func (c *jsContext) UnconvertBlendOp(o int) gfx.BlendOp {
	switch o {
	case c.gl.ZERO:
		return gfx.BZero
	case c.gl.ONE:
		return gfx.BOne
	case c.gl.SRC_COLOR:
		return gfx.BSrcColor
	case c.gl.ONE_MINUS_SRC_COLOR:
		return gfx.BOneMinusSrcColor
	case c.gl.DST_COLOR:
		return gfx.BDstColor
	case c.gl.ONE_MINUS_DST_COLOR:
		return gfx.BOneMinusDstColor
	case c.gl.SRC_ALPHA:
		return gfx.BSrcAlpha
	case c.gl.ONE_MINUS_SRC_ALPHA:
		return gfx.BOneMinusSrcAlpha
	case c.gl.DST_ALPHA:
		return gfx.BDstAlpha
	case c.gl.ONE_MINUS_DST_ALPHA:
		return gfx.BOneMinusDstAlpha
	case c.gl.CONSTANT_COLOR:
		return gfx.BConstantColor
	case c.gl.ONE_MINUS_CONSTANT_COLOR:
		return gfx.BOneMinusConstantColor
	case c.gl.CONSTANT_ALPHA:
		return gfx.BConstantAlpha
	case c.gl.ONE_MINUS_CONSTANT_ALPHA:
		return gfx.BOneMinusConstantAlpha
	case c.gl.SRC_ALPHA_SATURATE:
		return gfx.BSrcAlphaSaturate
	default:
		panic("failed to convert")
	}
}

func (c *jsContext) ConvertBlendEq(eq gfx.BlendEq) int {
	switch eq {
	case gfx.BAdd:
		return c.gl.FUNC_ADD
	case gfx.BSub:
		return c.gl.FUNC_SUBTRACT
	case gfx.BReverseSub:
		return c.gl.FUNC_REVERSE_SUBTRACT
	default:
		panic("failed to convert")
	}
}

func (c *jsContext) UnconvertBlendEq(eq int) gfx.BlendEq {
	switch eq {
	case c.gl.FUNC_ADD:
		return gfx.BAdd
	case c.gl.FUNC_SUBTRACT:
		return gfx.BSub
	case c.gl.FUNC_REVERSE_SUBTRACT:
		return gfx.BReverseSub
	default:
		panic("failed to convert")
	}
}
