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
// A shader will have it's inputs (from the gfx.Shader.Inputs map) mapped by
// name to GLSL uniforms. Types map directly to their GLSL counterparts (e.g.
// gfx.Vec4 -> GLSL "vec4" type), the only two notable ones are:
//
//  gfx.Color -> GLSL "vec4" type
//  gfx.TexCoord -> GLSL "vec2" type
//
package gl2
