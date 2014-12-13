// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

// GLInfo holds information about the OpenGL implementation.
type GLInfo struct {
	// Major and minor versions of the OpenGL version in use. For example:
	//
	//  3, 0 (for OpenGL 3.0)
	//
	MajorVersion, MinorVersion int

	// A read-only slice of OpenGL extension strings.
	Extensions []string
}

// GLSLInfo holds information about the GLSL implementation.
type GLSLInfo struct {
	// Major and minor versions of the OpenGL Shading Language version that is
	// present. For example:
	//
	//  1, 30 (for GLSL 1.30)
	//
	MajorVersion, MinorVersion int

	// MaxVaryingFloats is the number of floating-point varying variables
	// available inside GLSL programs.
	//
	// Generally at least 32.
	MaxVaryingFloats int

	// MaxVertexInputs is the maximum number of vertex shader inputs (i.e.
	// floating-point values, where a 4x4 matrix is 16 floating-point values).
	//
	// Generally at least 512.
	MaxVertexInputs int

	// MaxFragmentInputs is the maximum number of fragment shader inputs (i.e.
	// floating-point values, where a 4x4 matrix is 16 floating-point values).
	//
	// Generally at least 64.
	MaxFragmentInputs int
}

// GLSLSources represents the sources to a GLSL shader program.
type GLSLSources struct {
	// The GLSL vertex shader source code.
	Vertex []byte

	// The GLSL fragment shader source code.
	Fragment []byte
}

// Copy returns a deep copy of this shader and it's source byte slices.
func (s *GLSLSources) Copy() *GLSLSources {
	cpy := &GLSLSources{
		Vertex:   make([]byte, len(s.Vertex)),
		Fragment: make([]byte, len(s.Fragment)),
	}
	copy(cpy.Vertex, s.Vertex)
	copy(cpy.Fragment, s.Fragment)
	return cpy
}
