// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glc

import (
	"image"

	"azul3d.org/gfx.v2-dev"
	"azul3d.org/gfx.v2-dev/internal/glutil"
	"azul3d.org/gfx.v2-dev/internal/tag"
)

// Background
//
// OpenGL uses a state-binding API, where there is a single graphics state. You
// bind a state through the OpenGL API and later unbind it when you no longer
// want that state. For example:
//
//  glEnable(gl.DITHER)
//  drawTriangle()
//  drawCube()
//  drawMonkey()
//  glDisable(gl.DITHER)
//
// All three of the objects (triangle, cube, and monkey) are drawn with
// dithering enabled. It is NOT efficient to constantly bind/unbind state like
// so:
//
//  glEnable(gl.DITHER)
//  drawTriangle()
//  glDisable(gl.DITHER)
//
//  glEnable(gl.DITHER)
//  drawCube()
//  glDisable(gl.DITHER)
//
//  glEnable(gl.DITHER)
//  drawMonkey()
//  glDisable(gl.DITHER)
//
// The gfx package, however, allows each object to hold it's own state. This is
// very useful as it allows a graphical object to solely define how it will be
// drawn. A naive translation of this model to OpenGL would be something like
// the innefficient code above.
//
// To translate gfx states to OpenGL, instead, we divert a series of checks to
// see if that state is _already_ bound. This looks something like this:
//
//  var currentDithering bool
//  func dithering(newDithering bool) {
//      if currentDithering == newDithering {
//          // Dithering is already in this state.
//          return
//      }
//
//      if newDithering {
//          glEnable(gl.DITHER)
//          return
//      }
//      glDisable(gl.DITHER)
//  }
//
//  dithering(true)
//  drawTriangle()
//
//  dithering(true)
//  drawCube()
//
//  dithering(true)
//  drawMonkey()
//
// Not only does this yield similiar performance to the efficient code, but it
// also allows us to carry the OpenGL state much further than you could when
// writing pure OpenGL code directly.
//
// This is because we can even carry the OpenGL state across multiple frames,
// and also because we no longer suffer from the CGO overhead of unbinding as
// often.

// Set this to true to disable state guarding (i.e. avoiding useless OpenGL
// state calls). This is useful for debugging the state guard code.
const noStateGuard = tag.Gsgdebug

type GraphicsState struct {
	C     *Context
	S     *glutil.CommonState
	Saved *glutil.CommonState
}

// Begin begins use of this graphics state by saving the existing OpenGL state
// for restoration later (see Restore).
//
// If Begin was already called without a call to Restore later, the call to
// this function is no-op, and false is returned.
func (g *GraphicsState) Begin(bounds image.Rectangle, custom func()) bool {
	if g.Saved != nil {
		return false
	}

	g.S.Scissor = g.getScissor(bounds)
	g.S.WriteRed, g.S.WriteGreen, g.S.WriteBlue, g.S.WriteAlpha = g.C.gl.GetColorWriteMask()
	g.S.ClearColor = g.C.gl.GetParameterColor(g.C.COLOR_CLEAR_VALUE)
	g.S.DepthWrite = g.C.gl.GetParameterBool(g.C.DEPTH_WRITEMASK)
	g.S.ClearDepth = g.C.gl.GetParameterFloat64(g.C.DEPTH_CLEAR_VALUE)
	g.S.State.Blend.Color = g.C.gl.GetParameterColor(g.C.BLEND_COLOR)
	g.S.ClearStencil = g.C.gl.GetParameterInt(g.C.STENCIL_CLEAR_VALUE)
	g.S.DepthCmp = g.C.UnconvertCmp(g.C.gl.GetParameterInt(g.C.DEPTH_FUNC))
	g.S.FaceCulling = g.getFaceCulling()
	g.getBlendFuncSeparate(&g.S.State.Blend)
	g.getBlendEquationSeparate(&g.S.State.Blend)
	g.getStencilOpSeparate(&g.S.StencilFront, &g.S.StencilBack)
	g.S.Dithering = g.C.gl.GetParameterBool(g.C.DITHER)
	g.S.ScissorTest = g.C.gl.GetParameterBool(g.C.SCISSOR_TEST)
	g.S.StencilTest = g.C.gl.GetParameterBool(g.C.STENCIL_TEST)
	g.S.DepthTest = g.C.gl.GetParameterBool(g.C.DEPTH_TEST)
	g.S.Blend = g.C.gl.GetParameterBool(g.C.BLEND)
	g.S.SampleAlphaToCoverage = g.C.gl.GetParameterBool(g.C.SAMPLE_ALPHA_TO_COVERAGE)
	g.S.Multisample = g.C.gl.GetParameterBool(g.C.MULTISAMPLE)

	custom()

	// Save the OpenGL state now.
	stateCpy := *g.S
	g.Saved = &stateCpy
	return true
}

