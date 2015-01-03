// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"azul3d.org/gfx.v2-dev"
	"azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"
	"azul3d.org/gfx.v2-dev/internal/glc"
	"azul3d.org/gfx.v2-dev/internal/tag"
)

// Set this to true to disable state guarding (i.e. avoiding useless OpenGL
// state calls). This is useful for debugging the state guard code.
const noStateGuard = tag.Gsgdebug

type graphicsState struct {
	*glc.GraphicsState
	lastProgramPointSizeExt bool
}

func (g *graphicsState) Begin(d *device) {
	// Update viewport bounds.
	bounds := d.BaseCanvas.Bounds()
	gl.Viewport(0, 0, int32(bounds.Dx()), int32(bounds.Dy()))

	// Begin use of the graphics state.
	if !g.GraphicsState.Begin(bounds, g.beginCustom) {
		return
	}

	// Enable scissor testing.
	g.ScissorTest(true)

	// Enable setting point size in shader programs.
	g.programPointSizeExt(true)

	// Enable multisampling, if available and wanted.
	if d.glArbMultisample {
		if d.BaseCanvas.MSAA() {
			g.Multisample(true)
		}
	}
}

func (g *graphicsState) Restore(d *device) {
	bounds := d.BaseCanvas.Bounds()
	if !g.GraphicsState.Restore(bounds, g.restoreCustom) {
		return
	}
}

func (g *graphicsState) beginCustom() {
	// useProgram
	var sp int32
	gl.GetIntegerv(gl.CURRENT_PROGRAM, &sp)
	g.S.ShaderProgram = uint32(sp)

	// depthClamp
	gl.GetBooleanv(gl.DEPTH_CLAMP, &g.S.DepthClamp)

	// programPointSizeExt
	gl.GetBooleanv(gl.PROGRAM_POINT_SIZE_EXT, &g.lastProgramPointSizeExt)

	// stencilMaskSeparate
	g.getStencilMaskSeparate(&g.S.StencilFront, &g.S.StencilBack)

	// stencilFuncSeparate
	g.getStencilFuncSeparate(&g.S.StencilFront, &g.S.StencilBack)
}

func (g *graphicsState) restoreCustom() {
	g.useProgram(g.S.ShaderProgram)
	g.depthClamp(g.S.DepthClamp)
	g.programPointSizeExt(g.lastProgramPointSizeExt)
	g.stencilMaskSeparate(g.S.StencilFront.WriteMask, g.S.StencilBack.WriteMask)
	g.stencilFuncSeparate(g.S.StencilFront, g.S.StencilBack)
}

// Uncommon because WebGL needs a js.Object data type.
func (g *graphicsState) useProgram(p uint32) {
	if noStateGuard || g.S.ShaderProgram != p {
		g.S.ShaderProgram = p
		gl.UseProgram(p)
	}
}

// Specific to OpenGL 2.
//
// TODO(slimsag): See if WebGL or OpenGL ES 2 expose this through an extension.
func (g *graphicsState) depthClamp(v bool) {
	if noStateGuard || g.S.DepthClamp != v {
		g.C.Feature(gl.DEPTH_CLAMP, v)
	}
}

// Specific to OpenGL 2 (OpenGL ES 2 and WebGL 1.0 both have shader program
// point size enabled by default).
func (g *graphicsState) programPointSizeExt(v bool) {
	if noStateGuard || g.lastProgramPointSizeExt != v {
		g.lastProgramPointSizeExt = v
		g.C.Feature(gl.PROGRAM_POINT_SIZE_EXT, v)
	}
}

// Uncommon because WebGL doesn't support seperate stencil masks:
//
// https://www.khronos.org/registry/webgl/specs/latest/1.0/#6.10
//
// TODO(slimsag): See if WebGL exposes this through an extension.
func (g *graphicsState) stencilMaskSeparate(front, back uint) {
	if noStateGuard || g.S.StencilFront.WriteMask != front || g.S.StencilBack.WriteMask != back {
		g.S.StencilFront.WriteMask = front
		g.S.StencilBack.WriteMask = back

		// Save a call if front and back are identical.
		if front == back {
			gl.StencilMaskSeparate(gl.FRONT_AND_BACK, uint32(front))
			return
		}

		gl.StencilMaskSeparate(gl.FRONT, uint32(front))
		gl.StencilMaskSeparate(gl.BACK, uint32(back))
	}
}

func (g *graphicsState) getStencilMaskSeparate(front, back *gfx.StencilState) {
	var frontWriteMask, backWriteMask int32
	gl.GetIntegerv(gl.STENCIL_WRITEMASK, &frontWriteMask)
	gl.GetIntegerv(gl.STENCIL_BACK_WRITEMASK, &backWriteMask)
	front.WriteMask = uint(frontWriteMask)
	back.WriteMask = uint(backWriteMask)
}

// Uncommon because WebGL doesn't support seperate stencil refs:
//
// https://www.khronos.org/registry/webgl/specs/latest/1.0/#6.10
//
// TODO(slimsag): See if WebGL exposes this through an extension.
func (g *graphicsState) stencilFuncSeparate(front, back gfx.StencilState) {
	diff := func(a, b gfx.StencilState) bool {
		return a.Cmp != b.Cmp || a.Reference != b.Reference || a.ReadMask != b.ReadMask
	}

	if noStateGuard || diff(g.S.StencilFront, front) || diff(g.S.StencilBack, back) {
		g.S.StencilFront.Cmp = front.Cmp
		g.S.StencilFront.Reference = front.Reference
		g.S.StencilFront.ReadMask = front.ReadMask

		g.S.StencilBack.Cmp = back.Cmp
		g.S.StencilBack.Reference = back.Reference
		g.S.StencilBack.ReadMask = back.ReadMask

		// Save a call if front and back are identical.
		if front == back {
			gl.StencilFuncSeparate(
				gl.FRONT_AND_BACK,
				uint32(g.C.ConvertCmp(front.Cmp)),
				int32(front.Reference),
				uint32(front.ReadMask),
			)
			return
		}

		gl.StencilFuncSeparate(
			gl.FRONT,
			uint32(g.C.ConvertCmp(front.Cmp)),
			int32(front.Reference),
			uint32(front.ReadMask),
		)
		gl.StencilFuncSeparate(
			gl.BACK,
			uint32(g.C.ConvertCmp(back.Cmp)),
			int32(back.Reference),
			uint32(back.ReadMask),
		)
	}
}

func (g *graphicsState) getStencilFuncSeparate(front, back *gfx.StencilState) {
	// Front
	var frontFunc, frontRef, frontValueMask int32
	gl.GetIntegerv(gl.STENCIL_FUNC, &frontFunc)
	gl.GetIntegerv(gl.STENCIL_REF, &frontRef)
	gl.GetIntegerv(gl.STENCIL_VALUE_MASK, &frontValueMask)
	front.Cmp = g.C.UnconvertCmp(int(frontFunc))
	front.Reference = uint(frontRef)
	front.ReadMask = uint(frontValueMask)

	// Back
	var backFunc, backRef, backValueMask int32
	gl.GetIntegerv(gl.STENCIL_BACK_FUNC, &backFunc)
	gl.GetIntegerv(gl.STENCIL_BACK_REF, &backRef)
	gl.GetIntegerv(gl.STENCIL_BACK_VALUE_MASK, &backValueMask)
	back.Cmp = g.C.UnconvertCmp(int(backFunc))
	back.Reference = uint(backRef)
	back.ReadMask = uint(backValueMask)
}
