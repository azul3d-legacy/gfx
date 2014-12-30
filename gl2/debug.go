// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
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
	r := (*device)(userParam)
	r.logf("OpenGL Debug (source=%d type=%d severity=%d):\n", source, gltype, severity)
	r.logf("    %s\n", message)
}

func (r *device) debugInit(exts map[string]bool) {
	// If we have the GL_ARB_debug_output extension we utilize it.
	r.glArbDebugOutput = extension("GL_ARB_debug_output", exts)
	if r.glArbDebugOutput {
		gl.Enable(gl.DEBUG_OUTPUT_SYNCHRONOUS_ARB)
		gl.DebugMessageCallbackARB(gl.DebugProc(glDebugCallback), unsafe.Pointer(r))
	}
}

func (r *device) debugRender() {
	// After each frame has been rendered we check for OpenGL errors.
	if err := r.Common.GetError(); err != nil {
		r.logf("OpenGL Error: %v\n", err)
	}
}
