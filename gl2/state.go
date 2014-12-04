// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"image"

	"azul3d.org/gfx.v2-dev"
	"azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"
	"azul3d.org/gfx.v2-dev/internal/glutil"
)

// Set this to true to disable state guarding (i.e. avoiding useless OpenGL
// state calls). This is useful for debugging the state guard code.
const noStateGuard = false

// Please ensure these values match the default OpenGL state values listed in
// the OpenGL documentation.
var defaultGraphicsState = &graphicsState{
	image.Rect(0, 0, 0, 0),                    // scissor - Whole screen
	gfx.Color{R: 0.0, G: 0.0, B: 0.0, A: 0.0}, // clear color
	glutil.DefaultBlendState.Color,            // blend color
	1.0, // clear depth
	0,   // clear stencil
	[4]bool{true, true, true, true}, // color write
	gfx.Less,                        // depth func
	glutil.DefaultBlendState,        // blend func seperate
	glutil.DefaultBlendState,        // blend equation seperate
	glutil.DefaultStencilState,      // stencil front
	glutil.DefaultStencilState,      // stencil back
	0xFFFF,            // stencil mask front
	0xFFFF,            // stencil mask back
	true,              // dithering
	false,             // depth test
	true,              // depth write
	false,             // stencil test
	false,             // blend
	false,             // alpha to coverage
	gfx.NoFaceCulling, // face culling
	0,                 // program
}