// Restore restores the OpenGL graphics state. If Restore was called without a
// preceding call to Begin, then the call to this function is no-op  and false
// is returned.
func (g *GraphicsState) Restore(bounds image.Rectangle, custom func()) bool {
	if g.Saved == nil {
		return false
	}
	sv := g.Saved
	g.Saved = nil

	g.Scissor(bounds, sv.Scissor)
	g.ColorWrite(sv.WriteRed, sv.WriteGreen, sv.WriteBlue, sv.WriteAlpha)
	g.ClearColor(sv.ClearColor)
	g.DepthWrite(sv.DepthWrite)
	g.ClearDepth(sv.ClearDepth)
	g.BlendColor(sv.State.Blend.Color)
	g.ClearStencil(sv.ClearStencil)
	g.DepthCmp(sv.DepthCmp)
	g.FaceCulling(sv.FaceCulling)
	g.BlendFuncSeparate(sv.State.Blend)
	g.BlendEquationSeparate(sv.State.Blend)
	g.StencilOpSeparate(sv.StencilFront, sv.StencilBack)
	g.Dithering(sv.Dithering)
	g.ScissorTest(sv.ScissorTest)
	g.StencilTest(sv.StencilTest)
	g.DepthTest(sv.DepthTest)
	g.Blend(sv.Blend)
	g.SampleAlphaToCoverage(sv.SampleAlphaToCoverage)
	g.Multisample(sv.Multisample)

	custom()

	// Clear the saved state.
	g.Saved = nil
	return true
}

// Scissor sets the scissor rectangle, bounds should be the framebuffer's
// bounds (used to convert the scissor rectangle to OpenGL's coordinate
// system).
func (g *GraphicsState) Scissor(bounds, rect image.Rectangle) {
	// If the intersected scissor rectangle is different from the last one then
	// we need to make the OpenGL call.
	rect = bounds.Intersect(rect)
	if noStateGuard || g.S.Scissor != rect {
		g.S.Scissor = rect
		x, y, width, height := glutil.ConvertRect(rect, bounds)
		g.C.gl.Scissor(x, y, width, height)
	}
}

func (g *GraphicsState) getScissor(bounds image.Rectangle) image.Rectangle {
	x, y, width, height := g.C.gl.GetScissorBox()
	return glutil.UnconvertRect(bounds, x, y, width, height)
}

func (g *GraphicsState) ColorWrite(red, green, blue, alpha bool) {
	if noStateGuard || g.S.WriteRed != red || g.S.WriteGreen != green || g.S.WriteBlue != blue || g.S.WriteAlpha != alpha {
		g.S.WriteRed = red
		g.S.WriteGreen = green
		g.S.WriteBlue = blue
		g.S.WriteAlpha = alpha
		g.C.gl.ColorMask(red, green, blue, alpha)
	}
}

func (g *GraphicsState) ClearColor(color gfx.Color) {
	if noStateGuard || g.S.ClearColor != color {
		g.S.ClearColor = color
		g.C.gl.ClearColor(color.R, color.G, color.B, color.A)
	}
}

func (g *GraphicsState) DepthWrite(write bool) {
	if noStateGuard || g.S.DepthWrite != write {
		g.S.DepthWrite = write
		g.C.gl.DepthMask(write)
	}
}

func (g *GraphicsState) ClearDepth(depth float64) {
	if noStateGuard || g.S.ClearDepth != depth {
		g.S.ClearDepth = depth
		g.C.gl.ClearDepth(depth)
	}
}

func (g *GraphicsState) BlendColor(c gfx.Color) {
	if noStateGuard || g.S.State.Blend.Color != c {
		g.S.State.Blend.Color = c
		g.C.gl.BlendColor(c.R, c.G, c.B, c.A)
	}
}

func (g *GraphicsState) ClearStencil(stencil int) {
	if noStateGuard || g.S.ClearStencil != stencil {
		g.S.ClearStencil = stencil
		g.C.gl.ClearStencil(stencil)
	}
}

