// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"azul3d.org/gfx.v1"
	"azul3d.org/native/gl.v1"
	"image"
)

// Set this to true to disable state guarding (i.e. avoiding useless OpenGL
// state calls). This is useful for debugging the state guard code.
const noStateGuard = false

func unconvertFaceCull(fc int32) gfx.FaceCullMode {
	switch fc {
	case gl.FRONT:
		return gfx.FrontFaceCulling
	case gl.BACK:
		return gfx.BackFaceCulling
	case gl.FRONT_AND_BACK:
		return gfx.NoFaceCulling
	}
	panic("failed to convert")
}

func convertStencilOp(o gfx.StencilOp) int32 {
	switch o {
	case gfx.SKeep:
		return gl.KEEP
	case gfx.SZero:
		return gl.ZERO
	case gfx.SReplace:
		return gl.REPLACE
	case gfx.SIncr:
		return gl.INCR
	case gfx.SIncrWrap:
		return gl.INCR_WRAP
	case gfx.SDecr:
		return gl.DECR
	case gfx.SDecrWrap:
		return gl.DECR_WRAP
	case gfx.SInvert:
		return gl.INVERT
	}
	panic("failed to convert")
}

func unconvertStencilOp(o int32) gfx.StencilOp {
	switch o {
	case gl.KEEP:
		return gfx.SKeep
	case gl.ZERO:
		return gfx.SZero
	case gl.REPLACE:
		return gfx.SReplace
	case gl.INCR:
		return gfx.SIncr
	case gl.INCR_WRAP:
		return gfx.SIncrWrap
	case gl.DECR:
		return gfx.SDecr
	case gl.DECR_WRAP:
		return gfx.SDecrWrap
	case gl.INVERT:
		return gfx.SInvert
	}
	panic("failed to convert")
}

func convertCmp(c gfx.Cmp) int32 {
	switch c {
	case gfx.Always:
		return gl.ALWAYS
	case gfx.Never:
		return gl.NEVER
	case gfx.Less:
		return gl.LESS
	case gfx.LessOrEqual:
		return gl.LEQUAL
	case gfx.Greater:
		return gl.GREATER
	case gfx.GreaterOrEqual:
		return gl.GEQUAL
	case gfx.Equal:
		return gl.EQUAL
	case gfx.NotEqual:
		return gl.NOTEQUAL
	}
	panic("failed to convert")
}

func unconvertCmp(c int32) gfx.Cmp {
	switch c {
	case gl.ALWAYS:
		return gfx.Always
	case gl.NEVER:
		return gfx.Never
	case gl.LESS:
		return gfx.Less
	case gl.LEQUAL:
		return gfx.LessOrEqual
	case gl.GREATER:
		return gfx.Greater
	case gl.GEQUAL:
		return gfx.GreaterOrEqual
	case gl.EQUAL:
		return gfx.Equal
	case gl.NOTEQUAL:
		return gfx.NotEqual
	}
	panic("failed to convert")
}

func convertBlendOp(o gfx.BlendOp) int32 {
	switch o {
	case gfx.BZero:
		return gl.ZERO
	case gfx.BOne:
		return gl.ONE
	case gfx.BSrcColor:
		return gl.SRC_COLOR
	case gfx.BOneMinusSrcColor:
		return gl.ONE_MINUS_SRC_COLOR
	case gfx.BDstColor:
		return gl.DST_COLOR
	case gfx.BOneMinusDstColor:
		return gl.ONE_MINUS_DST_COLOR
	case gfx.BSrcAlpha:
		return gl.SRC_ALPHA
	case gfx.BOneMinusSrcAlpha:
		return gl.ONE_MINUS_SRC_ALPHA
	case gfx.BDstAlpha:
		return gl.DST_ALPHA
	case gfx.BOneMinusDstAlpha:
		return gl.ONE_MINUS_DST_ALPHA
	case gfx.BConstantColor:
		return gl.CONSTANT_COLOR
	case gfx.BOneMinusConstantColor:
		return gl.ONE_MINUS_CONSTANT_COLOR
	case gfx.BConstantAlpha:
		return gl.CONSTANT_ALPHA
	case gfx.BOneMinusConstantAlpha:
		return gl.ONE_MINUS_CONSTANT_ALPHA
	case gfx.BSrcAlphaSaturate:
		return gl.SRC_ALPHA_SATURATE
	}
	panic("failed to convert")
}

