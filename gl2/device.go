// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"fmt"
	"image"
	"io"
	"runtime"
	"strings"
	"sync"

	"azul3d.org/gfx.v2-dev"
	"azul3d.org/gfx.v2-dev/clock"
	"azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"
	"azul3d.org/gfx.v2-dev/internal/glc"
	"azul3d.org/gfx.v2-dev/internal/glutil"
	"azul3d.org/gfx.v2-dev/internal/tag"
	"azul3d.org/gfx.v2-dev/internal/util"
)

type pendingQuery struct {
	// The ID of the pending occlusion query.
	id uint32

	// The object of the pending occlusion query.
	o *gfx.Object
}

// renderer implements the Renderer interface defined in the gl2.go file.
type device struct {
	*util.BaseCanvas

	Common glc.Context

	// Render execution channel.
	renderExec chan func() bool

	// The other shared renderer to be used for loading assets, or nil.
	shared struct {
		sync.RWMutex
		*device
	}

	// Whether or not the existing graphics state should be kept between
	// frames. If set to true before rendering a frame the renderer will ask
	// OpenGL for the existing state, the frame will be rendered, and the old
	// OpenGL state restored. This is particularly useful when the renderer
	// must interoperate with other renderers (e.g. QT5).
	keepState bool

	// The graphics clock.
	clock *clock.Clock

	// GPU limitations.
	gpuInfo gfx.DeviceInfo

	// Whether or not certain extensions we use are present or not.
	glArbDebugOutput, glArbMultisample, glArbFramebufferObject,
	glArbOcclusionQuery bool

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
	// touched inside renderExec.
	*rttCanvas

	// Whether or not our global OpenGL state has been set for this frame.
	stateSetForFrame bool

	// Channel to wait for a Render() call to finish.
	renderComplete chan struct{}
}

// Exec implements the Renderer interface.
func (r *device) Exec() chan func() bool {
	return r.renderExec
}

// Clock implements the gfx.Renderer interface.
func (r *device) Clock() *clock.Clock {
	return r.clock
}

// Short methods that just call the hooked methods (hooked methods are used in
// rtt.go file for render to texture things).

// Clear implements the gfx.Canvas interface.
func (r *device) Clear(rect image.Rectangle, bg gfx.Color) {
	r.hookedClear(rect, bg, nil, nil)
}

// ClearDepth implements the gfx.Canvas interface.
func (r *device) ClearDepth(rect image.Rectangle, depth float64) {
	r.hookedClearDepth(rect, depth, nil, nil)
}

// ClearStencil implements the gfx.Canvas interface.
func (r *device) ClearStencil(rect image.Rectangle, stencil int) {
	r.hookedClearStencil(rect, stencil, nil, nil)
}

// Draw implements the gfx.Canvas interface.
func (r *device) Draw(rect image.Rectangle, o *gfx.Object, c *gfx.Camera) {
	r.hookedDraw(rect, o, c, nil, nil)
}

// QueryWait implements the gfx.Canvas interface.
func (r *device) QueryWait() {
	r.hookedQueryWait(nil, nil)
}

// Render implements the gfx.Canvas interface.
func (r *device) Render() {
	r.hookedRender(nil, nil)
}

