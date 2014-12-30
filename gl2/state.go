// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"image"

	"azul3d.org/gfx.v2-dev"
	"azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"
	"azul3d.org/gfx.v2-dev/internal/glutil"
	"azul3d.org/gfx.v2-dev/internal/tag"
)

// Set this to true to disable state guarding (i.e. avoiding useless OpenGL
// state calls). This is useful for debugging the state guard code.
const noStateGuard = tag.Gsgdebug

// glFeature enables or disables the given feature depending on the given
// boolean.
func glFeature(feature uint32, enabled bool) {
	if enabled {
		gl.Enable(feature)
		return
	}
	gl.Disable(feature)
}

// Please ensure these values match the default OpenGL state values listed in
// the OpenGL documentation.
var defaultGraphicsState = &graphicsState{
	glutil.DefaultCommonState,
	glutil.DefaultState,
	false, // alpha to coverage
}

// queryStencilFrontState queries the existing front-face stencil graphics
// state from OpenGL and returns it.
func queryStencilFrontState() gfx.StencilState {
	var (
		stencilFrontWriteMask, stencilFrontReadMask, stencilFrontRef,
		stencilFrontOpFail, stencilFrontOpDepthFail, stencilFrontOpDepthPass,
		stencilFrontCmp int32
	)

	gl.GetIntegerv(gl.STENCIL_FAIL, &stencilFrontOpFail)
	gl.GetIntegerv(gl.STENCIL_PASS_DEPTH_FAIL, &stencilFrontOpDepthFail)
	gl.GetIntegerv(gl.STENCIL_PASS_DEPTH_PASS, &stencilFrontOpDepthPass)
	gl.GetIntegerv(gl.STENCIL_WRITEMASK, &stencilFrontWriteMask)
	gl.GetIntegerv(gl.STENCIL_VALUE_MASK, &stencilFrontReadMask)
	gl.GetIntegerv(gl.STENCIL_REF, &stencilFrontRef)
	gl.GetIntegerv(gl.STENCIL_FUNC, &stencilFrontCmp)

	return gfx.StencilState{
		uint(stencilFrontWriteMask),
		uint(stencilFrontReadMask),
		uint(stencilFrontRef),
		unconvertStencilOp(stencilFrontOpFail),
		unconvertStencilOp(stencilFrontOpDepthFail),
		unconvertStencilOp(stencilFrontOpDepthPass),
		unconvertCmp(stencilFrontCmp),
	}
}

// queryStencilBackState queries the existing back-face stencil graphics state
// from OpenGL and returns it.
func queryStencilBackState() gfx.StencilState {
	var (
		stencilBackWriteMask, stencilBackReadMask, stencilBackRef,
		stencilBackOpFail, stencilBackOpDepthFail, stencilBackOpDepthPass,
		stencilBackCmp int32
	)

	gl.GetIntegerv(gl.STENCIL_BACK_FAIL, &stencilBackOpFail)
	gl.GetIntegerv(gl.STENCIL_BACK_PASS_DEPTH_FAIL, &stencilBackOpDepthFail)
	gl.GetIntegerv(gl.STENCIL_BACK_PASS_DEPTH_PASS, &stencilBackOpDepthPass)
	gl.GetIntegerv(gl.STENCIL_BACK_WRITEMASK, &stencilBackWriteMask)
	gl.GetIntegerv(gl.STENCIL_BACK_VALUE_MASK, &stencilBackReadMask)
	gl.GetIntegerv(gl.STENCIL_BACK_REF, &stencilBackRef)
	gl.GetIntegerv(gl.STENCIL_BACK_FUNC, &stencilBackCmp)

	return gfx.StencilState{
		uint(stencilBackWriteMask),
		uint(stencilBackReadMask),
		uint(stencilBackRef),
		unconvertStencilOp(stencilBackOpFail),
		unconvertStencilOp(stencilBackOpDepthFail),
		unconvertStencilOp(stencilBackOpDepthPass),
		unconvertCmp(stencilBackCmp),
	}
}

