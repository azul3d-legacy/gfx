// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"fmt"
	"strings"
	"unsafe"

	"azul3d.org/gfx.v2-unstable/internal/gl/2.0/gl"
	"azul3d.org/gfx.v2-unstable/internal/glutil"
)

func debugType(t uint32) string {
	switch t {
	case gl.DEBUG_TYPE_ERROR:
		return "ERROR"
	case gl.DEBUG_TYPE_DEPRECATED_BEHAVIOR:
		return "DEPRECATED_BEHAVIOR"
	case gl.DEBUG_TYPE_UNDEFINED_BEHAVIOR:
		return "UNDEFINED_BEHAVIOR"
	case gl.DEBUG_TYPE_PORTABILITY:
		return "PORTABILITY"
	case gl.DEBUG_TYPE_PERFORMANCE:
		return "PERFORMANCE"
	case gl.DEBUG_TYPE_OTHER:
		return "OTHER"
	default:
		return fmt.Sprintf("Type(0x%x)")
	}
}

func debugSeverity(t uint32) string {
	switch t {
	case gl.DEBUG_SEVERITY_LOW:
		return "LOW"
	case gl.DEBUG_SEVERITY_MEDIUM:
		return "MEDIUM"
	case gl.DEBUG_SEVERITY_HIGH:
		return "HIGH"
	default:
		return fmt.Sprintf("Severity(0x%x)")
	}
}

func glDebugCallback(
	source uint32,
	gltype uint32,
	id uint32,
	severity uint32,
	length int32,
	message string,
	userParam unsafe.Pointer) {

	// Log the error using the device.
	r := (*device)(userParam)
	r.warner.Warnf("OpenGL: %s\n", strings.TrimSpace(message))
	r.warner.Warnf("    Type: %s\n", debugType(gltype))
	r.warner.Warnf("    Severity: %s\n", debugSeverity(severity))
	r.warner.Warnf("    Source: %d\n", source)
	r.warner.Warnf("    ID: %d\n", id)
}

func (r *device) debugInit(exts glutil.Extensions) {
	// If we have the GL_ARB_debug_output extension we utilize it.
	r.glArbDebugOutput = exts.Present("GL_ARB_debug_output")
	if r.glArbDebugOutput {
		gl.Enable(gl.DEBUG_OUTPUT_SYNCHRONOUS_ARB)
		gl.DebugMessageCallbackARB(gl.DebugProc(glDebugCallback), unsafe.Pointer(r))
	}
}

func (r *device) debugRender() {
	// After each frame has been rendered we check for OpenGL errors.
	if err := r.common.GetError(); err != nil {
		r.warner.Warnf("OpenGL Error: %v\n", err)
	}
}