// Queries the existing OpenGL graphics state and returns it.
func queryExistingState(gpuInfo *gfx.GPUInfo, bounds image.Rectangle) *graphicsState {
	var (
		scissor                      [4]int32
		clearColor, blendColor       gfx.Color
		clearDepth                   float64
		clearStencil                 int32
		colorWrite                   [4]bool
		depthFunc                    int32
		blendDstRGB, blendSrcRGB     int32
		blendDstAlpha, blendSrcAlpha int32
		blendEqRGB, blendEqAlpha     int32

		stencilFrontWriteMask, stencilFrontReadMask, stencilFrontRef,
		stencilFrontOpFail, stencilFrontOpDepthFail, stencilFrontOpDepthPass,
		stencilFrontCmp int32

		stencilBackWriteMask, stencilBackReadMask, stencilBackRef,
		stencilBackOpFail, stencilBackOpDepthFail, stencilBackOpDepthPass,
		stencilBackCmp int32

		dithering, depthTest, depthWrite, stencilTest, blend,
		alphaToCoverage bool
		faceCullMode int32
	)
	gl.GetIntegerv(gl.SCISSOR_BOX, &scissor[0])
	gl.GetFloatv(gl.COLOR_CLEAR_VALUE, &clearColor.R)
	gl.GetFloatv(gl.BLEND_COLOR, &blendColor.R)
	gl.GetDoublev(gl.DEPTH_CLEAR_VALUE, &clearDepth)
	gl.GetIntegerv(gl.STENCIL_CLEAR_VALUE, &clearStencil)
	gl.GetBooleanv(gl.COLOR_WRITEMASK, &colorWrite[0])
	gl.GetIntegerv(gl.DEPTH_FUNC, &depthFunc)
	gl.GetIntegerv(gl.BLEND_DST_RGB, &blendDstRGB)
	gl.GetIntegerv(gl.BLEND_SRC_RGB, &blendSrcRGB)
	gl.GetIntegerv(gl.BLEND_DST_ALPHA, &blendDstAlpha)
	gl.GetIntegerv(gl.BLEND_SRC_ALPHA, &blendSrcAlpha)
	gl.GetIntegerv(gl.BLEND_EQUATION_RGB, &blendEqRGB)
	gl.GetIntegerv(gl.BLEND_EQUATION_ALPHA, &blendEqAlpha)
	gl.GetIntegerv(gl.BLEND_EQUATION_RGB, &blendEqRGB)
	gl.GetIntegerv(gl.BLEND_EQUATION_ALPHA, &blendEqAlpha)

	gl.GetIntegerv(gl.STENCIL_FAIL, &stencilFrontOpFail)
	gl.GetIntegerv(gl.STENCIL_PASS_DEPTH_FAIL, &stencilFrontOpDepthFail)
	gl.GetIntegerv(gl.STENCIL_PASS_DEPTH_PASS, &stencilFrontOpDepthPass)
	gl.GetIntegerv(gl.STENCIL_WRITEMASK, &stencilFrontWriteMask)
	gl.GetIntegerv(gl.STENCIL_VALUE_MASK, &stencilFrontReadMask)
	gl.GetIntegerv(gl.STENCIL_REF, &stencilFrontRef)
	gl.GetIntegerv(gl.STENCIL_FUNC, &stencilFrontCmp)

	gl.GetIntegerv(gl.STENCIL_BACK_FAIL, &stencilBackOpFail)
	gl.GetIntegerv(gl.STENCIL_BACK_PASS_DEPTH_FAIL, &stencilBackOpDepthFail)
	gl.GetIntegerv(gl.STENCIL_BACK_PASS_DEPTH_PASS, &stencilBackOpDepthPass)
	gl.GetIntegerv(gl.STENCIL_BACK_WRITEMASK, &stencilBackWriteMask)
	gl.GetIntegerv(gl.STENCIL_BACK_VALUE_MASK, &stencilBackReadMask)
	gl.GetIntegerv(gl.STENCIL_BACK_REF, &stencilBackRef)
	gl.GetIntegerv(gl.STENCIL_BACK_FUNC, &stencilBackCmp)

	gl.GetBooleanv(gl.DITHER, &dithering)
	gl.GetBooleanv(gl.DEPTH_TEST, &depthTest)
	gl.GetBooleanv(gl.DEPTH_WRITEMASK, &depthWrite)
	gl.GetBooleanv(gl.STENCIL_TEST, &stencilTest)
	gl.GetBooleanv(gl.BLEND, &blend)
	if gpuInfo.AlphaToCoverage {
		gl.GetBooleanv(gl.SAMPLE_ALPHA_TO_COVERAGE, &alphaToCoverage)
	}

	gl.GetIntegerv(gl.CULL_FACE_MODE, &faceCullMode)
	//gl.Execute()

	return &graphicsState{
		scissor:      glutil.UnconvertRect(bounds, scissor[0], scissor[1], scissor[2], scissor[3]),
		clearColor:   clearColor,
		blendColor:   blendColor,
		clearDepth:   clearDepth,
		clearStencil: int(clearStencil),
		colorWrite:   colorWrite,
		depthFunc:    unconvertCmp(depthFunc),
		blendFuncSeparate: gfx.BlendState{
			DstRGB:   unconvertBlendOp(blendDstRGB),
			SrcRGB:   unconvertBlendOp(blendSrcRGB),
			DstAlpha: unconvertBlendOp(blendDstAlpha),
			SrcAlpha: unconvertBlendOp(blendSrcAlpha),
		},
		blendEquationSeparate: gfx.BlendState{
			RGBEq:   unconvertBlendEq(blendEqRGB),
			AlphaEq: unconvertBlendEq(blendEqAlpha),
		},
		stencilFront: gfx.StencilState{
			Fail:      unconvertStencilOp(stencilFrontOpFail),
			DepthFail: unconvertStencilOp(stencilFrontOpDepthFail),
			DepthPass: unconvertStencilOp(stencilFrontOpDepthPass),
			Cmp:       unconvertCmp(stencilFrontCmp),
			Reference: uint(stencilFrontRef),
			ReadMask:  uint(stencilFrontReadMask),
		},
		stencilBack: gfx.StencilState{
			Fail:      unconvertStencilOp(stencilBackOpFail),
			DepthFail: unconvertStencilOp(stencilBackOpDepthFail),
			DepthPass: unconvertStencilOp(stencilBackOpDepthPass),
			Cmp:       unconvertCmp(stencilBackCmp),
			Reference: uint(stencilBackRef),
			ReadMask:  uint(stencilBackReadMask),
		},
		stencilMaskFront: uint(stencilFrontWriteMask),
		stencilMaskBack:  uint(stencilBackWriteMask),
		dithering:        dithering,
		depthTest:        depthTest,
		depthWrite:       depthWrite,
		stencilTest:      stencilTest,
		blend:            blend,
		alphaToCoverage:  alphaToCoverage,
		faceCulling:      unconvertFaceCull(faceCullMode),
	}

	/*
		program                                                                          uint32
	*/
}

