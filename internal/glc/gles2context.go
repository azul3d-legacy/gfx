// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build arm gles2

package glc

import (
	"azul3d.org/gfx.v2-dev"
	gl "azul3d.org/gfx.v2-dev/internal/gles2/2.0/gles2"
)

type glFuncs struct {
	GetError              func() int
	Enable                func(capability int)
	Disable               func(capability int)
	Scissor               func(x, y, width, height int)
	ColorMask             func(r, g, b, a bool)
	ClearColor            func(r, g, b, a float32)
	ClearDepth            func(depth float64)
	ClearStencil          func(s int)
	DepthMask             func(b bool)
	DepthFunc             func(f int)
	CullFace              func(m int)
	BlendColor            func(r, g, b, a float32)
	BlendFuncSeparate     func(srcRGB, dstRGB, srcAlpha, dstAlpha int)
	BlendEquationSeparate func(modeRGB, modeAlpha int)
	StencilOpSeparate     func(face, fail, zfail, zpass int)

	GetScissorBox       func() (x, y, width, height int)
	GetColorWriteMask   func() (r, g, b, a bool)
	GetParameterColor   func(p int) gfx.Color
	GetParameterBool    func(p int) bool
	GetParameterInt     func(p int) int
	GetParameterFloat64 func(p int) float64
}

type Context struct {
	gl *glFuncs

	NO_ERROR                      int
	INVALID_ENUM                  int
	INVALID_VALUE                 int
	INVALID_OPERATION             int
	INVALID_FRAMEBUFFER_OPERATION int
	OUT_OF_MEMORY                 int
	STACK_UNDERFLOW               int
	STACK_OVERFLOW                int

	TRIANGLES      int
	POINTS         int
	LINES          int
	FRONT          int
	BACK           int
	FRONT_AND_BACK int

	KEEP      int
	ZERO      int
	REPLACE   int
	INCR      int
	INCR_WRAP int
	DECR      int
	DECR_WRAP int
	INVERT    int
	NEVER     int
	LESS      int
	LEQUAL    int
	ALWAYS    int
	GREATER   int
	GEQUAL    int
	EQUAL     int
	NOTEQUAL  int

	ONE                      int
	SRC_COLOR                int
	ONE_MINUS_SRC_COLOR      int
	DST_COLOR                int
	ONE_MINUS_DST_COLOR      int
	SRC_ALPHA                int
	ONE_MINUS_SRC_ALPHA      int
	DST_ALPHA                int
	ONE_MINUS_DST_ALPHA      int
	CONSTANT_COLOR           int
	ONE_MINUS_CONSTANT_COLOR int
	CONSTANT_ALPHA           int
	ONE_MINUS_CONSTANT_ALPHA int
	SRC_ALPHA_SATURATE       int

	FUNC_ADD              int
	FUNC_SUBTRACT         int
	FUNC_REVERSE_SUBTRACT int

	DITHER                   int
	SCISSOR_TEST             int
	STENCIL_TEST             int
	DEPTH_TEST               int
	CULL_FACE                int
	BLEND                    int
	SAMPLE_ALPHA_TO_COVERAGE int
	MULTISAMPLE              int

	DEPTH_WRITEMASK              int
	COLOR_CLEAR_VALUE            int
	BLEND_COLOR                  int
	DEPTH_CLEAR_VALUE            int
	STENCIL_CLEAR_VALUE          int
	DEPTH_FUNC                   int
	CULL_FACE_MODE               int
	BLEND_SRC_RGB                int
	BLEND_DST_RGB                int
	BLEND_SRC_ALPHA              int
	BLEND_DST_ALPHA              int
	BLEND_EQUATION_RGB           int
	BLEND_EQUATION_ALPHA         int
	STENCIL_FAIL                 int
	STENCIL_PASS_DEPTH_FAIL      int
	STENCIL_PASS_DEPTH_PASS      int
	STENCIL_BACK_FAIL            int
	STENCIL_BACK_PASS_DEPTH_FAIL int
	STENCIL_BACK_PASS_DEPTH_PASS int

	REPEAT                 int
	CLAMP_TO_EDGE          int
	CLAMP_TO_BORDER        int
	MIRRORED_REPEAT        int
	NEAREST                int
	LINEAR                 int
	NEAREST_MIPMAP_NEAREST int
	LINEAR_MIPMAP_NEAREST  int
	NEAREST_MIPMAP_LINEAR  int
	LINEAR_MIPMAP_LINEAR   int

	FRAMEBUFFER_COMPLETE                      int
	FRAMEBUFFER_INCOMPLETE_ATTACHMENT         int
	FRAMEBUFFER_INCOMPLETE_DIMENSIONS         int
	FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT int
	FRAMEBUFFER_INCOMPLETE_DRAW_BUFFER        int
	FRAMEBUFFER_INCOMPLETE_READ_BUFFER        int
	FRAMEBUFFER_INCOMPLETE_MULTISAMPLE        int
	FRAMEBUFFER_UNSUPPORTED                   int
	FRAMEBUFFER_UNDEFINED                     int
}