func unconvertBlendOp(o int32) gfx.BlendOp {
	switch o {
	case gl.ZERO:
		return gfx.BZero
	case gl.ONE:
		return gfx.BOne
	case gl.SRC_COLOR:
		return gfx.BSrcColor
	case gl.ONE_MINUS_SRC_COLOR:
		return gfx.BOneMinusSrcColor
	case gl.DST_COLOR:
		return gfx.BDstColor
	case gl.ONE_MINUS_DST_COLOR:
		return gfx.BOneMinusDstColor
	case gl.SRC_ALPHA:
		return gfx.BSrcAlpha
	case gl.ONE_MINUS_SRC_ALPHA:
		return gfx.BOneMinusSrcAlpha
	case gl.DST_ALPHA:
		return gfx.BDstAlpha
	case gl.ONE_MINUS_DST_ALPHA:
		return gfx.BOneMinusDstAlpha
	case gl.CONSTANT_COLOR:
		return gfx.BConstantColor
	case gl.ONE_MINUS_CONSTANT_COLOR:
		return gfx.BOneMinusConstantColor
	case gl.CONSTANT_ALPHA:
		return gfx.BConstantAlpha
	case gl.ONE_MINUS_CONSTANT_ALPHA:
		return gfx.BOneMinusConstantAlpha
	case gl.SRC_ALPHA_SATURATE:
		return gfx.BSrcAlphaSaturate
	}
	panic("failed to convert")
}

func convertBlendEq(eq gfx.BlendEq) int32 {
	switch eq {
	case gfx.BAdd:
		return gl.FUNC_ADD
	case gfx.BSub:
		return gl.FUNC_SUBTRACT
	case gfx.BReverseSub:
		return gl.FUNC_REVERSE_SUBTRACT
	}
	panic("failed to convert")
}

func unconvertBlendEq(eq int32) gfx.BlendEq {
	switch eq {
	case gl.FUNC_ADD:
		return gfx.BAdd
	case gl.FUNC_SUBTRACT:
		return gfx.BSub
	case gl.FUNC_REVERSE_SUBTRACT:
		return gfx.BReverseSub
	}
	panic("failed to convert")
}

func convertRect(rect, bounds image.Rectangle) (x, y int32, width, height uint32) {
	// We must flip the Y axis because image.Rectangle uses top-left as
	// the origin but OpenGL uses bottom-left as the origin.
	y = int32(bounds.Dy() - (rect.Min.Y + rect.Dy())) // bottom
	height = uint32(rect.Dy())                        // top

	x = int32(rect.Min.X)
	width = uint32(rect.Dx())
	return
}

func unconvertRect(bounds image.Rectangle, x, y, width, height int32) (rect image.Rectangle) {
	// We must unflip the Y axis because image.Rectangle uses top-left as
	// the origin but OpenGL uses bottom-left as the origin.
	x0 := int(x)
	x1 := int(x + width)
	y0 := bounds.Dy() - int(y+height)
	y1 := y0 + int(height)
	return image.Rect(x0, y0, x1, y1)
}

var glDefaultStencil = gfx.StencilState{
	WriteMask: 0xFFFF,
	Fail:      gfx.SKeep,
	DepthFail: gfx.SKeep,
	DepthPass: gfx.SKeep,
	Cmp:       gfx.Always,
}

var glDefaultBlend = gfx.BlendState{
	Color:    gfx.Color{R: 0, G: 0, B: 0, A: 0},
	SrcRGB:   gfx.BOne,
	DstRGB:   gfx.BZero,
	SrcAlpha: gfx.BOne,
	DstAlpha: gfx.BZero,
	RGBEq:    gfx.BAdd,
	AlphaEq:  gfx.BAdd,
}

