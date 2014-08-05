// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"azul3d.org/clock.v1"
	"azul3d.org/gfx.v1"
	"azul3d.org/native/gl.v1"
	"errors"
	"fmt"
	"image"
	"io"
	"runtime"
	"sync"
)

// Used when attempting to create an OpenGL 2.0 renderer in a lesser OpenGL context.
var ErrInvalidVersion = errors.New("invalid OpenGL version; must be at least OpenGL 2.0")

type pendingQuery struct {
	// The ID of the pending occlusion query.
	id uint32

	// The object of the pending occlusion query.
	o *gfx.Object
}

// Renderer is an OpenGL 2 based graphics renderer, it runs independant of the
// window management library being used (GLFW, SDL, Chippy, QML, etc).
//
// The renderer primarily uses two independant OpenGL contexts, one is used for
// rendering and one is used for managing resources like meshes, textures, and
// shaders which allows for asynchronous loading (although it is also possible
// to use only a single OpenGL context for windowing libraries that do not
// support multiple, but this will inheritly disable asynchronous loading).
type Renderer struct {
	*baseCanvas

	// Render and loader execution channels.
	RenderExec chan func() bool
	LoaderExec chan func()

	// Whether or not the existing graphics state should be kept between
	// frames. If set to true before rendering a frame the renderer will ask
	// OpenGL for the existing state, the frame will be rendered, and the old
	// OpenGL state restored. This is particularly useful when the renderer
	// must interoperate with other renderers (e.g. QT5).
	keepState bool

	// render/loader context API.
	render, loader *gl.Context

	// The graphics clock.
	clock *clock.Clock

	// GPU limitations.
	gpuInfo gfx.GPUInfo

	// Whether or not certain extensions we use are present or not.
	glArbMultisample, glArbFramebufferObject, glArbOcclusionQuery bool

	// Number of multisampling samples, buffers.
	samples, sampleBuffers int32

	// List of OpenGL texture compression format identifiers.
	compressedTextureFormats []int32

	// A channel which will have one empty struct inside it in the event that
	// a finalizer for a mesh, texture, etc has ran and something needs to be
	// free'd.
	wantFree chan struct{}

	// List of native meshes to free at next frame.
	meshesToFree struct {
		sync.RWMutex
		slice []*nativeMesh
	}

	// List of native shaders to free at next frame.
	shadersToFree struct {
		sync.RWMutex
		slice []*nativeShader
	}

	// List of native texture id's to free at next frame.
	texturesToFree struct {
		sync.RWMutex
		slice []uint32
	}

	// List of native FBO id's to free at next frame.
	fbosToFree struct {
		sync.RWMutex
		slice []uint32
	}

	// List of native render buffer id's to free at next frame.
	renderbuffersToFree struct {
		sync.RWMutex
		slice []uint32
	}

	*graphicsState
	prevGraphicsState *graphicsState
	lastShader        *gfx.Shader

	// Structure used to manage the debug output stream.
	debug struct {
		sync.RWMutex
		W io.Writer
	}

	// Structure used to manage pending occlusion queries.
	pending struct {
		sync.Mutex
		queries []pendingQuery
	}

	// RTT format lookups (from gfx formats to GL ones).
	rttTexFormats map[gfx.TexFormat]int32
	rttDSFormats  map[gfx.DSFormat]int32

	// If non-nil, then we are currently rendering to a texture. It is only
	// touched inside RenderExec.
	*rttCanvas

	// Whether or not our global OpenGL state has been set for this frame.
	stateSetForFrame bool

	// Channel to wait for a Render() call to finish.
	renderComplete chan struct{}
}

// Implements gfx.Renderer interface.
func (r *Renderer) Clock() *clock.Clock {
	return r.clock
}

// Short methods that just call the hooked methods (hooked methods are used in
// rtt.go file for render to texture things).

// Implements gfx.Canvas interface.
func (r *Renderer) Clear(rect image.Rectangle, bg gfx.Color) {
	r.hookedClear(rect, bg, nil, nil)
}

// Implements gfx.Canvas interface.
func (r *Renderer) ClearDepth(rect image.Rectangle, depth float64) {
	r.hookedClearDepth(rect, depth, nil, nil)
}

// Implements gfx.Canvas interface.
func (r *Renderer) ClearStencil(rect image.Rectangle, stencil int) {
	r.hookedClearStencil(rect, stencil, nil, nil)
}

// Implements gfx.Canvas interface.
func (r *Renderer) Draw(rect image.Rectangle, o *gfx.Object, c *gfx.Camera) {
	r.hookedDraw(rect, o, c, nil, nil)
}

