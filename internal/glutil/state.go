// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

import (
	"azul3d.org/gfx.v2-dev"
)

// Ensure these values match the default OpenGL state values listed in the
// OpenGL documentation!

var DefaultStencilState = gfx.StencilState{
	0xFFFF,     // WriteMask
	0xFFFF,     // ReadMask
	0,          // Reference
	gfx.SKeep,  // Fail
	gfx.SKeep,  // DepthFail
	gfx.SKeep,  // DepthPass
	gfx.Always, // Cmp
}

var DefaultBlendState = gfx.BlendState{
	gfx.Color{R: 0, G: 0, B: 0, A: 0}, // Color
	gfx.BOne,  // SrcRGB
	gfx.BZero, // DstRGB
	gfx.BOne,  // SrcAlpha
	gfx.BZero, // DstAlpha
	gfx.BAdd,  // RGBEq
	gfx.BAdd,  // AlphaEq
}

var DefaultState = &gfx.State{
	gfx.NoAlpha,         // AlphaMode
	DefaultBlendState,   // Blend
	true,                // WriteRed
	true,                // WriteGreen
	true,                // WriteBlue
	true,                // WriteAlpha
	true,                // Dithering
	false,               // DepthTest
	true,                // DepthWrite
	gfx.Less,            // DepthCmp
	false,               // StencilTest
	gfx.NoFaceCulling,   // FaceCulling
	DefaultStencilState, // FIXME: StencilFront
	DefaultStencilState, // FIXME: StencilBack
}