// queryBlendState queries the existing blend graphics state from OpenGL and
// returns it.
func queryBlendState() gfx.BlendState {
	var (
		blendColor                   gfx.Color
		blendDstRGB, blendSrcRGB     int32
		blendDstAlpha, blendSrcAlpha int32
		blendEqRGB, blendEqAlpha     int32
	)

	gl.GetFloatv(gl.BLEND_COLOR, &blendColor.R)
	gl.GetIntegerv(gl.BLEND_DST_RGB, &blendDstRGB)
	gl.GetIntegerv(gl.BLEND_SRC_RGB, &blendSrcRGB)
	gl.GetIntegerv(gl.BLEND_DST_ALPHA, &blendDstAlpha)
	gl.GetIntegerv(gl.BLEND_SRC_ALPHA, &blendSrcAlpha)
	gl.GetIntegerv(gl.BLEND_EQUATION_RGB, &blendEqRGB)
	gl.GetIntegerv(gl.BLEND_EQUATION_ALPHA, &blendEqAlpha)
	gl.GetIntegerv(gl.BLEND_EQUATION_RGB, &blendEqRGB)
	gl.GetIntegerv(gl.BLEND_EQUATION_ALPHA, &blendEqAlpha)

	return gfx.BlendState{
		blendColor,
		unconvertBlendOp(blendSrcRGB),
		unconvertBlendOp(blendDstRGB),
		unconvertBlendOp(blendSrcAlpha),
		unconvertBlendOp(blendDstAlpha),
		unconvertBlendEq(blendEqRGB),
		unconvertBlendEq(blendEqAlpha),
	}
}

// queryState queries the existing OpenGL gfx.State and returns it.
func queryState(gpuInfo *gfx.DeviceInfo) *gfx.State {
	var (
		colorWrite                                    [4]bool
		depthClamp                                    bool
		depthFunc                                     int32
		dithering, depthTest, depthWrite, stencilTest bool
		faceCullMode                                  int32
	)
	gl.GetBooleanv(gl.COLOR_WRITEMASK, &colorWrite[0])
	gl.GetIntegerv(gl.DEPTH_FUNC, &depthFunc)

	gl.GetBooleanv(gl.DITHER, &dithering)
	if gpuInfo.DepthClamp {
		gl.GetBooleanv(gl.DEPTH_CLAMP, &depthClamp)
	}
	gl.GetBooleanv(gl.DEPTH_TEST, &depthTest)
	gl.GetBooleanv(gl.DEPTH_WRITEMASK, &depthWrite)
	gl.GetBooleanv(gl.STENCIL_TEST, &stencilTest)
	gl.GetIntegerv(gl.CULL_FACE_MODE, &faceCullMode)
	//gl.Execute()

	return &gfx.State{
		FaceCulling:  unconvertFaceCull(faceCullMode),
		Blend:        queryBlendState(),
		StencilFront: queryStencilFrontState(),
		StencilBack:  queryStencilBackState(),
		DepthCmp:     unconvertCmp(depthFunc),
		WriteRed:     colorWrite[0],
		WriteGreen:   colorWrite[1],
		WriteBlue:    colorWrite[2],
		WriteAlpha:   colorWrite[3],
		Dithering:    dithering,
		DepthClamp:   depthClamp,
		DepthTest:    depthTest,
		DepthWrite:   depthWrite,
		StencilTest:  stencilTest,
	}
}