// Please ensure these values match the default OpenGL state values listed in
// the OpenGL documentation.
var defaultGraphicsState = &graphicsState{
	image.Rect(0, 0, 0, 0),                    // scissor - Whole screen
	gfx.Color{R: 0.0, G: 0.0, B: 0.0, A: 0.0}, // clear color
	glDefaultBlend.Color,                      // blend color
	1.0,                                       // clear depth
	0,                                         // clear stencil
	[4]bool{true, true, true, true}, // color write
	gfx.Less,                        // depth func
	glDefaultBlend,                  // blend func seperate
	glDefaultBlend,                  // blend equation seperate
	glDefaultStencil,                // stencil op front
	glDefaultStencil,                // stencil op back
	glDefaultStencil,                // stencil func front
	glDefaultStencil,                // stencil func back
	0xFFFF,                          // stencil mask front
	0xFFFF,                          // stencil mask back
	true,                            // dithering
	false,                           // depth test
	true,                            // depth write
	false,                           // stencil test
	false,                           // blend
	false,                           // alpha to coverage
	gfx.NoFaceCulling,               // face culling
	0,                               // program
}

// Queries the existing OpenGL graphics state and returns it.
func queryExistingState(ctx *gl.Context, gpuInfo *gfx.GPUInfo, bounds image.Rectangle) *graphicsState {
	var (
		GL_BLEND_COLOR int32 = 0x8005 // gl.xml spec file missing this.

		scissor                      [4]int32
		clearColor, blendColor       gfx.Color
		clearDepth                   float64
		clearStencil                 int32
		colorWrite                   [4]uint8
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
		alphaToCoverage uint8
		faceCullMode int32
	)
	ctx.GetIntegerv(gl.SCISSOR_BOX, &scissor[0])
	ctx.GetFloatv(gl.COLOR_CLEAR_VALUE, &clearColor.R)
	ctx.GetFloatv(GL_BLEND_COLOR, &blendColor.R)
	ctx.GetDoublev(gl.DEPTH_CLEAR_VALUE, &clearDepth)
	ctx.GetIntegerv(gl.STENCIL_CLEAR_VALUE, &clearStencil)
	ctx.GetBooleanv(gl.COLOR_WRITEMASK, &colorWrite[0])
	ctx.GetIntegerv(gl.DEPTH_FUNC, &depthFunc)
	ctx.GetIntegerv(gl.BLEND_DST_RGB, &blendDstRGB)
	ctx.GetIntegerv(gl.BLEND_SRC_RGB, &blendSrcRGB)
	ctx.GetIntegerv(gl.BLEND_DST_ALPHA, &blendDstAlpha)
	ctx.GetIntegerv(gl.BLEND_SRC_ALPHA, &blendSrcAlpha)
	ctx.GetIntegerv(gl.BLEND_EQUATION_RGB, &blendEqRGB)
	ctx.GetIntegerv(gl.BLEND_EQUATION_ALPHA, &blendEqAlpha)
	ctx.GetIntegerv(gl.BLEND_EQUATION_RGB, &blendEqRGB)
	ctx.GetIntegerv(gl.BLEND_EQUATION_ALPHA, &blendEqAlpha)

	ctx.GetIntegerv(gl.STENCIL_FAIL, &stencilFrontOpFail)
	ctx.GetIntegerv(gl.STENCIL_PASS_DEPTH_FAIL, &stencilFrontOpDepthFail)
	ctx.GetIntegerv(gl.STENCIL_PASS_DEPTH_PASS, &stencilFrontOpDepthPass)
	ctx.GetIntegerv(gl.STENCIL_WRITEMASK, &stencilFrontWriteMask)
	ctx.GetIntegerv(gl.STENCIL_VALUE_MASK, &stencilFrontReadMask)
	ctx.GetIntegerv(gl.STENCIL_REF, &stencilFrontRef)
	ctx.GetIntegerv(gl.STENCIL_FUNC, &stencilFrontCmp)

	ctx.GetIntegerv(gl.STENCIL_BACK_FAIL, &stencilBackOpFail)
	ctx.GetIntegerv(gl.STENCIL_BACK_PASS_DEPTH_FAIL, &stencilBackOpDepthFail)
	ctx.GetIntegerv(gl.STENCIL_BACK_PASS_DEPTH_PASS, &stencilBackOpDepthPass)
	ctx.GetIntegerv(gl.STENCIL_BACK_WRITEMASK, &stencilBackWriteMask)
	ctx.GetIntegerv(gl.STENCIL_BACK_VALUE_MASK, &stencilBackReadMask)
	ctx.GetIntegerv(gl.STENCIL_BACK_REF, &stencilBackRef)
	ctx.GetIntegerv(gl.STENCIL_BACK_FUNC, &stencilBackCmp)

	ctx.GetBooleanv(gl.DITHER, &dithering)
	ctx.GetBooleanv(gl.DEPTH_TEST, &depthTest)
	ctx.GetBooleanv(gl.DEPTH_WRITEMASK, &depthWrite)
	ctx.GetBooleanv(gl.STENCIL_WRITEMASK, &stencilTest)
	ctx.GetBooleanv(gl.BLEND, &blend)
	if gpuInfo.AlphaToCoverage {
		ctx.GetBooleanv(gl.SAMPLE_ALPHA_TO_COVERAGE, &alphaToCoverage)
	}

	ctx.GetIntegerv(gl.CULL_FACE_MODE, &faceCullMode)
	ctx.Execute()

	return &graphicsState{
		scissor:      unconvertRect(bounds, scissor[0], scissor[1], scissor[2], scissor[3]),
		clearColor:   clearColor,
		blendColor:   blendColor,
		clearDepth:   clearDepth,
		clearStencil: int(clearStencil),
		colorWrite:   [4]bool{gl.Bool(colorWrite[0]), gl.Bool(colorWrite[1]), gl.Bool(colorWrite[2]), gl.Bool(colorWrite[3])},
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
		stencilOpFront: gfx.StencilState{
			Fail:      unconvertStencilOp(stencilFrontOpFail),
			DepthFail: unconvertStencilOp(stencilFrontOpDepthFail),
			DepthPass: unconvertStencilOp(stencilFrontOpDepthPass),
		},
		stencilOpBack: gfx.StencilState{
			Fail:      unconvertStencilOp(stencilBackOpFail),
			DepthFail: unconvertStencilOp(stencilBackOpDepthFail),
			DepthPass: unconvertStencilOp(stencilBackOpDepthPass),
		},
		stencilFuncFront: gfx.StencilState{
			Cmp:       unconvertCmp(stencilFrontCmp),
			Reference: uint(stencilFrontRef),
			ReadMask:  uint(stencilFrontReadMask),
		},
		stencilFuncBack: gfx.StencilState{
			Cmp:       unconvertCmp(stencilBackCmp),
			Reference: uint(stencilBackRef),
			ReadMask:  uint(stencilBackReadMask),
		},
		stencilMaskFront: uint(stencilFrontWriteMask),
		stencilMaskBack:  uint(stencilBackWriteMask),
		dithering:        gl.Bool(dithering),
		depthTest:        gl.Bool(depthTest),
		depthWrite:       gl.Bool(depthWrite),
		stencilTest:      gl.Bool(stencilTest),
		blend:            gl.Bool(blend),
		alphaToCoverage:  gl.Bool(alphaToCoverage),
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
	stencilOpFront, stencilOpBack, stencilFuncFront, stencilFuncBack      gfx.StencilState
	stencilMaskFront, stencilMaskBack                                     uint
	dithering, depthTest, depthWrite, stencilTest, blend, alphaToCoverage bool
	faceCulling                                                           gfx.FaceCullMode
	program                                                               uint32
}

// loads the graphics state, g, making OpenGL calls as neccesarry to components
// that differ between the states s and r.
//
// bounds is the renderer's bounds (e.g. r.Bounds()) to pass into stateScissor().
func (s *graphicsState) load(ctx *gl.Context, gpuInfo *gfx.GPUInfo, bounds image.Rectangle, g *graphicsState) {
	s.stateScissor(ctx, bounds, g.scissor)
	s.stateClearColor(ctx, g.clearColor)
	s.stateBlendColor(ctx, g.blendColor)
	s.stateClearDepth(ctx, g.clearDepth)
	s.stateClearStencil(ctx, g.clearStencil)
	s.stateColorWrite(ctx, g.colorWrite)
	s.stateDepthFunc(ctx, g.depthFunc)
	s.stateBlendFuncSeparate(ctx, g.blendFuncSeparate)
	s.stateBlendEquationSeparate(ctx, g.blendEquationSeparate)
	s.stateStencilOp(ctx, g.stencilOpFront, g.stencilOpBack)
	s.stateStencilFunc(ctx, g.stencilFuncFront, g.stencilFuncBack)
	s.stateStencilMask(ctx, g.stencilMaskFront, g.stencilMaskBack)
	s.stateDithering(ctx, g.dithering)
	s.stateDepthTest(ctx, g.depthTest)
	s.stateDepthWrite(ctx, g.depthWrite)
	s.stateStencilTest(ctx, g.stencilTest)
	s.stateBlend(ctx, g.blend)
	s.stateAlphaToCoverage(ctx, gpuInfo, g.alphaToCoverage)
	s.stateFaceCulling(ctx, g.faceCulling)
	s.stateProgram(ctx, g.program)
}

// bounds is the renderer's bounds (e.g. r.Bounds()).
func (s *graphicsState) stateScissor(ctx *gl.Context, bounds, rect image.Rectangle) {
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
		x, y, width, height := convertRect(rect, bounds)
		ctx.Scissor(x, y, width, height)
	}
}