// Implements gfx.Canvas interface.
func (r *device) hookedClear(rect image.Rectangle, bg gfx.Color, pre, post func()) {
	// Clearing an empty rectangle is effectively no-op.
	if rect.Empty() {
		return
	}
	r.renderExec <- func() bool {
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
func (r *device) hookedClearDepth(rect image.Rectangle, depth float64, pre, post func()) {
	// Clearing an empty rectangle is effectively no-op.
	if rect.Empty() {
		return
	}
	r.renderExec <- func() bool {
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
func (r *device) hookedClearStencil(rect image.Rectangle, stencil int, pre, post func()) {
	// Clearing an empty rectangle is effectively no-op.
	if rect.Empty() {
		return
	}
	r.renderExec <- func() bool {
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

func (r *device) hookedQueryWait(pre, post func()) {
	// Ask the render channel to wait for query results now.
	r.renderExec <- func() bool {
		if pre != nil {
			pre()
		}

		// Flush and execute any pending OpenGL commands.
		gl.Flush()
		//gl.Execute()

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

func (r *device) hookedRender(pre, post func()) {
	// Ask the render channel to render things now.
	r.renderExec <- func() bool {
		// If any finalizers have ran and actually want us to free something,
		// then we perform this operation now.
		r.freeMeshes()
		r.freeShaders()
		r.freeTextures()
		r.freeFBOs()
		r.freeRenderbuffers()

		if pre != nil {
			pre()
		}

		// Execute all pending operations.
		for i := 0; i < len(r.renderExec); i++ {
			f := <-r.renderExec
			f()
		}

		// Flush and execute any pending OpenGL commands.
		gl.Flush()
		//gl.Execute()

		// Wait for occlusion query results to come in.
		r.queryWait()

		if post != nil {
			post()
		}

		if tag.Gfxdebug {
			r.debugRender()
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
func (r *device) queryYield() int {
	if !r.glArbOcclusionQuery {
		return 0
	}
	r.pending.Lock()
	var (
		available, result int32
	)
	for queryIndex, query := range r.pending.queries {
		gl.GetQueryObjectiv(query.id, gl.QUERY_RESULT_AVAILABLE, &available)
		//gl.Execute()
		if available == gl.TRUE {
			// Get the result then.
			gl.GetQueryObjectiv(query.id, gl.QUERY_RESULT, &result)

			// Delete the query.
			gl.DeleteQueries(1, &query.id)
			//gl.Execute()

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
func (r *device) queryWait() {
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

// GPUInfo implements the gfx.Renderer interface.
func (r *device) Info() gfx.DeviceInfo {
	return r.gpuInfo
}

// Effectively just calls stateScissor(), but passes in the proper bounds
// according to whether or not we are rendering to an rttCanvas or not.
func (r *device) performScissor(rect image.Rectangle) {
	if r.rttCanvas != nil {
		r.stateScissor(r.rttCanvas.Bounds(), rect)
	} else {
		r.stateScissor(r.Bounds(), rect)
	}
}

func (r *device) performClear(rect image.Rectangle, bg gfx.Color) {
	r.setGlobalState()

	// Color write mask effects the glClear call below.
	r.stateColorWrite(true, true, true, true)

	// Perform clearing.
	r.performScissor(rect)
	r.stateClearColor(bg)
	gl.Clear(uint32(gl.COLOR_BUFFER_BIT))
}

func (r *device) performClearDepth(rect image.Rectangle, depth float64) {
	r.setGlobalState()

	// Depth write mask effects the glClear call below.
	r.stateDepthWrite(true)

	// Perform clearing.
	r.performScissor(rect)
	r.stateClearDepth(depth)
	gl.Clear(uint32(gl.DEPTH_BUFFER_BIT))
}

func (r *device) performClearStencil(rect image.Rectangle, stencil int) {
	r.setGlobalState()

	// Stencil mask effects the glClear call below.
	r.stateStencilMask(0xFFFF, 0xFFFF)

	// Perform clearing.
	r.performScissor(rect)
	r.stateClearStencil(stencil)
	gl.Clear(uint32(gl.STENCIL_BUFFER_BIT))
}

func (r *device) setGlobalState() {
	if !r.stateSetForFrame {
		r.stateSetForFrame = true

		if r.keepState {
			// We want to maintain state between frames for cooperation with
			// another renderer. Store the existing graphics state now so that
			// we can restore it after the frame is rendered.
			r.prevGraphicsState = queryGraphicsState(r.Common, &r.gpuInfo, r.Bounds())

			// Since the existing state is also not what we think it is, we
			// must update our state now.
			cpy := *r.prevGraphicsState
			r.graphicsState = &cpy
		}

		// Update viewport bounds.
		bounds := r.BaseCanvas.Bounds()
		gl.Viewport(0, 0, int32(bounds.Dx()), int32(bounds.Dy()))

		// Enable scissor testing.
		r.stateScissorTest(true)

		// Enable setting point size in shader programs.
		r.stateProgramPointSizeExt(true)

		// Enable multisampling, if available and wanted.
		if r.glArbMultisample {
			if r.BaseCanvas.MSAA() {
				r.stateMultisample(true)
			}
		}
	}
}

func (r *device) clearGlobalState() {
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
		r.graphicsState.load(&r.gpuInfo, r.Bounds(), oldState)
	}
}

func (r *device) logf(format string, args ...interface{}) {
	// Log the error.
	r.debug.RLock()
	if r.debug.W != nil {
		fmt.Fprintf(r.debug.W, format, args...)
	}
	r.debug.RUnlock()
}

// SetDebugOutput implements the Renderer interface.
func (r *device) SetDebugOutput(w io.Writer) {
	r.debug.RLock()
	r.debug.W = w
	r.debug.RUnlock()
}

// Destroy implements the Renderer interface.
func (r *device) Destroy() {
}

func queryVersion() (major, minor, release int, vendorVersion string) {
	versionString := gl.GoStr(gl.GetString(gl.VERSION))
	return glutil.ParseVersionString(versionString)
}

func queryShaderVersion() (major, minor, release int, vendorVersion string) {
	versionString := gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION))
	return glutil.ParseVersionString(versionString)
}

func queryExtensions() map[string]bool {
	// Initialize extensions map
	var (
		extensions = make(map[string]bool)
		extString  = gl.GoStr(gl.GetString(gl.EXTENSIONS))
	)
	if len(extString) > 0 {
		split := strings.Split(extString, " ")
		for _, ext := range split {
			if len(ext) > 0 {
				extensions[ext] = true
			}
		}
	}
	return extensions
}

func extension(name string, extensions map[string]bool) bool {
	_, ok := extensions[name]
	return ok
}

func glStr(s string) *int8 {
	return gl.Str(s + "\x00")
}

// Initialization of OpenGL in two seperate thread at the same time is racy
// because it is storing information on the OpenGL function pointers.
var initLock sync.Mutex

// newDevice is the implementation of New.
func newDevice(opts ...Option) (Device, error) {
	r := &device{
		BaseCanvas: &util.BaseCanvas{
			VMSAA: true,
		},
		Common:         glc.NewContext(),
		renderExec:     make(chan func() bool, 1024),
		renderComplete: make(chan struct{}, 8),
		wantFree:       make(chan struct{}, 1),
		clock:          clock.New(),
	}
	for _, opt := range opts {
		opt(r)
	}

	// Initialize OpenGL.
	initLock.Lock()
	err := gl.Init()
	if err != nil {
		return nil, fmt.Errorf("OpenGL Error: %v", err)
	}
	initLock.Unlock()

	// Note: we don't need r.gl.Lock() here because no other goroutines
	// can be using r.ctx yet since we haven't returned from New().

	// Find the renderer's precision.
	var redBits, greenBits, blueBits, alphaBits, depthBits, stencilBits int32
	gl.GetIntegerv(gl.RED_BITS, &redBits)
	gl.GetIntegerv(gl.GREEN_BITS, &greenBits)
	gl.GetIntegerv(gl.BLUE_BITS, &blueBits)
	gl.GetIntegerv(gl.ALPHA_BITS, &alphaBits)
	gl.GetIntegerv(gl.DEPTH_BITS, &depthBits)
	gl.GetIntegerv(gl.STENCIL_BITS, &stencilBits)
	//gl.Execute()

	r.BaseCanvas.VPrecision.RedBits = uint8(redBits)
	r.BaseCanvas.VPrecision.GreenBits = uint8(greenBits)
	r.BaseCanvas.VPrecision.BlueBits = uint8(blueBits)
	r.BaseCanvas.VPrecision.AlphaBits = uint8(alphaBits)
	r.BaseCanvas.VPrecision.DepthBits = uint8(depthBits)
	r.BaseCanvas.VPrecision.StencilBits = uint8(stencilBits)

	exts := queryExtensions()
	extsStr := make([]string, len(exts))
	ei := 0
	for s := range exts {
		extsStr[ei] = s
		ei++
	}

	if tag.Gfxdebug {
		r.debugInit(exts)
	}

	// Query whether we have the GL_ARB_framebuffer_object extension.
	r.glArbFramebufferObject = extension("GL_ARB_framebuffer_object", exts)

	// Query whether we have the GL_ARB_occlusion_query extension.
	r.glArbOcclusionQuery = extension("GL_ARB_occlusion_query", exts)

	// Query whether we have the GL_ARB_multisample extension.
	r.glArbMultisample = extension("GL_ARB_multisample", exts)
	if r.glArbMultisample {
		// Query the number of samples and sample buffers we have, if any.
		gl.GetIntegerv(gl.SAMPLES, &r.samples)
		gl.GetIntegerv(gl.SAMPLE_BUFFERS, &r.sampleBuffers)
		//gl.Execute() // Needed because glGetIntegerv must execute now.
		r.BaseCanvas.VPrecision.Samples = int(r.samples)
	}

	// Store GPU info.
	var maxTextureSize, maxVaryingFloats, maxVertexInputs, maxFragmentInputs, occlusionQueryBits int32
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &maxTextureSize)
	gl.GetIntegerv(gl.MAX_VARYING_FLOATS, &maxVaryingFloats)
	gl.GetIntegerv(gl.MAX_VERTEX_UNIFORM_COMPONENTS, &maxVertexInputs)
	gl.GetIntegerv(gl.MAX_FRAGMENT_UNIFORM_COMPONENTS, &maxFragmentInputs)
	if r.glArbOcclusionQuery {
		gl.GetQueryiv(gl.SAMPLES_PASSED, gl.QUERY_COUNTER_BITS, &occlusionQueryBits)
	}
	//gl.Execute()

	// Collect GPU information.
	r.gpuInfo.DepthClamp = extension("GL_ARB_depth_clamp", exts)
	r.gpuInfo.MaxTextureSize = int(maxTextureSize)
	r.gpuInfo.AlphaToCoverage = r.glArbMultisample && r.samples > 0 && r.sampleBuffers > 0
	r.gpuInfo.Name = gl.GoStr(gl.GetString(gl.RENDERER))
	r.gpuInfo.Vendor = gl.GoStr(gl.GetString(gl.VENDOR))
	r.gpuInfo.OcclusionQuery = r.glArbOcclusionQuery && occlusionQueryBits > 0
	r.gpuInfo.OcclusionQueryBits = int(occlusionQueryBits)
	r.gpuInfo.NPOT = extension("GL_ARB_texture_non_power_of_two", exts)
	r.gpuInfo.TexWrapBorderColor = true

	// OpenGL Information.
	glInfo := &gfx.GLInfo{
		Extensions: extsStr,
	}
	glInfo.MajorVersion, glInfo.MinorVersion, _, _ = queryVersion()
	r.gpuInfo.GL = glInfo

	// GLSL information.
	glslInfo := &gfx.GLSLInfo{
		MaxVaryingFloats:  int(maxVaryingFloats),
		MaxVertexInputs:   int(maxVertexInputs),
		MaxFragmentInputs: int(maxFragmentInputs),
	}
	glslInfo.MajorVersion, glslInfo.MinorVersion, _, _ = queryShaderVersion()
	r.gpuInfo.GLSL = glslInfo

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
		gl.GetIntegerv(gl.MAX_SAMPLES, &maxSamples)
		//gl.Execute()
		for i := 0; i < int(maxSamples); i++ {
			fmts.Samples = append(fmts.Samples, i)
		}

		r.gpuInfo.RTTFormats = fmts
	}

	// Grab the current renderer bounds (opengl viewport).
	var viewport [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &viewport[0])
	//gl.Execute()
	r.BaseCanvas.VBounds = image.Rect(0, 0, int(viewport[2]), int(viewport[3]))

	if r.keepState {
		// Load the existing graphics state.
		r.graphicsState = queryGraphicsState(r.Common, &r.gpuInfo, r.BaseCanvas.VBounds)
	} else {
		r.graphicsState = defaultGraphicsState
	}

	// Update scissor rectangle.
	r.stateScissor(r.BaseCanvas.VBounds, r.BaseCanvas.VBounds)

	// Grab the number of texture compression formats.
	var numFormats int32
	gl.GetIntegerv(gl.NUM_COMPRESSED_TEXTURE_FORMATS, &numFormats)
	//gl.Execute() // Needed because glGetIntegerv must execute now.

	// Store the slice of texture compression formats.
	if numFormats > 0 {
		r.compressedTextureFormats = make([]int32, numFormats)
		gl.GetIntegerv(gl.COMPRESSED_TEXTURE_FORMATS, &r.compressedTextureFormats[0])
		//gl.Execute() // Needed because glGetIntegerv must execute now.
	}
	return r, nil
}