// Implements gfx.Canvas interface.
func (r *Renderer) QueryWait() {
	r.hookedQueryWait(nil, nil)
}

// Implements gfx.Canvas interface.
func (r *Renderer) Render() {
	r.hookedRender(nil, nil)
}

// Implements gfx.Canvas interface.
func (r *Renderer) hookedClear(rect image.Rectangle, bg gfx.Color, pre, post func()) {
	r.RenderExec <- func() bool {
		if pre != nil {
			pre()
		}
		r.performClear(rect, bg)
		r.queryYield()
		if post != nil {
			post()
		}
		return false
	}
}

// Implements gfx.Canvas interface.
func (r *Renderer) hookedClearDepth(rect image.Rectangle, depth float64, pre, post func()) {
	r.RenderExec <- func() bool {
		if pre != nil {
			pre()
		}
		r.performClearDepth(rect, depth)
		r.queryYield()
		if post != nil {
			post()
		}
		return false
	}
}

// Implements gfx.Canvas interface.
func (r *Renderer) hookedClearStencil(rect image.Rectangle, stencil int, pre, post func()) {
	r.RenderExec <- func() bool {
		if pre != nil {
			pre()
		}
		r.performClearStencil(rect, stencil)
		r.queryYield()
		if post != nil {
			post()
		}
		return false
	}
}

func (r *Renderer) hookedQueryWait(pre, post func()) {
	// Ask the render channel to wait for query results now.
	r.RenderExec <- func() bool {
		if pre != nil {
			pre()
		}

		// Flush and execute any pending OpenGL commands.
		r.render.Flush()
		r.render.Execute()

		// Wait for occlusion query results to come in.
		r.queryWait()

		if post != nil {
			post()
		}

		// signal render completion.
		r.renderComplete <- struct{}{}
		return false
	}
	<-r.renderComplete
}

func (r *Renderer) hookedRender(pre, post func()) {
	// If any finalizers have ran and actually want us to free something, then
	// we will ask the loader to do so now.
	r.LoaderExec <- func() {
		r.freeMeshes()
		r.freeShaders()
		r.freeTextures()
		r.freeFBOs()
		r.freeRenderbuffers()
	}

	// Ask the render channel to render things now.
	r.RenderExec <- func() bool {
		if pre != nil {
			pre()
		}

		// Execute all pending operations. 
		for i := 0; i < len(r.RenderExec); i++ {
			f := <-r.RenderExec
			f()
		}

		// Flush and execute any pending OpenGL commands.
		r.render.Flush()
		r.render.Execute()

		// Wait for occlusion query results to come in.
		r.queryWait()

		if post != nil {
			post()
		}

		if r.rttCanvas != nil {
			// We are rendering to a texture. We do not need to clear global
			// state, tick the clock, or return true (frame rendered).

			// We do still need to signal render completion.
			r.renderComplete <- struct{}{}
			return false
		}

		// Clear our OpenGL state now.
		r.clearGlobalState()

		// Tick the clock.
		r.clock.Tick()

		// signal render completion.
		r.renderComplete <- struct{}{}
		return true
	}
	<-r.renderComplete
}

// Tries to receive pending occlusion query results, returns immediately if
// none are available yet. Returns the number of queries still pending.
func (r *Renderer) queryYield() int {
	if !r.glArbOcclusionQuery {
		return 0
	}
	r.pending.Lock()
	var (
		available, result int32
	)
	for queryIndex, query := range r.pending.queries {
		r.render.GetQueryObjectiv(query.id, gl.QUERY_RESULT_AVAILABLE, &available)
		r.render.Execute()
		if available == gl.TRUE {
			// Get the result then.
			r.render.GetQueryObjectiv(query.id, gl.QUERY_RESULT, &result)

			// Delete the query.
			r.render.DeleteQueries(1, &query.id)
			r.render.Execute()

			// Update object's sample count.
			nativeObj := query.o.NativeObject.(nativeObject)
			nativeObj.sampleCount = int(result)
			query.o.NativeObject = nativeObj

			// Remove from pending slice.
			r.pending.queries = append(r.pending.queries[:queryIndex], r.pending.queries[queryIndex+1:]...)
		}
	}
	length := len(r.pending.queries)
	r.pending.Unlock()
	return length
}

// Blocks until all pending occlusion query results are received.
func (r *Renderer) queryWait() {
	if !r.glArbOcclusionQuery {
		return
	}

	// We have no choice except to busy-wait until the results come: OpenGL
	// doesn't provide a blocking mechanism for waiting for query results but
	// at least we can runtime.Gosched() other goroutines.
	for i := 0; r.queryYield() > 0; i++ {
		// Only runtime.Gosched() every 16th iteration to avoid bogging down
		// rendering.
		if i != 0 && (i%16) == 0 {
			runtime.Gosched()
		}
	}
}

