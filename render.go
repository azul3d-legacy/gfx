// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"image"

	"azul3d.org/clock.v1"
)

// Precision represents the precision in bits of the color, depth, and stencil
// buffers as well as the number of samples per pixel.
type Precision struct {
	// The precision in bits of each pixel in the color buffer, per color (e.g.
	// 8/8/8/8 would be 32bpp RGBA color, 8/8/8/0 would be 24bpp RGB color, and
	// so on).
	RedBits, GreenBits, BlueBits, AlphaBits uint8

	// The precision in bits of each pixel in the depth buffer (e.g. 8, 16, 24,
	// etc).
	DepthBits uint8

	// The precision in bits of each pixel in the stencil buffer (e.g. 8, 16,
	// 24, etc).
	StencilBits uint8

	// The number of samples available per pixel (e.g. the number of MSAA
	// samples).
	Samples int
}

// Canvas defines a canvas that can be drawn to (i.e. a window that the user
// will visibly see, or a texture that will store the results for later use).
//
// All methods must be safe to call from multiple goroutines.
type Canvas interface {
	Downloadable

	// SetMSAA should request that this canvas use multi-sample anti-aliasing
	// during rendering. By default MSAA is enabled.
	//
	// Even if MSAA is requested to be enabled, there is no guarantee that it
	// will actually be used. For instance if the graphics hardware or
	// rendering API does not support it.
	SetMSAA(enabled bool)

	// MSAA returns the last value passed into SetMSAA on this renderer.
	MSAA() bool

	// Precision should return the precision of the canvas's color, depth, and
	// stencil buffers.
	Precision() Precision

	// Bounds should return the bounding rectangle of this canvas, any and all
	// methods of this canvas that take rectangles as parameters will be
	// clamped to these bounds.
	// The bounds returned by this method may change at any given time (e.g.
	// when a user resizes the window).
	Bounds() image.Rectangle

	// Clear submits a clear operation to the renderer. It will clear the given
	// rectangle of the canvas's color buffer to the specified background
	// color.
	//
	// If the rectangle is empty the entire canvas is cleared.
	Clear(r image.Rectangle, bg Color)

	// ClearDepth submits a depth-clear operation to the renderer. It will
	// clear the given rectangle of the canvas's depth buffer to the specified
	// depth value (in the range of 0.0 to 1.0, where 1.0 is furthest away).
	//
	// If the rectangle is empty the entire canvas is cleared.
	ClearDepth(r image.Rectangle, depth float64)

	// ClearStencil submits a stencil-clear operation to the renderer. It will
	// clear the given rectangle of the canvas's stencil buffer to the
	// specified stencil value.
	//
	// If the rectangle is empty the entire canvas is cleared.
	ClearStencil(r image.Rectangle, stencil int)

	// Draw submits a draw operation to the renderer. It will draw the given
	// graphics object onto the specified rectangle of the canvas.
	//
	// The canvas will lock the object and camera object and they may stay
	// locked until some point in the future when the draw operation completes.
	//
	// If not nil, then the object is drawn according to how it is seen by the
	// given camera object (taking into account the camera object's
	// transformation and projection matrices).
	//
	// If the GPU supports occlusion queries (see GPUInfo.OcclusionQuery) and
	// o.OcclusionTest is set to true then at some point in the future (or when
	// QueryWait() is called) the native object will record the number of
	// samples that passed depth and stencil testing phases such that when
	// SampleCount() is called it will return the number of samples last drawn
	// by the object.
	//
	// The canvas must invoke o.Bounds() some time before clearing data slices
	// of loaded meshes, such that the object has a chance to determine it's
	// bounding box.
	//
	// The object will not be drawn if any of the following cases are true:
	//  o.Shader == nil
	//  len(o.Shader.Error) > 0
	//  len(o.Meshes) == 0
	//  !o.Meshes[N].Loaded && len(o.Meshes[N].Vertices) == 0
	//
	// If the rectangle is empty the entire canvas is drawn to.
	Draw(r image.Rectangle, o *Object, c *Camera)

	// QueryWait blocks until all pending draw object's occlusion queries
	// completely finish. Most clients should avoid this call as it can easilly
	// cause graphics pipeline stalls if not handled with care.
	//
	// Instead of calling QueryWait immediately for conditional rendering, it is
	// common practice to instead make use of the previous frame's occlusion
	// query results as this allows the CPU and GPU to work in parralel instead
	// of being directly synchronized with one another.
	//
	// If the GPU does not support occlusion queries (see
	// GPUInfo.OcclusionQuery) then this function is no-op.
	QueryWait()

	// Render should finalize all pending clear and draw operations as if they
	// where all submitted over a single channel like so:
	//  pending := len(ops)
	//  for i := 0; i < pending; i++ {
	//      op := <-ops
	//      finalize(op)
	//  }
	// and once complete the final frame should be sent to the graphics
	// hardware for rasterization.
	//
	// Additionally, a call to Render() means an implicit call to QueryWait().
	Render()
}

