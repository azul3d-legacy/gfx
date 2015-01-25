// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"image"

	"azul3d.org/gfx.v2-unstable"
	"azul3d.org/gfx.v2-unstable/clock"
)

// Swapper is a swappable gfx.Device implementation.
type Swapper struct {
	Yield chan struct{}
	Swap  chan gfx.Device
	clock *clock.Clock
	msaa  bool
	d     gfx.Device
}

// Clock returns this swapper's own clock.
func (s *Swapper) Clock() *clock.Clock {
	return s.clock
}

// Bounds returns the bounds of the current graphics device.
func (s *Swapper) Bounds() image.Rectangle {
	return s.d.Bounds()
}

// Bounds returns the precision of the current graphics device.
func (s *Swapper) Precision() gfx.Precision {
	return s.d.Precision()
}

// Bounds returns the info of the current graphics device.
func (s *Swapper) Info() gfx.DeviceInfo {
	return s.d.Info()
}

// Download performs a download from the current graphics device.
func (s *Swapper) Download(r image.Rectangle, complete chan image.Image) {
	s.d.Download(r, complete)
}

// SetMSAA sets the MSAA status of the current graphics device.
func (s *Swapper) SetMSAA(msaa bool) {
	s.msaa = msaa
	s.d.SetMSAA(msaa)
}

// MSAA gets the MSAA status of the current graphics device.
func (s *Swapper) MSAA() (msaa bool) {
	s.msaa = s.d.MSAA()
	return s.d.MSAA()
}

// Clear submits a clear operation to the current graphics device.
func (s *Swapper) Clear(r image.Rectangle, bg gfx.Color) {
	s.d.Clear(r, bg)
}

// ClearDepth submits a depth clear operation to the current graphics device.
func (s *Swapper) ClearDepth(r image.Rectangle, depth float64) {
	s.d.ClearDepth(r, depth)
}

// ClearStencil submits a stencil clear operation to the current graphics
// device.
func (s *Swapper) ClearStencil(r image.Rectangle, stencil int) {
	s.d.ClearStencil(r, stencil)
}

// Draw submits a draw operation to the current graphics device.
func (s *Swapper) Draw(r image.Rectangle, o *gfx.Object, c gfx.Camera) {
	s.d.Draw(r, o, c)
}

// QueryWait waits for occlusion queries to wait on the current graphics
// device.
func (s *Swapper) QueryWait() {
	s.d.QueryWait()
}

// Render renders a frame using the current graphics device. When it finishes
// the swapper considers yielding and swapping the underlying graphics device
// out with another.
func (s *Swapper) Render() {
	s.d.Render()
	s.clock.Tick()

	select {
	case <-s.Yield:
		s.Swap <- nil
		s.d = <-s.Swap
	default:
	}
}

// LoadMesh loads a mesh using the current graphics device.
func (s *Swapper) LoadMesh(m *gfx.Mesh, done chan *gfx.Mesh) {
	s.d.LoadMesh(m, done)
}

// LoadTexture loads a texture using the current graphics device.
func (s *Swapper) LoadTexture(t *gfx.Texture, done chan *gfx.Texture) {
	s.d.LoadTexture(t, done)
}

// LoadShader loads a shader using the current graphics device.
func (s *Swapper) LoadShader(sh *gfx.Shader, done chan *gfx.Shader) {
	s.d.LoadShader(sh, done)
}

// RenderToTexture returns a new RTT canvas using the current graphics device.
//
// TODO(slimsag): Do we require a swappable canvas, here, too?
func (s *Swapper) RenderToTexture(cfg gfx.RTTConfig) gfx.Canvas {
	return s.d.RenderToTexture(cfg)
}

// NewSwapper returns a new graphics device swapper, wrapping the given device.
func NewSwapper(d gfx.Device) *Swapper {
	s := &Swapper{
		Yield: make(chan struct{}, 1),
		Swap:  make(chan gfx.Device),
		clock: clock.New(),
		d:     d,
	}
	return s
}
