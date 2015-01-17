// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"image"

	"azul3d.org/gfx.v2-unstable/clock"
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
// A canvases method's are safe to call from multiple goroutines concurrently.
type Canvas interface {
	Downloadable

	// SetMSAA should request that this canvas use multi-sample anti-aliasing
	// during drawing. By default MSAA is enabled.
	//
	// Even if MSAA is requested to be enabled, there is no guarantee that it
	// will actually be used. For instance if the device does not support it.
	SetMSAA(enabled bool)

	// MSAA returns the last value passed into SetMSAA on this canvas.
	MSAA() bool

	// Precision should return the precision of the canvas's color, depth, and
	// stencil buffers.
	Precision() Precision

	// Bounds should return the bounding rectangle of this canvas, any and all
	// methods of this canvas that take rectangles as parameters will be
	// clamped to these bounds.
	//
	// The bounds that this method returns may change over time (e.g. when a
	// user resizes the window).
	Bounds() image.Rectangle

	// Clear submits a clear operation to the canvas. It will clear the given
	// rectangle of the canvas's color buffer to the specified background
	// color.
	//
	// If the rectangle is empty this function is no-op.
	Clear(r image.Rectangle, bg Color)

	// ClearDepth submits a depth-clear operation to the canvas. It will clear
	// the given rectangle of the canvas's depth buffer to the specified depth
	// value (in the range of 0.0 to 1.0, where 1.0 is furthest away).
	//
	// If the rectangle is empty this function is no-op.
	ClearDepth(r image.Rectangle, depth float64)

	// ClearStencil submits a stencil-clear operation to the canvas. It will
	// clear the given rectangle of the canvas's stencil buffer to the
	// specified stencil value.
	//
	// If the rectangle is empty this function is no-op.
	ClearStencil(r image.Rectangle, stencil int)

	// Draw submits a draw operation to the canvas. It will draw the given
	// graphics object onto the specified rectangle of the canvas.
	//
	// Upon calling this method, ownership of the graphics object is given to
	// the canvas itself and you may no longer access it safely until the
	// canvas gives ownership of the object back to you.
	//
	// The canvas returns owernship of the object via the channel, if done !=
	// nil, or more effectively after Render has returned.
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
	//
	//  o.State == nil
	//  o.Shader == nil
	//  len(o.Shader.Error) > 0
	//  len(o.Meshes) == 0
	//  !o.Meshes[N].Loaded && len(o.Meshes[N].Vertices) == 0
	//  !o.Textures[n].Loaded && o.Textures[N].Source == nil
	//
	// If the rectangle is empty this function is no-op.
	Draw(r image.Rectangle, o *Object, c *Camera)

	// QueryWait blocks until all pending draw object's occlusion queries
	// completely finish. Most clients should avoid this call as it can easilly
	// cause graphics pipeline stalls if not handled with care.
	//
	// Instead of calling QueryWait immediately for conditional drawing, it is
	// common practice to instead make use of the previous frame's occlusion
	// query results as this allows the CPU and GPU to work in parralel instead
	// of being directly synchronized with one another.
	//
	// If the GPU does not support occlusion queries (see DeviceInfo's
	// OcclusionQuery field) then this function is no-op.
	QueryWait()

	// Render should finalize all pending clear and draw operations as if they
	// where all submitted over a single channel like so:
	//
	//  pending := len(ops)
	//  for i := 0; i < pending; i++ {
	//      op := <-ops
	//      finalize(op)
	//  }
	//
	// and once complete the final frame should be sent to the graphics
	// hardware for rasterization.
	//
	// Additionally, a call to Render() means an implicit call to QueryWait().
	Render()
}