func NewContext() *Context {
	f := &glFuncs{
		GetError:     func() int { return int(gl.GetError()) },
		Enable:       func(capability int) { gl.Enable(uint32(capability)) },
		Disable:      func(capability int) { gl.Disable(uint32(capability)) },
		Scissor:      func(x, y, width, height int) { gl.Scissor(int32(x), int32(y), int32(width), int32(height)) },
		ColorMask:    gl.ColorMask,
		ClearColor:   gl.ClearColor,
		ClearDepth:   func(depth float64) { gl.ClearDepthf(float32(depth)) },
		ClearStencil: func(stencil int) { gl.ClearStencil(int32(stencil)) },
		DepthMask:    gl.DepthMask,
		DepthFunc:    func(f int) { gl.DepthFunc(uint32(f)) },
		CullFace:     func(m int) { gl.CullFace(uint32(m)) },
		BlendColor:   gl.BlendColor,
		BlendFuncSeparate: func(srcRGB, dstRGB, srcAlpha, dstAlpha int) {
			gl.BlendFuncSeparate(uint32(srcRGB), uint32(dstRGB), uint32(srcAlpha), uint32(dstAlpha))
		},
		BlendEquationSeparate: func(modeRGB, modeAlpha int) {
			gl.BlendEquationSeparate(uint32(modeRGB), uint32(modeAlpha))
		},
		StencilOpSeparate: func(face, fail, zfail, zpass int) {
			gl.StencilOpSeparate(uint32(face), uint32(fail), uint32(zfail), uint32(zpass))
		},
		GetScissorBox: func() (x, y, width, height int) {
			var s [4]int32
			gl.GetIntegerv(gl.SCISSOR_BOX, &s[0])
			return int(s[0]), int(s[1]), int(s[2]), int(s[3])
		},
		GetColorWriteMask: func() (r, g, b, a bool) {
			var s [4]bool
			gl.GetBooleanv(gl.COLOR_WRITEMASK, &s[0])
			return s[0], s[1], s[2], s[3]
		},
		GetParameterColor: func(p int) gfx.Color {
			var c gfx.Color
			gl.GetFloatv(uint32(p), &c.R)
			return c
		},
		GetParameterBool: func(p int) bool {
			var b bool
			gl.GetBooleanv(uint32(p), &b)
			return b
		},
		GetParameterInt: func(p int) int {
			var i int32
			gl.GetIntegerv(uint32(p), &i)
			return int(i)
		},
		GetParameterFloat64: func(p int) float64 {
			var f float32
			gl.GetFloatv(uint32(p), &f)
			return float64(f)
		},
	}

	return &Context{
		gl: f,

		NO_ERROR:                      gl.NO_ERROR,
		INVALID_ENUM:                  gl.INVALID_ENUM,
		INVALID_VALUE:                 gl.INVALID_VALUE,
		INVALID_OPERATION:             gl.INVALID_OPERATION,
		INVALID_FRAMEBUFFER_OPERATION: gl.INVALID_FRAMEBUFFER_OPERATION,
		OUT_OF_MEMORY:                 gl.OUT_OF_MEMORY,
		STACK_UNDERFLOW:               gl.STACK_UNDERFLOW,
		STACK_OVERFLOW:                gl.STACK_OVERFLOW,

		TRIANGLES:      gl.TRIANGLES,
		POINTS:         gl.POINTS,
		LINES:          gl.LINES,
		FRONT:          gl.FRONT,
		BACK:           gl.BACK,
		FRONT_AND_BACK: gl.FRONT_AND_BACK,

		KEEP:      gl.KEEP,
		ZERO:      gl.ZERO,
		REPLACE:   gl.REPLACE,
		INCR:      gl.INCR,
		INCR_WRAP: gl.INCR_WRAP,
		DECR:      gl.DECR,
		DECR_WRAP: gl.DECR_WRAP,
		INVERT:    gl.INVERT,
		NEVER:     gl.NEVER,
		LESS:      gl.LESS,
		LEQUAL:    gl.LEQUAL,
		ALWAYS:    gl.ALWAYS,
		GREATER:   gl.GREATER,
		GEQUAL:    gl.GEQUAL,
		EQUAL:     gl.EQUAL,
		NOTEQUAL:  gl.NOTEQUAL,

		ONE:                      gl.ONE,
		SRC_COLOR:                gl.SRC_COLOR,
		ONE_MINUS_SRC_COLOR:      gl.ONE_MINUS_SRC_COLOR,
		DST_COLOR:                gl.DST_COLOR,
		ONE_MINUS_DST_COLOR:      gl.ONE_MINUS_DST_COLOR,
		SRC_ALPHA:                gl.SRC_ALPHA,
		ONE_MINUS_SRC_ALPHA:      gl.ONE_MINUS_SRC_ALPHA,
		DST_ALPHA:                gl.DST_ALPHA,
		ONE_MINUS_DST_ALPHA:      gl.ONE_MINUS_DST_ALPHA,
		CONSTANT_COLOR:           gl.CONSTANT_COLOR,
		ONE_MINUS_CONSTANT_COLOR: gl.ONE_MINUS_CONSTANT_COLOR,
		CONSTANT_ALPHA:           gl.CONSTANT_ALPHA,
		ONE_MINUS_CONSTANT_ALPHA: gl.ONE_MINUS_CONSTANT_ALPHA,
		SRC_ALPHA_SATURATE:       gl.SRC_ALPHA_SATURATE,

		FUNC_ADD:              gl.FUNC_ADD,
		FUNC_SUBTRACT:         gl.FUNC_SUBTRACT,
		FUNC_REVERSE_SUBTRACT: gl.FUNC_REVERSE_SUBTRACT,

		DITHER:       gl.DITHER,
		SCISSOR_TEST: gl.SCISSOR_TEST,
		STENCIL_TEST: gl.STENCIL_TEST,
		DEPTH_TEST:   gl.DEPTH_TEST,
		CULL_FACE:    gl.CULL_FACE,
		BLEND:        gl.BLEND,
		SAMPLE_ALPHA_TO_COVERAGE: gl.SAMPLE_ALPHA_TO_COVERAGE,

		// TODO(slimsag): Find out if this is valid OpenGL ES 2.0 ?
		//
		// OpenGL ES 2.0 spec makes various references to the GL_MULTISAMPLE
		// constant, but it's not defined anywhere. Apparently multisampling
		// is enabled by default (i.e. multisample is always on if you request
		// a multisample EGL framebuffer).
		MULTISAMPLE: 0x809D,

		DEPTH_WRITEMASK:              gl.DEPTH_WRITEMASK,
		COLOR_CLEAR_VALUE:            gl.COLOR_CLEAR_VALUE,
		BLEND_COLOR:                  gl.BLEND_COLOR,
		DEPTH_CLEAR_VALUE:            gl.DEPTH_CLEAR_VALUE,
		STENCIL_CLEAR_VALUE:          gl.STENCIL_CLEAR_VALUE,
		DEPTH_FUNC:                   gl.DEPTH_FUNC,
		CULL_FACE_MODE:               gl.CULL_FACE_MODE,
		BLEND_SRC_RGB:                gl.BLEND_SRC_RGB,
		BLEND_DST_RGB:                gl.BLEND_DST_RGB,
		BLEND_SRC_ALPHA:              gl.BLEND_SRC_ALPHA,
		BLEND_DST_ALPHA:              gl.BLEND_DST_ALPHA,
		BLEND_EQUATION_RGB:           gl.BLEND_EQUATION_RGB,
		BLEND_EQUATION_ALPHA:         gl.BLEND_EQUATION_ALPHA,
		STENCIL_FAIL:                 gl.STENCIL_FAIL,
		STENCIL_PASS_DEPTH_FAIL:      gl.STENCIL_PASS_DEPTH_FAIL,
		STENCIL_PASS_DEPTH_PASS:      gl.STENCIL_PASS_DEPTH_PASS,
		STENCIL_BACK_FAIL:            gl.STENCIL_BACK_FAIL,
		STENCIL_BACK_PASS_DEPTH_FAIL: gl.STENCIL_BACK_PASS_DEPTH_FAIL,
		STENCIL_BACK_PASS_DEPTH_PASS: gl.STENCIL_BACK_PASS_DEPTH_PASS,

		REPEAT:                 gl.REPEAT,
		CLAMP_TO_EDGE:          gl.CLAMP_TO_EDGE,
		MIRRORED_REPEAT:        gl.MIRRORED_REPEAT,
		NEAREST:                gl.NEAREST,
		LINEAR:                 gl.LINEAR,
		NEAREST_MIPMAP_NEAREST: gl.NEAREST_MIPMAP_NEAREST,
		LINEAR_MIPMAP_NEAREST:  gl.LINEAR_MIPMAP_NEAREST,
		NEAREST_MIPMAP_LINEAR:  gl.NEAREST_MIPMAP_LINEAR,
		LINEAR_MIPMAP_LINEAR:   gl.LINEAR_MIPMAP_LINEAR,

		// OpenGL ES 2 does not support BorderColor (CLAMP_TO_BORDER), per the
		// gfx package spec we choose just Clamp (CLAMP_TO_EDGE) instead.
		CLAMP_TO_BORDER: gl.CLAMP_TO_EDGE,

		FRAMEBUFFER_COMPLETE:                      gl.FRAMEBUFFER_COMPLETE,
		FRAMEBUFFER_INCOMPLETE_ATTACHMENT:         gl.FRAMEBUFFER_INCOMPLETE_ATTACHMENT,
		FRAMEBUFFER_INCOMPLETE_DIMENSIONS:         gl.FRAMEBUFFER_INCOMPLETE_DIMENSIONS,
		FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT: gl.FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT,
		FRAMEBUFFER_INCOMPLETE_DRAW_BUFFER:        gl.FRAMEBUFFER_INCOMPLETE_DRAW_BUFFER,
		FRAMEBUFFER_INCOMPLETE_READ_BUFFER:        gl.FRAMEBUFFER_INCOMPLETE_READ_BUFFER,
		FRAMEBUFFER_INCOMPLETE_MULTISAMPLE:        gl.FRAMEBUFFER_INCOMPLETE_MULTISAMPLE,
		FRAMEBUFFER_UNSUPPORTED:                   gl.FRAMEBUFFER_UNSUPPORTED,
		FRAMEBUFFER_UNDEFINED:                     gl.FRAMEBUFFER_UNDEFINED,
	}
}