func (g *GraphicsState) DepthCmp(cmp gfx.Cmp) {
	if noStateGuard || g.S.DepthCmp != cmp {
		g.S.DepthCmp = cmp
		g.C.gl.DepthFunc(g.C.ConvertCmp(cmp))
	}
}

func (g *GraphicsState) FaceCulling(m gfx.FaceCullMode) {
	if noStateGuard || g.S.FaceCulling != m {
		g.S.FaceCulling = m
		switch m {
		case gfx.BackFaceCulling:
			g.C.gl.Enable(g.C.CULL_FACE)
			g.C.gl.CullFace(g.C.BACK)
		case gfx.FrontFaceCulling:
			g.C.gl.Enable(g.C.CULL_FACE)
			g.C.gl.CullFace(g.C.FRONT)
		case gfx.NoFaceCulling:
			g.C.gl.Disable(g.C.CULL_FACE)
		default:
			panic("never here")
		}
	}
}

func (g *GraphicsState) getFaceCulling() gfx.FaceCullMode {
	cullFace := g.C.gl.GetParameterBool(g.C.CULL_FACE)
	if !cullFace {
		return gfx.NoFaceCulling
	}
	mode := g.C.gl.GetParameterInt(g.C.CULL_FACE_MODE)
	switch mode {
	case g.C.FRONT:
		return gfx.FrontFaceCulling
	case g.C.BACK:
		return gfx.BackFaceCulling
	default:
		panic("never here")
	}
}

func (g *GraphicsState) BlendFuncSeparate(bs gfx.BlendState) {
	diff := func(a, b gfx.BlendState) bool {
		return a.SrcRGB != b.SrcRGB || a.DstRGB != b.DstRGB || a.SrcAlpha != b.SrcAlpha || a.DstAlpha != b.DstAlpha
	}

	if noStateGuard || diff(g.S.State.Blend, bs) {
		g.S.State.Blend.SrcRGB = bs.SrcRGB
		g.S.State.Blend.DstRGB = bs.DstRGB
		g.S.State.Blend.SrcAlpha = bs.SrcAlpha
		g.S.State.Blend.DstAlpha = bs.DstAlpha

		g.C.gl.BlendFuncSeparate(
			g.C.ConvertBlendOp(bs.SrcRGB),
			g.C.ConvertBlendOp(bs.DstRGB),
			g.C.ConvertBlendOp(bs.SrcAlpha),
			g.C.ConvertBlendOp(bs.DstAlpha),
		)
	}
}

func (g *GraphicsState) getBlendFuncSeparate(bs *gfx.BlendState) {
	bs.SrcRGB = g.C.UnconvertBlendOp(g.C.gl.GetParameterInt(g.C.BLEND_SRC_RGB))
	bs.DstRGB = g.C.UnconvertBlendOp(g.C.gl.GetParameterInt(g.C.BLEND_DST_RGB))
	bs.SrcAlpha = g.C.UnconvertBlendOp(g.C.gl.GetParameterInt(g.C.BLEND_SRC_ALPHA))
	bs.DstAlpha = g.C.UnconvertBlendOp(g.C.gl.GetParameterInt(g.C.BLEND_DST_ALPHA))
	return
}

func (g *GraphicsState) BlendEquationSeparate(bs gfx.BlendState) {
	if noStateGuard || (g.S.State.Blend.RGBEq != bs.RGBEq || g.S.State.Blend.AlphaEq != bs.AlphaEq) {
		g.S.State.Blend.RGBEq = bs.RGBEq
		g.S.State.Blend.AlphaEq = bs.AlphaEq

		g.C.gl.BlendEquationSeparate(
			g.C.ConvertBlendEq(bs.RGBEq),
			g.C.ConvertBlendEq(bs.AlphaEq),
		)
	}
}

func (g *GraphicsState) getBlendEquationSeparate(bs *gfx.BlendState) {
	bs.RGBEq = g.C.UnconvertBlendEq(g.C.gl.GetParameterInt(g.C.BLEND_EQUATION_RGB))
	bs.AlphaEq = g.C.UnconvertBlendEq(g.C.gl.GetParameterInt(g.C.BLEND_EQUATION_ALPHA))
	return
}