func (s *graphicsState) stateClearColor(ctx *gl.Context, color gfx.Color) {
	if noStateGuard || s.clearColor != color {
		s.clearColor = color
		ctx.ClearColor(color.R, color.G, color.B, color.A)
	}
}

func (s *graphicsState) stateBlendColor(ctx *gl.Context, c gfx.Color) {
	if noStateGuard || s.blendColor != c {
		s.blendColor = c
		ctx.BlendColor(c.R, c.G, c.B, c.A)
	}
}

func (s *graphicsState) stateClearDepth(ctx *gl.Context, depth float64) {
	if noStateGuard || s.clearDepth != depth {
		s.clearDepth = depth
		ctx.ClearDepth(depth)
	}
}

func (s *graphicsState) stateClearStencil(ctx *gl.Context, stencil int) {
	if noStateGuard || s.clearStencil != stencil {
		s.clearStencil = stencil
		ctx.ClearStencil(int32(stencil))
	}
}

func (s *graphicsState) stateColorWrite(ctx *gl.Context, cw [4]bool) {
	if noStateGuard || s.colorWrite != cw {
		s.colorWrite = cw
		ctx.ColorMask(
			gl.GLBool(cw[0]),
			gl.GLBool(cw[1]),
			gl.GLBool(cw[2]),
			gl.GLBool(cw[3]),
		)
	}
}