// DeviceInfo describes general information and limitations of the graphics
// device, such as the maximum texture size and other features which may or may
// not be supported by the graphics device.
type DeviceInfo struct {
	// GL is a pointer to information about the OpenGL implementation, if the
	// device is a OpenGL device. Otherwise it is nil.
	GL *GLInfo

	// GLSL is a pointer to information about the GLSL implementation, if the
	// device is a OpenGL device. Otherwise it is nil.
	GLSL *GLSLInfo

	// MaxTextureSize is the maximum size of either X or Y dimension of texture
	// images for use with the device, or -1 if not available.
	MaxTextureSize int

	// Whether or not the AlphaToCoverage alpha mode is supported (if false
	// then BinaryAlpha will automatically be used as a fallback).
	AlphaToCoverage bool

	// Whether or not rendering objects with the DepthClamp state enabled is
	// supported.
	DepthClamp bool

	// Whether or not occlusion queries are supported or not.
	OcclusionQuery bool

	// The number of bits reserved for the sample count when performing
	// occlusion queries, if the number goes above what this many bits could
	// store then it is generally (but not always) clamped to that value.
	//
	// Some devices (e.g. OpenGL ES ones without certain extensions) only
	// support boolean occlusion queries (i.e. you can only tell if some
	// samples passed, but not how many specifically).
	OcclusionQueryBits int

	// The name of the graphics hardware, or an empty string if not available.
	// For example it may look something like:
	//
	//  Mesa DRI Intel(R) Sandybridge Mobile
	//
	Name string

	// The vendor name of the graphics hardware, or an empty string if not
	// available. For example:
	//
	//  Intel Open Source Technology Center
	//
	Vendor string

	// Whether or not the graphics hardware supports Non Power Of Two texture
	// sizes.
	//
	// If true, then textures may be any arbitrary size (while keeping in mind
	// this often incurs a performance cost, and does not work well with
	// compression or mipmapping).
	//
	// If false, then texture dimensions must be a power of two (e.g. 32x64,
	// 512x512, etc) or else the texture will be resized by the device to the
	// nearest power-of-two.
	NPOT bool

	// The formats available for render-to-texture (RTT).
	RTTFormats

	// Whether or not the graphics hardware supports the use of the BorderColor
	// TexWrap mode. If the hardware doesn't support it the device falls back
	// to the Clamp TexWrap mode in it's place.
	//
	// (Mobile) OpenGL ES 2 never supports BorderColor.
	//
	// (Desktop) OpenGL 2 always supports BorderColor.
	TexWrapBorderColor bool
}

// Device represents a graphics device and is capable of loading meshes,
// textures, and shaders. A device itself has a base canvas which can be drawn
// to (typically a window on the screen, for instance).
//
// A devices method's are safe to call from multiple goroutines concurrently.
type Device interface {
	// The base canvas of the device (typically the window on the screen). The
	// canvas, like all other canvases, must not be accessed from multiple
	// goroutines concurrently.
	Canvas

	// Clock should return the graphics clock object which monitors the time
	// between frames, etc. The device is responsible for ticking it every
	// time a frame is rendered.
	Clock() *clock.Clock

	// Info should return information about the graphics hardware.
	Info() DeviceInfo

	// LoadMesh should begin loading the specified mesh asynchronously.
	//
	// Additionally, the device will set m.Loaded to true, and then invoke
	// m.ClearData(), thus allowing the data slices to be garbage collected).
	//
	// When the mesh is fully loaded, it is sent to the done channel if != nil,
	// and as long as sending would not block (i.e. ensure a buffer size of at
	// least one).
	//
	// Upon calling this method, ownership of the mesh is transferred to the
	// device itself and you may no longer access it safely until the device
	// passes ownership back to you over the done channel.
	LoadMesh(m *Mesh, done chan *Mesh)

	// LoadTexture should begin loading the specified texture asynchronously.
	//
	// Additionally, the device will set t.Loaded to true, and then invoke
	// t.ClearData(), thus allowing the source image to be garbage collected.
	//
	// When the texture is fully loaded, it is sent to the done channel
	// if != nil, and as long as sending would not block (i.e. ensure a buffer
	// size of at least one).
	//
	// Upon calling this method, ownership of the mesh is transferred to the
	// device itself and you may no longer access it safely until the device
	// passes ownership back to you over the done channel.
	LoadTexture(t *Texture, done chan *Texture)

	// LoadShader should begin loading the specified shader asynchronously.
	//
	// Additionally, if the shader was successfully loaded (no error log was
	// written) then the device will set s.Loaded to true, and then invoke
	// s.ClearData(), thus allowing the source data slices to be garbage
	// collected.
	//
	// When the shader is fully loaded (or has an error), it is sent to the
	// done channel if != nil, and as long as sending would not block (i.e.
	// ensure a buffer size of at least one).
	//
	// Upon calling this method, ownership of the shader is transferred to the
	// device itself and you may no longer access it safely until the device
	// passes ownership back to you over the done channel.
	LoadShader(s *Shader, done chan *Shader)

	// RenderToTexture creates and returns a canvas that when drawn to, stores
	// the results into one or multiple of the three textures (Color, Depth,
	// Stencil) of the given configuration.
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