// queryGraphicsState queries the existing OpenGL graphics state and returns
// it.
func queryGraphicsState(gpuInfo *gfx.DeviceInfo, bounds image.Rectangle) *graphicsState {
	var (
		scissor                                           [4]int32
		clearColor                                        gfx.Color
		clearDepth                                        float64
		clearStencil                                      int32
		scissorTest, blend                                bool
		multisample, programPointSizeExt, alphaToCoverage bool
		shaderProgram                                     int32
	)
	gl.GetIntegerv(gl.SCISSOR_BOX, &scissor[0])
	gl.GetFloatv(gl.COLOR_CLEAR_VALUE, &clearColor.R)
	gl.GetDoublev(gl.DEPTH_CLEAR_VALUE, &clearDepth)
	gl.GetIntegerv(gl.STENCIL_CLEAR_VALUE, &clearStencil)

	gl.GetBooleanv(gl.SCISSOR_TEST, &scissorTest)
	gl.GetBooleanv(gl.BLEND, &blend)
	gl.GetBooleanv(gl.MULTISAMPLE, &multisample)
	gl.GetBooleanv(gl.PROGRAM_POINT_SIZE_EXT, &programPointSizeExt)
	if gpuInfo.AlphaToCoverage {
		gl.GetBooleanv(gl.SAMPLE_ALPHA_TO_COVERAGE, &alphaToCoverage)
	}
	gl.GetIntegerv(gl.CURRENT_PROGRAM, &shaderProgram)
	//gl.Execute()

	return &graphicsState{
		&glutil.CommonState{
			glutil.UnconvertRect(bounds, scissor[0], scissor[1], scissor[2], scissor[3]),
			clearColor,
			clearDepth,
			int(clearStencil),
			blend,
			scissorTest,
			programPointSizeExt,
			multisample,
			uint32(shaderProgram),
		},
		queryState(gpuInfo),
		alphaToCoverage,
	}
}

// Structure for various previously set render states, used to avoid uselessly
// setting OpenGL state twice and keeping state between frames if needed for
// interoperability with, e.g. QT5's renderer.
type graphicsState struct {
	*glutil.CommonState
	*gfx.State
	alphaToCoverage bool
}

// loads the graphics state, g, making OpenGL calls as necessary to components
// that differ between the states s and r.
//
// bounds is the renderer's bounds (e.g. r.Bounds) to pass into stateScissor.
func (s *graphicsState) load(gpuInfo *gfx.DeviceInfo, bounds image.Rectangle, g *graphicsState) {
	s.stateScissor(bounds, g.CommonState.Scissor)
	s.stateClearColor(g.CommonState.ClearColor)
	s.stateBlendColor(g.State.Blend.Color)
	s.stateClearDepth(g.CommonState.ClearDepth)
	s.stateClearStencil(g.CommonState.ClearStencil)
	s.stateColorWrite(g.State.WriteRed, g.State.WriteGreen, g.State.WriteBlue, g.State.WriteAlpha)
	s.stateDepthClamp(gpuInfo, g.State.DepthClamp)
	s.stateDepthFunc(g.State.DepthCmp)
	s.stateBlendFuncSeparate(g.State.Blend)
	s.stateBlendEquationSeparate(g.State.Blend)
	s.stateStencilOp(g.State.StencilFront, g.State.StencilBack)
	s.stateStencilFunc(g.State.StencilFront, g.State.StencilBack)
	s.stateStencilMask(g.State.StencilFront.WriteMask, g.State.StencilBack.WriteMask)
	s.stateDithering(g.State.Dithering)
	s.stateDepthTest(g.State.DepthTest)
	s.stateDepthWrite(g.State.DepthWrite)
	s.stateStencilTest(g.State.StencilTest)
	s.stateScissorTest(g.CommonState.ScissorTest)
	s.stateBlend(g.CommonState.Blend)
	s.stateProgramPointSizeExt(g.CommonState.ProgramPointSizeExt)
	s.stateMultisample(g.CommonState.Multisample)
	s.stateAlphaToCoverage(gpuInfo, g.alphaToCoverage)
	s.stateFaceCulling(g.State.FaceCulling)
	s.stateProgram(g.CommonState.ShaderProgram)
}

