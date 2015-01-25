// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package camera

import (
	"math"

	"azul3d.org/gfx.v2-unstable"
)

var vertShader = []byte(`
#version 120

attribute vec3 Vertex;
attribute vec4 Color;
uniform mat4 MVP;

varying vec4 color;

void main(void) {
	gl_Position = MVP * vec4(Vertex, 1.0);
	color = Color;
}
`)

var fragShader = []byte(`
#version 120

varying vec4 color;

void main(void) {
	gl_FragColor = color;
}
`)

var shader *gfx.Shader

func init() {
	shader = gfx.NewShader("camera")
	shader.GLSL = &gfx.GLSLSources{
		Vertex:   vertShader,
		Fragment: fragShader,
	}
}

// DrawCamera draws the camera's viewing frustum to the canvas as a wire-frame.
// The first camera (cam) is what the wire-frame is attached to, while cam2 is
// the perspective from which to render said wire-frame.
func (c *Camera) debugUpdate() {
	c.State = gfx.NewState()
	c.Shader = shader
	c.State.FaceCulling = gfx.BackFaceCulling

	m := gfx.NewMesh()
	m.Primitive = gfx.Lines

	m.Vertices = []gfx.Vec3{}
	m.Colors = []gfx.Color{}

	near := float32(c.Near)
	far := float32(c.Far)

	if c.Ortho {
		width := float32(c.View.Dx())
		height := float32(c.View.Dy())

		m.Vertices = []gfx.Vec3{
			{width / 2, 0, height / 2},

			// Near
			{0, near, 0},
			{width, near, 0},
			{width, near, height},
			{0, near, height},

			// Far
			{0, far, 0},
			{width, far, 0},
			{width, far, height},
			{0, far, height},

			{width / 2, far, height / 2},

			// Up
			{0, near, height},
			{0, near, height},
			{width, near, height},
		}
	} else {
		ratio := float32(c.View.Dx()) / float32(c.View.Dy())
		fovRad := c.FOV / 180 * math.Pi

		hNear := float32(2 * math.Tan(fovRad/2) * c.Near)
		wNear := hNear * ratio

		hFar := float32(2 * math.Tan(fovRad/2) * c.Far)
		wFar := hFar * ratio

		m.Vertices = []gfx.Vec3{
			{0, 0, 0},

			// Near
			{-wNear / 2, near, -hNear / 2},
			{wNear / 2, near, -hNear / 2},
			{wNear / 2, near, hNear / 2},
			{-wNear / 2, near, hNear / 2},

			// Far
			{-wFar / 2, far, -hFar / 2},
			{wFar / 2, far, -hFar / 2},
			{wFar / 2, far, hFar / 2},
			{-wFar / 2, far, hFar / 2},

			{0, far, 0},

			// Up
			{0, near, hNear},
			{-wNear / 2 * 0.7, near, hNear / 2 * 1.1},
			{wNear / 2 * 0.7, near, hNear / 2 * 1.1},
		}
	}

	m.Colors = []gfx.Color{
		{1, 1, 1, 1},
		{1, 0.67, 0, 1},
		{1, 0.67, 0, 1},
		{1, 0.67, 0, 1},
		{1, 0.67, 0, 1},

		{1, 0.67, 0, 1},
		{1, 0.67, 0, 1},
		{1, 0.67, 0, 1},
		{1, 0.67, 0, 1},
		{1, 1, 1, 1},

		{0, 0.67, 1, 1},
		{0, 0.67, 1, 1},
		{0, 0.67, 1, 1},
	}

	m.Indices = []uint32{
		// From 0 to near plane
		0, 1,
		0, 2,
		0, 3,
		0, 4,

		// Near plane
		1, 2,
		2, 3,
		3, 4,
		4, 1,

		// Far plane
		5, 6,
		6, 7,
		7, 8,
		8, 5,

		// Lines from near to far plane
		1, 5,
		2, 6,
		3, 7,
		4, 8,

		0, 9,

		// Up
		10, 11,
		11, 12,
		12, 10,
	}

	c.Meshes = []*gfx.Mesh{m}
}