// Implements gfx.Renderer interface.
func (r *Renderer) GPUInfo() gfx.GPUInfo {
	return r.gpuInfo
}

// Effectively just calls stateScissor(), but passes in the proper bounds
// according to whether or not we are rendering to an rttCanvas or not.
func (r *Renderer) performScissor(ctx *gl.Context, rect image.Rectangle) {
	if r.rttCanvas != nil {
		r.stateScissor(ctx, r.rttCanvas.Bounds(), rect)
	} else {
		r.stateScissor(ctx, r.Bounds(), rect)
	}
}

func (r *Renderer) performClear(rect image.Rectangle, bg gfx.Color) {
	r.setGlobalState()

	// Color write mask effects the glClear call below.
	r.stateColorWrite(r.render, [4]bool{true, true, true, true})

	// Perform clearing.
	r.performScissor(r.render, rect)
	r.stateClearColor(r.render, bg)
	r.render.Clear(uint32(gl.COLOR_BUFFER_BIT))
}

func (r *Renderer) performClearDepth(rect image.Rectangle, depth float64) {
	r.setGlobalState()

	// Depth write mask effects the glClear call below.
	r.stateDepthWrite(r.render, true)

	// Perform clearing.
	r.performScissor(r.render, rect)
	r.stateClearDepth(r.render, depth)
	r.render.Clear(uint32(gl.DEPTH_BUFFER_BIT))
}

func (r *Renderer) performClearStencil(rect image.Rectangle, stencil int) {
	r.setGlobalState()

	// Stencil mask effects the glClear call below.
	r.stateStencilMask(r.render, 0xFFFF, 0xFFFF)

	// Perform clearing.
	r.performScissor(r.render, rect)
	r.stateClearStencil(r.render, stencil)
	r.render.Clear(uint32(gl.STENCIL_BUFFER_BIT))
}

// UpdateBounds updates the effective bounding rectangle of this renderer. It
// must be called whenever the OpenGL canvas size should change (e.g. on window
// resize).
func (r *Renderer) UpdateBounds(bounds image.Rectangle) {
	r.baseCanvas.setBounds(bounds)
}

func (r *Renderer) setGlobalState() {
	if !r.stateSetForFrame {
		r.stateSetForFrame = true

		if r.keepState {
			// We want to maintain state between frames for cooperation with
			// another renderer. Store the existing graphics state now so that
			// we can restore it after the frame is rendered.
			r.prevGraphicsState = queryExistingState(r.render, &r.gpuInfo, r.Bounds())

			// Since the existing state is also not what we think it is, we
			// must update our state now.
			cpy := *r.prevGraphicsState
			r.graphicsState = &cpy
		}

		// Update viewport bounds.
		bounds := r.baseCanvas.Bounds()
		r.render.Viewport(0, 0, uint32(bounds.Dx()), uint32(bounds.Dy()))

		// Enable scissor testing.
		r.render.Enable(gl.SCISSOR_TEST)

		// Enable multisampling, if available and wanted.
		if r.glArbMultisample {
			if r.baseCanvas.MSAA() {
				r.render.Enable(gl.MULTISAMPLE)
			}
		}
	}
}

func (r *Renderer) clearGlobalState() {
	if r.stateSetForFrame {
		r.stateSetForFrame = false

		// Clear last used state.
		oldState := defaultGraphicsState
		if r.keepState {
			// We want to maintain state between frames for cooperation with
			// another renderer. Use the graphics state that was active when
			// the frame started then.
			oldState = r.prevGraphicsState
		}
		r.graphicsState.load(r.render, &r.gpuInfo, r.Bounds(), oldState)

		// Reset last shader so that uniforms are loaded again next frame.
		r.lastShader = nil

		// Disable scissor testing.
		r.render.Disable(gl.SCISSOR_TEST)

		// Disable multisampling, if available.
		if r.glArbMultisample {
			r.render.Disable(gl.MULTISAMPLE)
		}
	}
}

func (r *Renderer) logf(format string, args ...interface{}) {
	// Log the error.
	r.debug.RLock()
	if r.debug.W != nil {
		fmt.Fprintf(r.debug.W, format, args...)
	}
	r.debug.RUnlock()
}

// SetDebugOutput sets the writer, w, to write debug output to. It will mostly
// contain just shader debug information, but other information may be written
// in the future as well.
func (r *Renderer) SetDebugOutput(w io.Writer) {
	r.debug.RLock()
	r.debug.W = w
	r.debug.RUnlock()
}