// bounds is the renderer's bounds (e.g. r.Bounds()).
func (s *graphicsState) stateScissor(bounds, rect image.Rectangle) {
	// Only if the intersected scissor rectangle has changed do we need to make
	// the OpenGL call.
	rect = bounds.Intersect(rect)

	if noStateGuard || s.CommonState.Scissor != rect {
		// Store the new scissor rectangle.
		s.CommonState.Scissor = rect
		x, y, width, height := glutil.ConvertRect(rect, bounds)
		gl.Scissor(x, y, width, height)
	}
}

func (s *graphicsState) stateClearColor(color gfx.Color) {
	if noStateGuard || s.CommonState.ClearColor != color {
		s.CommonState.ClearColor = color
		gl.ClearColor(color.R, color.G, color.B, color.A)
	}
}

func (s *graphicsState) stateBlendColor(c gfx.Color) {
	if noStateGuard || s.State.Blend.Color != c {
		s.State.Blend.Color = c
		gl.BlendColor(c.R, c.G, c.B, c.A)
	}
}

func (s *graphicsState) stateClearDepth(depth float64) {
	if noStateGuard || s.CommonState.ClearDepth != depth {
		s.CommonState.ClearDepth = depth
		gl.ClearDepth(depth)
	}
}

func (s *graphicsState) stateClearStencil(stencil int) {
	if noStateGuard || s.CommonState.ClearStencil != stencil {
		s.CommonState.ClearStencil = stencil
		gl.ClearStencil(int32(stencil))
	}
}

func (s *graphicsState) stateColorWrite(red, green, blue, alpha bool) {
	if noStateGuard || (s.State.WriteRed != red || s.State.WriteGreen != green || s.State.WriteBlue != blue || s.State.WriteAlpha != alpha) {
		s.State.WriteRed = red
		s.State.WriteGreen = green
		s.State.WriteBlue = blue
		s.State.WriteAlpha = alpha
		gl.ColorMask(red, green, blue, alpha)
	}
}

func (s *graphicsState) stateDepthFunc(cmp gfx.Cmp) {
	if noStateGuard || s.State.DepthCmp != cmp {
		s.State.DepthCmp = cmp
		gl.DepthFunc(convertCmp(cmp))
	}
}

func (s *graphicsState) stateBlendFuncSeparate(bs gfx.BlendState) {
	diff := func(a, b gfx.BlendState) bool {
		return a.SrcRGB != b.SrcRGB || a.DstRGB != b.DstRGB || a.SrcAlpha != b.SrcAlpha
	}

	if noStateGuard || diff(s.State.Blend, bs) {
		s.State.Blend.SrcRGB = bs.SrcRGB
		s.State.Blend.DstRGB = bs.DstRGB
		s.State.Blend.SrcAlpha = bs.SrcAlpha

		gl.BlendFuncSeparate(
			convertBlendOp(bs.SrcRGB),
			convertBlendOp(bs.DstRGB),
			convertBlendOp(bs.SrcAlpha),
			convertBlendOp(bs.SrcAlpha),
		)
	}
}