// Structure for various previously set render states, used to avoid uselessly
// setting OpenGL state twice and keeping state between frames if needed for
// interoperability with, e.g. QT5's renderer.
type graphicsState struct {
	scissor                                                               image.Rectangle
	clearColor, blendColor                                                gfx.Color
	clearDepth                                                            float64
	clearStencil                                                          int
	colorWrite                                                            [4]bool
	depthFunc                                                             gfx.Cmp
	blendFuncSeparate, blendEquationSeparate                              gfx.BlendState
	stencilFront, stencilBack                                             gfx.StencilState
	stencilMaskFront, stencilMaskBack                                     uint
	dithering, depthTest, depthWrite, stencilTest, blend, alphaToCoverage bool
	faceCulling                                                           gfx.FaceCullMode
	program                                                               uint32
}

// loads the graphics state, g, making OpenGL calls as neccesarry to components
// that differ between the states s and r.
//
// bounds is the renderer's bounds (e.g. r.Bounds()) to pass into stateScissor().
func (s *graphicsState) load(gpuInfo *gfx.GPUInfo, bounds image.Rectangle, g *graphicsState) {
	s.stateScissor(bounds, g.scissor)
	s.stateClearColor(g.clearColor)
	s.stateBlendColor(g.blendColor)
	s.stateClearDepth(g.clearDepth)
	s.stateClearStencil(g.clearStencil)
	s.stateColorWrite(g.colorWrite)
	s.stateDepthFunc(g.depthFunc)
	s.stateBlendFuncSeparate(g.blendFuncSeparate)
	s.stateBlendEquationSeparate(g.blendEquationSeparate)
	s.stateStencilOp(g.stencilFront, g.stencilBack)
	s.stateStencilFunc(g.stencilFront, g.stencilBack)
	s.stateStencilMask(g.stencilMaskFront, g.stencilMaskBack)
	s.stateDithering(g.dithering)
	s.stateDepthTest(g.depthTest)
	s.stateDepthWrite(g.depthWrite)
	s.stateStencilTest(g.stencilTest)
	s.stateBlend(g.blend)
	s.stateAlphaToCoverage(gpuInfo, g.alphaToCoverage)
	s.stateFaceCulling(g.faceCulling)
	s.stateProgram(g.program)
}

// bounds is the renderer's bounds (e.g. r.Bounds()).
func (s *graphicsState) stateScissor(bounds, rect image.Rectangle) {
	// Only if the (final) scissor rectangle has changed do we need to make the
	// OpenGL call.

	// If the rectangle is empty use the entire area.
	if rect.Empty() {
		rect = bounds
	} else {
		// Intersect the rectangle with the renderer's bounds.
		rect = bounds.Intersect(rect)
	}

	if noStateGuard || s.scissor != rect {
		// Store the new scissor rectangle.
		s.scissor = rect
		x, y, width, height := glutil.ConvertRect(rect, bounds)
		gl.Scissor(x, y, width, height)
	}
}

func (s *graphicsState) stateClearColor(color gfx.Color) {
	if noStateGuard || s.clearColor != color {
		s.clearColor = color
		gl.ClearColor(color.R, color.G, color.B, color.A)
	}
}