func (s *graphicsState) stateDepthFunc(ctx *gl.Context, df gfx.Cmp) {
	if noStateGuard || s.depthFunc != df {
		s.depthFunc = df
		ctx.DepthFunc(convertCmp(df))
	}
}

func (s *graphicsState) stateBlendFuncSeparate(ctx *gl.Context, bs gfx.BlendState) {
	if noStateGuard || s.blendFuncSeparate != bs {
		s.blendFuncSeparate = bs
		ctx.BlendFuncSeparate(
			convertBlendOp(bs.SrcRGB),
			convertBlendOp(bs.DstRGB),
			convertBlendOp(bs.SrcAlpha),
			convertBlendOp(bs.SrcAlpha),
		)
	}
}

func (s *graphicsState) stateBlendEquationSeparate(ctx *gl.Context, bs gfx.BlendState) {
	if noStateGuard || s.blendEquationSeparate != bs {
		s.blendEquationSeparate = bs
		ctx.BlendEquationSeparate(
			convertBlendEq(bs.RGBEq),
			convertBlendEq(bs.AlphaEq),
		)
	}
}

func (s *graphicsState) stateStencilOp(ctx *gl.Context, front, back gfx.StencilState) {
	if noStateGuard || s.stencilOpFront != front || s.stencilOpBack != back {
		s.stencilOpFront = front
		s.stencilOpBack = back
		if front == back {
			// We can save a few calls.
			ctx.StencilOpSeparate(
				gl.FRONT_AND_BACK,
				convertStencilOp(front.Fail),
				convertStencilOp(front.DepthFail),
				convertStencilOp(front.DepthPass),
			)
		} else {
			ctx.StencilOpSeparate(
				gl.FRONT,
				convertStencilOp(front.Fail),
				convertStencilOp(front.DepthFail),
				convertStencilOp(front.DepthPass),
			)
			ctx.StencilOpSeparate(
				gl.BACK,
				convertStencilOp(back.Fail),
				convertStencilOp(back.DepthFail),
				convertStencilOp(back.DepthPass),
			)
		}
	}
}