func (s *graphicsState) stateBlendEquationSeparate(bs gfx.BlendState) {
	if noStateGuard || (s.State.Blend.RGBEq != bs.RGBEq || s.State.Blend.AlphaEq != bs.AlphaEq) {
		s.State.Blend.RGBEq = bs.RGBEq
		s.State.Blend.AlphaEq = bs.AlphaEq

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

	if noStateGuard || diff(s.State.StencilFront, front) || diff(s.State.StencilBack, back) {
		s.State.StencilFront.Fail = front.Fail
		s.State.StencilFront.DepthFail = front.DepthFail
		s.State.StencilFront.DepthPass = front.DepthPass

		s.State.StencilBack.Fail = back.Fail
		s.State.StencilBack.DepthFail = back.DepthFail
		s.State.StencilBack.DepthPass = back.DepthPass

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

	if noStateGuard || diff(s.State.StencilFront, front) || diff(s.State.StencilBack, back) {
		s.State.StencilFront.Cmp = front.Cmp
		s.State.StencilFront.Reference = front.Reference
		s.State.StencilFront.ReadMask = front.ReadMask

		s.State.StencilBack.Cmp = back.Cmp
		s.State.StencilBack.Reference = back.Reference
		s.State.StencilBack.ReadMask = back.ReadMask

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
	if noStateGuard || s.State.StencilFront.WriteMask != front || s.State.StencilBack.WriteMask != back {
		s.State.StencilFront.WriteMask = front
		s.State.StencilBack.WriteMask = back
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
	if noStateGuard || s.State.Dithering != enabled {
		s.State.Dithering = enabled
		glFeature(gl.DITHER, enabled)
	}
}

func (s *graphicsState) stateDepthTest(enabled bool) {
	if noStateGuard || s.State.DepthTest != enabled {
		s.State.DepthTest = enabled
		glFeature(gl.DEPTH_TEST, enabled)
	}
}

func (s *graphicsState) stateDepthWrite(enabled bool) {
	if noStateGuard || s.State.DepthWrite != enabled {
		s.State.DepthWrite = enabled
		gl.DepthMask(enabled)
	}
}

func (s *graphicsState) stateStencilTest(stencilTest bool) {
	if noStateGuard || s.State.StencilTest != stencilTest {
		s.State.StencilTest = stencilTest
		glFeature(gl.STENCIL_TEST, stencilTest)
	}
}

func (s *graphicsState) stateScissorTest(scissorTest bool) {
	if noStateGuard || s.CommonState.ScissorTest != scissorTest {
		s.CommonState.ScissorTest = scissorTest
		glFeature(gl.SCISSOR_TEST, scissorTest)
	}
}

func (s *graphicsState) stateBlend(blend bool) {
	if noStateGuard || s.CommonState.Blend != blend {
		s.CommonState.Blend = blend
		glFeature(gl.BLEND, blend)
	}
}

func (s *graphicsState) stateProgramPointSizeExt(enabled bool) {
	if noStateGuard || s.CommonState.ProgramPointSizeExt != enabled {
		s.CommonState.ProgramPointSizeExt = enabled
		glFeature(gl.PROGRAM_POINT_SIZE_EXT, enabled)
	}
}

func (s *graphicsState) stateMultisample(multisample bool) {
	if noStateGuard || s.CommonState.Multisample != multisample {
		s.CommonState.Multisample = multisample
		glFeature(gl.MULTISAMPLE, multisample)
	}
}

func (s *graphicsState) stateAlphaToCoverage(gpuInfo *gfx.DeviceInfo, alphaToCoverage bool) {
	if noStateGuard || s.alphaToCoverage != alphaToCoverage {
		s.alphaToCoverage = alphaToCoverage
		if gpuInfo.AlphaToCoverage {
			glFeature(gl.SAMPLE_ALPHA_TO_COVERAGE, alphaToCoverage)
		}
	}
}

func (s *graphicsState) stateDepthClamp(gpuInfo *gfx.DeviceInfo, clamp bool) {
	if noStateGuard || s.State.DepthClamp != clamp {
		s.State.DepthClamp = clamp
		if gpuInfo.DepthClamp {
			glFeature(gl.DEPTH_CLAMP, clamp)
		}
	}
	glFeature(gl.DEPTH_CLAMP, true)
}

func (s *graphicsState) stateFaceCulling(m gfx.FaceCullMode) {
	if noStateGuard || s.State.FaceCulling != m {
		s.State.FaceCulling = m
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

func (s *graphicsState) stateProgram(p uint32) bool {
	if noStateGuard || s.CommonState.ShaderProgram != p {
		s.CommonState.ShaderProgram = p
		gl.UseProgram(p)
		return true
	}
	return false
}