// GPUInfo describes general information and limitations of the graphics
// hardware, such as the maximum texture size and other features which may or
// may not be supported by the graphics hardware.
type GPUInfo struct {
	// MaxTextureSize is the maximum size of either X or Y dimension of texture
	// images for use with the renderer, or -1 if not available.
	MaxTextureSize int

	// Whether or not the AlphaToCoverage alpha mode is supported (if false
	// then BinaryAlpha will automatically be used as a fallback).
	AlphaToCoverage bool

	// Whether or not occlusion queries are supported or not.
	OcclusionQuery bool

	// The number of bits reserved for the sample count when performing
	// occlusion queries, if the number goes above what this many bits could
	// store then it is generally (but not always) clamped to that value.
	//
	// Some renderers (e.g. OpenGL ES with certain extensions) only support
	// boolean occlusion queries (i.e. you can only tell if some samples
	// passed, but not how many specifically).
	OcclusionQueryBits int

	// The name of the graphics hardware, or an empty string if not available.
	// For example it may look something like:
	//  Mesa DRI Intel(R) Sandybridge Mobile
	Name string

	// The vendor name of the graphics hardware, or an empty string if not
	// available. For example:
	//  Intel Open Source Technology Center
	Vendor string

	// Whether or not the graphics hardware supports Non Power Of Two texture
	// sizes.
	//
	// If true, then textures may be any arbitrary size (while keeping in mind
	// this often incurs a performance cost, and does not work well with
	// compression or mipmapping).
	//
	// If false, then texture dimensions must be a power of two (e.g. 32x64,
	// 512x512, etc) or else the texture will be resized by the renderer to the
	// nearest power-of-two.
	NPOT bool

	// The formats available for render-to-texture (RTT).
	RTTFormats

	// Major and minor versions of the OpenGL version in use, or -1 if not
	// available. For example:
	//  3, 0 (for OpenGL 3.0)
	GLMajor, GLMinor int

	// A read-only slice of OpenGL extension strings, empty if not available.
	GLExtensions []string

	// Major and minor versions of the OpenGL Shading Language version in use,
	// or -1 if not available. For example:
	//  1, 30 (for GLSL 1.30)
	GLSLMajor, GLSLMinor int

	// The maximum number of floating-point variables available for varying
	// variables inside GLSL programs, or -1 if not available. Generally at
	// least 32.
	GLSLMaxVaryingFloats int

	// The maximum number of shader inputs (i.e. floating-point values, where a
	// 4x4 matrix is 16 floating-point values) that can be used inside a GLSL
	// vertex shader, or -1 if not available. Generally at least 512.
	GLSLMaxVertexInputs int

	// The maximum number of shader inputs (i.e. floating-point values, where a
	// 4x4 matrix is 16 floating-point values) that can be used inside a GLSL
	// fragment shader, or -1 if not available. Generally at least 64.
	GLSLMaxFragmentInputs int

	// Whether or not the graphics hardware supports the use of the
	// BorderColor TexWrap mode. If the hardware doesn't support it the
	// renderer falls back to the Clamp TexWrap mode in it's place.
	//
	// (Mobile) OpenGL ES 2 never supports BorderColor.
	//
	// (Desktop) OpenGL 2 always supports BorderColor.
	TexWrapBorderColor bool
}

// Renderer is capable of loading meshes, textures, and shaders. A renderer can
// be drawn to as it implements the Canvas interface, and can also be used to
// All methods must be safe to call from multiple goroutines.
type Renderer interface {
	Canvas

	// Clock should return the graphics clock object which monitors the time
	// between frames, etc. The renderer is responsible for ticking it every
	// time a frame is rendered.
	Clock() *clock.Clock

	// GPUInfo should return information about the graphics hardware.
	GPUInfo() GPUInfo

	// LoadMesh should begin loading the specified mesh asynchronously.
	//
	// Additionally, the renderer will set m.Loaded to true, and then invoke
	// m.ClearData(), thus allowing the data slices to be garbage collected).
	//
	// The renderer will lock the mesh and it may stay locked until sometime in
	// the future when the load operation completes. The mesh will be sent over
	// the done channel once the load operation has completed if the channel is
	// not nil and sending would not block.
	LoadMesh(m *Mesh, done chan *Mesh)

	// LoadTexture should begin loading the specified texture asynchronously.
	//
	// Additionally, the renderer will set t.Loaded to true, and then invoke
	// t.ClearData(), thus allowing the source image to be garbage collected.
	//
	// The renderer will lock the texture and it may stay locked until sometime
	// in the future when the load operation completes. The texture will be
	// sent over the done channel once the load operation has completed if the
	// channel is not nil and sending would not block.
	LoadTexture(t *Texture, done chan *Texture)

	// LoadShader should begin loading the specified shader asynchronously.
	//
	// Additionally, if the shader was successfully loaded (no error log was
	// written) then the renderer will set s.Loaded to true, and then invoke
	// s.ClearData(), thus allowing the source data slices to be garbage
	// collected.
	//
	// The renderer will lock the shader and it may stay locked until sometime
	// in the future when the load operation completes. The shader will be sent
	// over the done channel once the load operation has completed if the
	// channel is not nil and sending would not block.
	LoadShader(s *Shader, done chan *Shader)

	// RenderToTexture creates and returns a canvas that when rendered to,
	// stores the results into one or multiple of the three textures (Color,
	// Depth, Stencil) of the given configuration.
	//
	// If the any of the configuration's formats are not supported by the
	// graphics hardware (i.e. not in GPUInfo.RTTFormats), then nil is
	// returned.
	//
	// If the given configuration is not valid (see the cfg.Valid method) then
	// a panic will occur.
	//
	// Any non-nil texture in the configuration will be set to loaded, will
	// have ClearData() called on it, and will have it's bounds set to
	// cfg.Bounds.
	RenderToTexture(cfg RTTConfig) Canvas
}