func (s *graphicsState) stateStencilFunc(ctx *gl.Context, front, back gfx.StencilState) {
	if noStateGuard || s.stencilFuncFront != front || s.stencilFuncBack != back {
		s.stencilFuncFront = front
		s.stencilFuncBack = back
		if front == back {
			// We can save a few calls.
			ctx.StencilFuncSeparate(
				gl.FRONT_AND_BACK,
				convertCmp(front.Cmp),
				int32(front.Reference),
				uint32(front.ReadMask),
			)
		} else {
			ctx.StencilFuncSeparate(
				gl.FRONT,
				convertCmp(front.Cmp),
				int32(front.Reference),
				uint32(front.ReadMask),
			)
			ctx.StencilFuncSeparate(
				gl.BACK,
				convertCmp(back.Cmp),
				int32(back.Reference),
				uint32(back.ReadMask),
			)
		}
	}
}

func (s *graphicsState) stateStencilMask(ctx *gl.Context, front, back uint) {
	if noStateGuard || s.stencilMaskFront != front || s.stencilMaskBack != back {
		s.stencilMaskFront = front
		s.stencilMaskBack = back
		if front == back {
			// We can save a call.
			ctx.StencilMaskSeparate(gl.FRONT_AND_BACK, uint32(front))
		} else {
			ctx.StencilMaskSeparate(gl.FRONT, uint32(front))
			ctx.StencilMaskSeparate(gl.BACK, uint32(back))
		}
	}
}

func (s *graphicsState) stateDithering(ctx *gl.Context, enabled bool) {
	if noStateGuard || s.dithering != enabled {
		s.dithering = enabled
		if enabled {
			ctx.Enable(gl.DITHER)
		} else {
			ctx.Disable(gl.DITHER)
		}
	}
}

func (s *graphicsState) stateDepthTest(ctx *gl.Context, enabled bool) {
	if noStateGuard || s.depthTest != enabled {
		s.depthTest = enabled
		if enabled {
			ctx.Enable(gl.DEPTH_TEST)
		} else {
			ctx.Disable(gl.DEPTH_TEST)
		}
	}
}

func (s *graphicsState) stateDepthWrite(ctx *gl.Context, enabled bool) {
	if noStateGuard || s.depthWrite != enabled {
		s.depthWrite = enabled
		if enabled {
			ctx.DepthMask(gl.GLBool(true))
		} else {
			ctx.DepthMask(gl.GLBool(false))
		}
	}
}

func (s *graphicsState) stateStencilTest(ctx *gl.Context, stencilTest bool) {
	if noStateGuard || s.stencilTest != stencilTest {
		s.stencilTest = stencilTest
		if stencilTest {
			ctx.Enable(gl.STENCIL_TEST)
		} else {
			ctx.Disable(gl.STENCIL_TEST)
		}
	}
}

func (s *graphicsState) stateBlend(ctx *gl.Context, blend bool) {
	if noStateGuard || s.blend != blend {
		s.blend = blend
		if blend {
			ctx.Enable(gl.BLEND)
		} else {
			ctx.Disable(gl.BLEND)
		}
	}
}

func (s *graphicsState) stateAlphaToCoverage(ctx *gl.Context, gpuInfo *gfx.GPUInfo, alphaToCoverage bool) {
	if noStateGuard || s.alphaToCoverage != alphaToCoverage {
		s.alphaToCoverage = alphaToCoverage
		if gpuInfo.AlphaToCoverage {
			if alphaToCoverage {
				ctx.Enable(gl.SAMPLE_ALPHA_TO_COVERAGE)
			} else {
				ctx.Disable(gl.SAMPLE_ALPHA_TO_COVERAGE)
			}
		}
	}
}

func (s *graphicsState) stateFaceCulling(ctx *gl.Context, m gfx.FaceCullMode) {
	if noStateGuard || s.faceCulling != m {
		s.faceCulling = m
		switch m {
		case gfx.BackFaceCulling:
			ctx.Enable(gl.CULL_FACE)
			ctx.CullFace(gl.BACK)
		case gfx.FrontFaceCulling:
			ctx.Enable(gl.CULL_FACE)
			ctx.CullFace(gl.FRONT)
		default:
			ctx.Disable(gl.CULL_FACE)
		}
	}
}

func (s *graphicsState) stateProgram(ctx *gl.Context, p uint32) {
	if noStateGuard || s.program != p {
		s.program = p
		ctx.UseProgram(p)
	}
}