// New returns a new OpenGL 2 based graphics renderer. If any error is returned
// then a nil renderer is also returned. This function must be called only when
// an OpenGL 2 context is active.
//
// keepState specifies whether or not the existing graphics state should be
// maintained between frames. If set to true then before rendering a frame the
// graphics state will be saved, the frame rendered, and the old graphics state
// restored again. This is particularly useful when the renderer must cooperate
// with another renderer (e.g. QT5). Do not turn it on needlessly though as it
// does come with a performance cost.
func New(keepState bool) (*Renderer, error) {
	r := &Renderer{
		baseCanvas: &baseCanvas{
			msaa: true,
		},
		RenderExec:     make(chan func() bool, 1024),
		LoaderExec:     make(chan func(), 1024),
		keepState:      keepState,
		renderComplete: make(chan struct{}, 8),
		wantFree:       make(chan struct{}, 1),
		clock:          clock.New(),
	}

	// Initialize r.render now.
	r.render = gl.New()
	r.render.SetBatching(false)
	if !r.render.AtLeastVersion(2, 0) {
		return nil, ErrInvalidVersion
	}

	// Initialize r.loader now.
	r.loader = gl.New()
	r.loader.SetBatching(false)

	// Note: we don't need r.ctx.Lock() here because no other goroutines
	// can be using r.ctx yet since we haven't returned from New().

	// Find the renderer's precision.
	var redBits, greenBits, blueBits, alphaBits, depthBits, stencilBits int32
	r.render.GetIntegerv(gl.RED_BITS, &redBits)
	r.render.GetIntegerv(gl.GREEN_BITS, &greenBits)
	r.render.GetIntegerv(gl.BLUE_BITS, &blueBits)
	r.render.GetIntegerv(gl.ALPHA_BITS, &alphaBits)
	r.render.GetIntegerv(gl.DEPTH_BITS, &depthBits)
	r.render.GetIntegerv(gl.STENCIL_BITS, &stencilBits)
	r.render.Execute()

	r.precision.RedBits = uint8(redBits)
	r.precision.GreenBits = uint8(greenBits)
	r.precision.BlueBits = uint8(blueBits)
	r.precision.AlphaBits = uint8(alphaBits)
	r.precision.DepthBits = uint8(depthBits)
	r.precision.StencilBits = uint8(stencilBits)

	// Query whether we have the GL_ARB_framebuffer_object extension.
	r.glArbFramebufferObject = r.render.Extension("GL_ARB_framebuffer_object")

	// Query whether we have the GL_ARB_occlusion_query extension.
	r.glArbOcclusionQuery = r.render.Extension("GL_ARB_occlusion_query")

	// Query whether we have the GL_ARB_multisample extension.
	r.glArbMultisample = r.render.Extension("GL_ARB_multisample")
	if r.glArbMultisample {
		// Query the number of samples and sample buffers we have, if any.
		r.render.GetIntegerv(gl.SAMPLES, &r.samples)
		r.render.GetIntegerv(gl.SAMPLE_BUFFERS, &r.sampleBuffers)
		r.render.Execute() // Needed because glGetIntegerv must execute now.
		r.precision.Samples = int(r.samples)
	}

	// Store GPU info.
	var maxTextureSize, maxVaryingFloats, maxVertexInputs, maxFragmentInputs, occlusionQueryBits int32
	r.render.GetIntegerv(gl.MAX_TEXTURE_SIZE, &maxTextureSize)
	r.render.GetIntegerv(gl.MAX_VARYING_FLOATS, &maxVaryingFloats)
	r.render.GetIntegerv(gl.MAX_VERTEX_UNIFORM_COMPONENTS, &maxVertexInputs)
	r.render.GetIntegerv(gl.MAX_FRAGMENT_UNIFORM_COMPONENTS, &maxFragmentInputs)
	if r.glArbOcclusionQuery {
		r.render.GetQueryiv(gl.SAMPLES_PASSED, gl.QUERY_COUNTER_BITS, &occlusionQueryBits)
	}
	r.render.Execute()

	// Collect GPU information.
	r.gpuInfo.MaxTextureSize = int(maxTextureSize)
	r.gpuInfo.GLSLMaxVaryingFloats = int(maxVaryingFloats)
	r.gpuInfo.GLSLMaxVertexInputs = int(maxVertexInputs)
	r.gpuInfo.GLSLMaxFragmentInputs = int(maxFragmentInputs)
	r.gpuInfo.GLExtensions = r.render.Extensions()
	r.gpuInfo.AlphaToCoverage = r.glArbMultisample && r.samples > 0 && r.sampleBuffers > 0
	r.gpuInfo.Name = gl.String(r.render.GetString(gl.RENDERER))
	r.gpuInfo.Vendor = gl.String(r.render.GetString(gl.VENDOR))
	r.gpuInfo.GLMajor, r.gpuInfo.GLMinor, _ = r.render.Version()
	r.gpuInfo.GLSLMajor, r.gpuInfo.GLSLMinor, _ = r.render.ShaderVersion()
	r.gpuInfo.OcclusionQuery = r.glArbOcclusionQuery && occlusionQueryBits > 0
	r.gpuInfo.OcclusionQueryBits = int(occlusionQueryBits)
	r.gpuInfo.NPOT = r.render.Extension("GL_ARB_texture_non_power_of_two")
	if r.glArbFramebufferObject {
		// See http://www.opengl.org/wiki/Image_Format for more information.
		//
		// TODO:
		//  GL_DEPTH32F_STENCIL8 and GL_DEPTH_COMPONENT32F via Texture.Format
		//      option. (does it require an extension check with GL 2.0?)
		//  GL_STENCIL_INDEX8 (looks like 4.3+ GL hardware)
		//  GL_RGBA16F, GL_RGBA32F via Texture.Format
		//  Compressed formats (DXT ?)
		//  sRGB formats
		//
		//  GL_RGB16, GL_RGBA16

		r.rttTexFormats = make(map[gfx.TexFormat]int32, 16)
		r.rttDSFormats = make(map[gfx.DSFormat]int32, 16)

		// Formats below are guaranteed to be supported in OpenGL 2.x hardware:
		fmts := r.gpuInfo.RTTFormats

		// Color formats.
		fmts.ColorFormats = append(fmts.ColorFormats, []gfx.TexFormat{
			gfx.RGB,
			gfx.RGBA,
		}...)
		for _, cf := range fmts.ColorFormats {
			r.rttTexFormats[cf] = convertTexFormat(cf)
		}

		// Depth formats.
		fmts.DepthFormats = append(fmts.DepthFormats, []gfx.DSFormat{
			gfx.Depth16,
			gfx.Depth24,
			gfx.Depth32,
			gfx.Depth24AndStencil8,
		}...)
		r.rttDSFormats[gfx.Depth16] = gl.DEPTH_COMPONENT16
		r.rttDSFormats[gfx.Depth24] = gl.DEPTH_COMPONENT24
		r.rttDSFormats[gfx.Depth32] = gl.DEPTH_COMPONENT32

		// Stencil formats.
		fmts.StencilFormats = append(fmts.StencilFormats, []gfx.DSFormat{
			gfx.Depth24AndStencil8,
		}...)
		r.rttDSFormats[gfx.Depth24AndStencil8] = gl.DEPTH24_STENCIL8

		// Sample counts.
		// TODO: Beware integer texture formats -- MSAA can at max be
		//       GL_MAX_INTEGER_SAMPLES with those.
		var maxSamples int32
		r.render.GetIntegerv(gl.MAX_SAMPLES, &maxSamples)
		r.render.Execute()
		for i := 0; i < int(maxSamples); i++ {
			fmts.Samples = append(fmts.Samples, i)
		}

		r.gpuInfo.RTTFormats = fmts
	}

	// Grab the current renderer bounds (opengl viewport).
	var viewport [4]int32
	r.render.GetIntegerv(gl.VIEWPORT, &viewport[0])
	r.render.Execute()
	r.baseCanvas.bounds = image.Rect(0, 0, int(viewport[2]), int(viewport[3]))

	if keepState {
		// Load the existing graphics state.
		r.graphicsState = queryExistingState(r.render, &r.gpuInfo, r.baseCanvas.bounds)
	} else {
		r.graphicsState = defaultGraphicsState
	}

	// Update scissor rectangle.
	r.stateScissor(r.render, r.baseCanvas.bounds, r.baseCanvas.bounds)

	// Grab the number of texture compression formats.
	var numFormats int32
	r.render.GetIntegerv(gl.NUM_COMPRESSED_TEXTURE_FORMATS, &numFormats)
	r.render.Execute() // Needed because glGetIntegerv must execute now.

	// Store the slice of texture compression formats.
	if numFormats > 0 {
		r.compressedTextureFormats = make([]int32, numFormats)
		r.render.GetIntegerv(gl.COMPRESSED_TEXTURE_FORMATS, &r.compressedTextureFormats[0])
		r.render.Execute() // Needed because glGetIntegerv must execute now.
	}
	return r, nil
}