func (s *graphicsState) stateBlendColor(c gfx.Color) {
	if noStateGuard || s.blendColor != c {
		s.blendColor = c
		gl.BlendColor(c.R, c.G, c.B, c.A)
	}
}

func (s *graphicsState) stateClearDepth(depth float64) {
	if noStateGuard || s.clearDepth != depth {
		s.clearDepth = depth
		gl.ClearDepth(depth)
	}
}

func (s *graphicsState) stateClearStencil(stencil int) {
	if noStateGuard || s.clearStencil != stencil {
		s.clearStencil = stencil
		gl.ClearStencil(int32(stencil))
	}
}

func (s *graphicsState) stateColorWrite(cw [4]bool) {
	if noStateGuard || s.colorWrite != cw {
		s.colorWrite = cw
		gl.ColorMask(
			cw[0],
			cw[1],
			cw[2],
			cw[3],
		)
	}
}

func (s *graphicsState) stateDepthFunc(df gfx.Cmp) {
	if noStateGuard || s.depthFunc != df {
		s.depthFunc = df
		gl.DepthFunc(convertCmp(df))
	}
}

func (s *graphicsState) stateBlendFuncSeparate(bs gfx.BlendState) {
	if noStateGuard || s.blendFuncSeparate != bs {
		s.blendFuncSeparate = bs
		gl.BlendFuncSeparate(
			convertBlendOp(bs.SrcRGB),
			convertBlendOp(bs.DstRGB),
			convertBlendOp(bs.SrcAlpha),
			convertBlendOp(bs.SrcAlpha),
		)
	}
}

func (s *graphicsState) stateBlendEquationSeparate(bs gfx.BlendState) {
	if noStateGuard || s.blendEquationSeparate != bs {
		s.blendEquationSeparate = bs
		gl.BlendEquationSeparate(
			convertBlendEq(bs.RGBEq),
			convertBlendEq(bs.AlphaEq),
		)
	}
}

func (s *graphicsState) stateStencilOp(front, back gfx.StencilState) {
	diff := func(a, b gfx.StencilState) bool {
		return a.Fail != b.Fail || a.DepthFail != b.DepthFail || a.DepthPass != b.DepthPass
	}

	if noStateGuard || diff(s.stencilFront, front) || diff(s.stencilBack, back) {
		s.stencilFront.Fail = front.Fail
		s.stencilFront.DepthFail = front.DepthFail
		s.stencilFront.DepthPass = front.DepthPass

		s.stencilBack.Fail = back.Fail
		s.stencilBack.DepthFail = back.DepthFail
		s.stencilBack.DepthPass = back.DepthPass

		if front == back {
			// We can save a few calls.
			gl.StencilOpSeparate(
				gl.FRONT_AND_BACK,
				convertStencilOp(front.Fail),
				convertStencilOp(front.DepthFail),
				convertStencilOp(front.DepthPass),
			)
		} else {
			gl.StencilOpSeparate(
				gl.FRONT,
				convertStencilOp(front.Fail),
				convertStencilOp(front.DepthFail),
				convertStencilOp(front.DepthPass),
			)
			gl.StencilOpSeparate(
				gl.BACK,
				convertStencilOp(back.Fail),
				convertStencilOp(back.DepthFail),
				convertStencilOp(back.DepthPass),
			)
		}
	}
}

