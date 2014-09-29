// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gl2 provides an OpenGL 2 based graphics renderer.
//
// The behavior of the renderer is defined fully in the gfx package (as such
// this package only makes mention of strictly OpenGL related caveats like
// initialization, etc).
//
// When performing render-to-texture (RTT), feedback loops are explicitly
// prohibited. This means that the renderer will panic if you attempt to draw
// an object to an RTT canvas when the object uses the literal RTT texture in
// itself. Through OpenGL the result of this is at best a corrupt image -- and
// at worst driver-level memory corruption (hence it is not allowed).
//
// A texture can turn on and off mipmapped by setting it's minification filter
// to a mipmapped or non-mipmapped filter after the texture has been loaded,
// but mipmapped can only be turned on with a loaded texture if when it loaded
// it had a mipmapped minification filter set on it.
//
// The default GLSL shader inputs are as follows:
//
//  attribute vec3 Vertex;      -> from gfx.Mesh.Vertices and gfx.Mesh.Indices
//  attribute vec4 Color;       -> from gfx.Mesh.Colors
//  attribute vec3 Bary;        -> from gfx.Mesh.Bary
//  attribute vec2 TexCoord[N]; -> [N] is the nth index of gfx.Mesh.TexCoords
//
//  uniform mat4 Model;       -> Model matrix from gfx.Object.Transform
//  uniform mat4 View;        -> View matrix from gfx.Camera.Transform
//  uniform mat4 Projection;  -> Projection matrix from gfx.Camera.Projection
//  uniform mat4 MVP;         -> Premultiplied Model/View/Projection matrix.
//  uniform bool BinaryAlpha; -> See below.
//
// BinaryAlpha is a boolean uniform value that informs the shader of the chosen
// alpha transparency mode of an object. It is set to true if the gfx.Object
// being drawn has a gfx.State.AlphaMode of gfx.BinaryAlpha or if the alpha
// mode is gfx.AlphaToCoverage but the GPU does not support it.
package gl2
