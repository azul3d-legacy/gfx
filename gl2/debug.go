// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"errors"
	"unsafe"

	"azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"
)

func glDebugCallback(
	source uint32,
	gltype uint32,
	id uint32,
	severity uint32,
	length int32,
	message string,
	userParam unsafe.Pointer) {
	// TODO(slimsag): better printing of source, type, and severity.
	r := (*renderer)(userParam)
	r.logf("OpenGL Debug (source=%d type=%d severity=%d):\n", source, gltype, severity)
	r.logf("    %s\n", message)
}

func (r *renderer) debugInit(exts map[string]bool) {
	// If we have the GL_ARB_debug_output extension we use it. In all cases
	// we check glGetError after each frame has been rendered.
	r.glArbDebugOutput = extension("GL_ARB_debug_output", exts)
	if r.glArbDebugOutput {
		gl.Enable(gl.DEBUG_OUTPUT_SYNCHRONOUS_ARB)
		gl.DebugMessageCallbackARB(gl.DebugProc(glDebugCallback), unsafe.Pointer(r))
	}
}

func (r *renderer) debugRender() {
	var err error
	switch gl.GetError() {
	case gl.NO_ERROR:
		break
	case gl.INVALID_ENUM:
		err = errors.New("GL_INVALID_ENUM")
	case gl.INVALID_VALUE:
		err = errors.New("GL_INVALID_VALUE")
	case gl.INVALID_OPERATION:
		err = errors.New("GL_INVALID_OPERATION")
	case gl.INVALID_FRAMEBUFFER_OPERATION:
		err = errors.New("GL_INVALID_FRAMEBUFFER_OPERATION")
	case gl.OUT_OF_MEMORY:
		err = errors.New("GL_OUT_OF_MEMORY")
	case gl.STACK_UNDERFLOW:
		err = errors.New("GL_STACK_UNDERFLOW")
	case gl.STACK_OVERFLOW:
		err = errors.New("GL_STACK_OVERFLOW")
	}
	if err != nil {
		r.logf("OpenGL Error: %v\n", err)
	}
}
