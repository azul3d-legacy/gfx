// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import "fmt"

// GLInfo holds information about the OpenGL implementation.
type GLInfo struct {
	// Major, minor, and release versions of the OpenGL version in use. For
	// example:
	//
	//  // OpenGL v2.1.3
	//  MajorVersion: 2
	//  MinorVersion: 1
	//  ReleaseVersion: 3
	//
	MajorVersion, MinorVersion, ReleaseVersion int

	// VendorVersion is a vendor/manufacturer specific version string. It can
	// be anything at all (and sometimes nothing), but is typically the driver
	// version in use. For example:
	//
	//  "Mesa 10.5.0-devel (git-b3721cd 2014-11-26 trusty-oibaf-ppa)"
	//  ""
	//
	VendorVersion string

	// A read-only slice of OpenGL extension strings.
	Extensions []string
}

// String returns a OpenGL version string like so:
//
//  "OpenGL v2.1 - Mesa 10.5.0-devel (git-b3721cd 2014-11-26 trusty-oibaf-ppa)"
//
func (g *GLInfo) String() string {
	return fmt.Sprintf("OpenGL %s - %s", g.Version(), g.VendorVersion)
}

// Version returns a version string like so:
//
//  "v1"
//  "v2.1"
//  "v2.1.1"
//
func (g *GLInfo) Version() string {
	if g.MinorVersion == 0 && g.ReleaseVersion == 0 {
		return fmt.Sprintf("v%d", g.MajorVersion)
	}
	if g.ReleaseVersion == 0 {
		return fmt.Sprintf("v%d.%d", g.MajorVersion, g.MinorVersion)
	}
	return fmt.Sprintf("v%d.%d.%d")
}

// GLSLInfo holds information about the GLSL implementation.
type GLSLInfo struct {
	// Major, minor, and release versions of the OpenGL Shading Language
	// version present. For example:
	//
	//  // GLSL v1.30.2
	//  MajorVersion: 1
	//  MinorVersion: 30
	//  ReleaseVersion: 2
	//
	MajorVersion, MinorVersion, ReleaseVersion int

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

// String returns a GLSL version string like so:
//
//  "GLSL v1"
//
func (g *GLSLInfo) String() string {
	return fmt.Sprintf("GLSL %s", g.Version())
}

// Version returns a version string like so:
//
//  "v1"
//  "v2.1"
//  "v2.1.1"
//
func (g *GLSLInfo) Version() string {
	if g.MinorVersion == 0 && g.ReleaseVersion == 0 {
		return fmt.Sprintf("v%d", g.MajorVersion)
	}
	if g.ReleaseVersion == 0 {
		return fmt.Sprintf("v%d.%d", g.MajorVersion, g.MinorVersion)
	}
	return fmt.Sprintf("v%d.%d.%d")
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
