// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

// Primitive represents a single primitive type.
type Primitive uint8

const (
	// Triangles is a primitive type where each three mesh vertices forms a
	// single triangle.
	Triangles Primitive = iota

	// Lines is a primitive type where each two mesh vertices form a single
	// 1px wide line.
	//
	// Lines are restricted to 1px wide by the physical graphics hardware, if
	// you need wider lines you should form them from triangles.
	Lines

	// Points is a primitive type where each mesh vertex forms a single point
	// whose size is determined by the shader.
	Points
)
