// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build 386,!gles2 amd64,!gles2

package glc

import "azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"

type glFuncs struct {
	GetError func() int
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
}

func NewContext() *Context {
	f := &glFuncs{
		GetError: func() int { return int(gl.GetError()) },
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
	}
}