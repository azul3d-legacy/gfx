// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfxutil

import (
	"io/ioutil"
	"path/filepath"

	"azul3d.org/gfx.v2-dev"
)

// OpenShader opens the GLSL shader files specified by the given base path. For
// example:
//
//  s := OpenShader("glsl/basic")
//
// Would return a shader composed of the two GLSL shader sources:
//
//  glsl/basic.vert
//  glsl/basic.frag
//
// The filename (e.g. "basic") will be the name of the shader (which is used
// for debug output only).
//
// If a error is returned it is an IO error and a nil shader is also returned.
func OpenShader(basePath string) (*gfx.Shader, error) {
	// Load the GLSL vertex and fragment shader source files.
	vert, err := ioutil.ReadFile(basePath + ".vert")
	if err != nil {
		return nil, err
	}
	frag, err := ioutil.ReadFile(basePath + ".frag")
	if err != nil {
		return nil, err
	}

	// Create the new GLSL shader with the filename as the shader name.
	shader := gfx.NewShader(filepath.Base(basePath))
	shader.GLSL = &gfx.GLSLSources{
		Vertex:   vert,
		Fragment: frag,
	}
	return shader, nil
}