func (s *graphicsState) stateStencilFunc(front, back gfx.StencilState) {
	diff := func(a, b gfx.StencilState) bool {
		return a.Cmp != b.Cmp || a.Reference != b.Reference || a.ReadMask != b.ReadMask
	}

	if noStateGuard || diff(s.stencilFront, front) || diff(s.stencilBack, back) {
		s.stencilFront.Cmp = front.Cmp
		s.stencilFront.Reference = front.Reference
		s.stencilFront.ReadMask = front.ReadMask

		s.stencilBack.Cmp = back.Cmp
		s.stencilBack.Reference = back.Reference
		s.stencilBack.ReadMask = back.ReadMask

		if front == back {
			// We can save a few calls.
			gl.StencilFuncSeparate(
				gl.FRONT_AND_BACK,
				convertCmp(front.Cmp),
				int32(front.Reference),
				uint32(front.ReadMask),
			)
		} else {
			gl.StencilFuncSeparate(
				gl.FRONT,
				convertCmp(front.Cmp),
				int32(front.Reference),
				uint32(front.ReadMask),
			)
			gl.StencilFuncSeparate(
				gl.BACK,
				convertCmp(back.Cmp),
				int32(back.Reference),
				uint32(back.ReadMask),
			)
		}
	}
}

func (s *graphicsState) stateStencilMask(front, back uint) {
	if noStateGuard || s.stencilMaskFront != front || s.stencilMaskBack != back {
		s.stencilMaskFront = front
		s.stencilMaskBack = back
		if front == back {
			// We can save a call.
			gl.StencilMaskSeparate(gl.FRONT_AND_BACK, uint32(front))
		} else {
			gl.StencilMaskSeparate(gl.FRONT, uint32(front))
			gl.StencilMaskSeparate(gl.BACK, uint32(back))
		}
	}
}

func (s *graphicsState) stateDithering(enabled bool) {
	if noStateGuard || s.dithering != enabled {
		s.dithering = enabled
		if enabled {
			gl.Enable(gl.DITHER)
		} else {
			gl.Disable(gl.DITHER)
		}
	}
}

func (s *graphicsState) stateDepthTest(enabled bool) {
	if noStateGuard || s.depthTest != enabled {
		s.depthTest = enabled
		if enabled {
			gl.Enable(gl.DEPTH_TEST)
		} else {
			gl.Disable(gl.DEPTH_TEST)
		}
	}
}

func (s *graphicsState) stateDepthWrite(enabled bool) {
	if noStateGuard || s.depthWrite != enabled {
		s.depthWrite = enabled
		if enabled {
			gl.DepthMask(true)
		} else {
			gl.DepthMask(false)
		}
	}
}

func (s *graphicsState) stateStencilTest(stencilTest bool) {
	if noStateGuard || s.stencilTest != stencilTest {
		s.stencilTest = stencilTest
		if stencilTest {
			gl.Enable(gl.STENCIL_TEST)
		} else {
			gl.Disable(gl.STENCIL_TEST)
		}
	}
}

func (s *graphicsState) stateBlend(blend bool) {
	if noStateGuard || s.blend != blend {
		s.blend = blend
		if blend {
			gl.Enable(gl.BLEND)
		} else {
			gl.Disable(gl.BLEND)
		}
	}
}

func (s *graphicsState) stateAlphaToCoverage(gpuInfo *gfx.GPUInfo, alphaToCoverage bool) {
	if noStateGuard || s.alphaToCoverage != alphaToCoverage {
		s.alphaToCoverage = alphaToCoverage
		if gpuInfo.AlphaToCoverage {
			if alphaToCoverage {
				gl.Enable(gl.SAMPLE_ALPHA_TO_COVERAGE)
			} else {
				gl.Disable(gl.SAMPLE_ALPHA_TO_COVERAGE)
			}
		}
	}
}

func (s *graphicsState) stateFaceCulling(m gfx.FaceCullMode) {
	if noStateGuard || s.faceCulling != m {
		s.faceCulling = m
		switch m {
		case gfx.BackFaceCulling:
			gl.Enable(gl.CULL_FACE)
			gl.CullFace(gl.BACK)
		case gfx.FrontFaceCulling:
			gl.Enable(gl.CULL_FACE)
			gl.CullFace(gl.FRONT)
		default:
			gl.Disable(gl.CULL_FACE)
		}
	}
}

func (s *graphicsState) stateProgram(p uint32) {
	if noStateGuard || s.program != p {
		s.program = p
		gl.UseProgram(p)
	}
}