func (g *GraphicsState) StencilOpSeparate(front, back gfx.StencilState) {
	diff := func(a, b gfx.StencilState) bool {
		return a.Fail != b.Fail || a.DepthFail != b.DepthFail || a.DepthPass != b.DepthPass
	}

	if noStateGuard || diff(g.S.StencilFront, front) || diff(g.S.StencilBack, back) {
		g.S.StencilFront.Fail = front.Fail
		g.S.StencilFront.DepthFail = front.DepthFail
		g.S.StencilFront.DepthPass = front.DepthPass

		g.S.StencilBack.Fail = back.Fail
		g.S.StencilBack.DepthFail = back.DepthFail
		g.S.StencilBack.DepthPass = back.DepthPass

		// Save a call if front and back are identical.
		if front == back {
			g.C.gl.StencilOpSeparate(
				g.C.FRONT_AND_BACK,
				g.C.ConvertStencilOp(front.Fail),
				g.C.ConvertStencilOp(front.DepthFail),
				g.C.ConvertStencilOp(front.DepthPass),
			)
			return
		}

		g.C.gl.StencilOpSeparate(
			g.C.FRONT,
			g.C.ConvertStencilOp(front.Fail),
			g.C.ConvertStencilOp(front.DepthFail),
			g.C.ConvertStencilOp(front.DepthPass),
		)
		g.C.gl.StencilOpSeparate(
			g.C.BACK,
			g.C.ConvertStencilOp(back.Fail),
			g.C.ConvertStencilOp(back.DepthFail),
			g.C.ConvertStencilOp(back.DepthPass),
		)
	}
}

func (g *GraphicsState) getStencilOpSeparate(front, back *gfx.StencilState) {
	front.Fail = g.C.UnconvertStencilOp(g.C.gl.GetParameterInt(g.C.STENCIL_FAIL))
	front.DepthFail = g.C.UnconvertStencilOp(g.C.gl.GetParameterInt(g.C.STENCIL_PASS_DEPTH_FAIL))
	front.DepthPass = g.C.UnconvertStencilOp(g.C.gl.GetParameterInt(g.C.STENCIL_PASS_DEPTH_PASS))

	back.Fail = g.C.UnconvertStencilOp(g.C.gl.GetParameterInt(g.C.STENCIL_BACK_FAIL))
	back.DepthFail = g.C.UnconvertStencilOp(g.C.gl.GetParameterInt(g.C.STENCIL_BACK_PASS_DEPTH_FAIL))
	back.DepthPass = g.C.UnconvertStencilOp(g.C.gl.GetParameterInt(g.C.STENCIL_BACK_PASS_DEPTH_PASS))
}

func (g *GraphicsState) Dithering(v bool) {
	if noStateGuard || g.S.Dithering != v {
		g.S.Dithering = v
		g.C.Feature(g.C.DITHER, v)
	}
}

func (g *GraphicsState) ScissorTest(v bool) {
	if noStateGuard || g.S.ScissorTest != v {
		g.S.ScissorTest = v
		g.C.Feature(g.C.SCISSOR_TEST, v)
	}
}

func (g *GraphicsState) StencilTest(v bool) {
	if noStateGuard || g.S.StencilTest != v {
		g.S.StencilTest = v
		g.C.Feature(g.C.STENCIL_TEST, v)
	}
}

func (g *GraphicsState) DepthTest(v bool) {
	if noStateGuard || g.S.DepthTest != v {
		g.S.DepthTest = v
		g.C.Feature(g.C.DEPTH_TEST, v)
	}
}

func (g *GraphicsState) Blend(v bool) {
	if noStateGuard || g.S.Blend != v {
		g.S.Blend = v
		g.C.Feature(g.C.BLEND, v)
	}
}

func (g *GraphicsState) SampleAlphaToCoverage(v bool) {
	if noStateGuard || g.S.SampleAlphaToCoverage != v {
		g.S.SampleAlphaToCoverage = v
		g.C.Feature(g.C.SAMPLE_ALPHA_TO_COVERAGE, v)
	}
}

func (g *GraphicsState) Multisample(v bool) {
	if noStateGuard || g.S.Multisample != v {
		g.S.Multisample = v
		g.C.Feature(g.C.MULTISAMPLE, v)
	}
}

func NewGraphicsState(c *Context) *GraphicsState {
	cpy := *glutil.DefaultCommonState
	return &GraphicsState{
		C: c,
		S: &cpy,
	}
}
