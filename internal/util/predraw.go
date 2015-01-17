// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"errors"
	"image"

	"azul3d.org/gfx.v2-unstable"
)

var (
	ErrNilState    = errors.New("Draw: gfx.State is nil (ignoring object)")
	ErrNilShader   = errors.New("Draw: gfx.Shader is nil (ignoring object)")
	ErrNilSource   = errors.New("Draw: gfx.Texture has a nil Source image (ignoring object)")
	ErrNoVertices  = errors.New("Draw: gfx.Mesh has no vertices (ignoring object)")
	ErrNoMeshes    = errors.New("Draw: gfx.Object has no meshes (ignoring object)")
	ErrShaderError = errors.New("Draw: gfx.Shader has a compiler error (ignoring object)")
)

// PreDraw performs the commonplace tasks that occur before each object is
// drawn in a call to gfx.Canvas.Draw. It returns a draw boolean which signals
// whether or not the object should be drawn, and also may return a != nil
// error if the developer should be warned.
//
// It will return draw=false, err == nil in the following cases:
//
//  rect.Empty() == true
//  o.Shader != nil && len(o.Shader.Error) > 0
//
// It may return the following errors:
//
//  ErrNilState
//  ErrNilShader
//  ErrNilSource
//  ErrNoVertices
//  ErrNoMeshes
//  ErrShaderError
//
// If draw == true && err == nil, then it will:
//
// Make the implicit call to o.Bounds() required by gfx.Canvas such that the
// object has a chance to calculate a bounding box before it's data slices are
// set to nil.
//
// Ask the given device to load each shader, mesh, and texture that the object
// has associated with it and waits for loading to complete before returning.
func PreDraw(dev gfx.Device, rect image.Rectangle, o *gfx.Object, c *gfx.Camera) (draw bool, err error) {
	// Draw calls with empty rectangles are effectively no-op.
	if rect.Empty() {
		return false, nil
	}

	// Make the implicit o.Bounds() call required by gfx.Canvas.
	o.Bounds()

	// Test for basic cases of object invalidity.
	if o.State == nil {
		return false, ErrNilState
	}
	if o.Shader == nil {
		return false, ErrNilShader
	}
	if len(o.Shader.Error) > 0 {
		return false, ErrShaderError
	}
	if len(o.Meshes) == 0 {
		return false, ErrNoMeshes
	}
	meshesToLoad := 0
	for _, m := range o.Meshes {
		if m.Loaded && !m.HasChanged() {
			continue
		}
		meshesToLoad++
		if len(m.Vertices) == 0 {
			return false, ErrNoVertices
		}
	}
	texturesToLoad := 0
	for _, t := range o.Textures {
		if t.Loaded {
			continue
		}
		texturesToLoad++
		if t.Source == nil {
			return false, ErrNilSource
		}
	}

	// Begin loading of the object's resources now.
	var (
		shaderLoad  chan *gfx.Shader
		meshLoad    chan *gfx.Mesh
		textureLoad chan *gfx.Texture
	)
	if !o.Shader.Loaded {
		shaderLoad = make(chan *gfx.Shader, 1)
		dev.LoadShader(o.Shader, shaderLoad)
	}
	if meshesToLoad > 0 {
		meshLoad = make(chan *gfx.Mesh, meshesToLoad)
		for _, m := range o.Meshes {
			dev.LoadMesh(m, meshLoad)
		}
	}
	if texturesToLoad > 0 {
		textureLoad = make(chan *gfx.Texture, texturesToLoad)
		for _, t := range o.Textures {
			dev.LoadTexture(t, textureLoad)
		}
	}

	// Wait for loading of the object's resources to finish.
	for i := 0; i < cap(shaderLoad); i++ {
		<-shaderLoad
	}
	for i := 0; i < cap(meshLoad); i++ {
		<-meshLoad
	}
	for i := 0; i < cap(textureLoad); i++ {
		<-textureLoad
	}

	// Check the now-loaded shader for errors.
	if len(o.Shader.Error) > 0 {
		return false, ErrShaderError
	}
	return true, nil
}
