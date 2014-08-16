// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"image"
	"sync"

	"azul3d.org/clock.v1"
)

type nilNativeObject struct{}

func (n nilNativeObject) Destroy() {}
func (n nilNativeObject) SampleCount() int {
	return 0
}

type nilNativeTexture struct {
	format TexFormat
}

func (n nilNativeTexture) Destroy() {}
func (n nilNativeTexture) Download(r image.Rectangle, complete chan image.Image) {
	complete <- nil
}
func (n nilNativeTexture) ChosenFormat() TexFormat {
	return n.format
}

type nilNativeMesh struct{}

func (n nilNativeMesh) Destroy() {}

type nilNativeShader struct{}

func (n nilNativeShader) Destroy() {}

type nilRenderer struct {
	// The MSAA state.
	msaa struct {
		sync.RWMutex
		enabled bool
	}

	precision Precision

	// The graphics clock.
	clock *clock.Clock
}

func (n *nilRenderer) Clock() *clock.Clock {
	return n.clock
}

func (n *nilRenderer) Bounds() image.Rectangle {
	return image.Rect(0, 0, 640, 480)
}

func (n *nilRenderer) Precision() Precision {
	return n.precision
}

func (n *nilRenderer) GPUInfo() GPUInfo {
	return GPUInfo{
		MaxTextureSize:  8096,
		AlphaToCoverage: true,
		OcclusionQuery:  false,
	}
}
func (n *nilRenderer) Download(r image.Rectangle, complete chan image.Image) {
	complete <- nil
}
func (n *nilRenderer) SetMSAA(msaa bool) {
	n.msaa.Lock()
	n.msaa.enabled = msaa
	n.msaa.Unlock()
}
func (n *nilRenderer) MSAA() (msaa bool) {
	n.msaa.RLock()
	msaa = n.msaa.enabled
	n.msaa.RUnlock()
	return
}
func (n *nilRenderer) Clear(r image.Rectangle, bg Color)           {}
func (n *nilRenderer) ClearDepth(r image.Rectangle, depth float64) {}
func (n *nilRenderer) ClearStencil(r image.Rectangle, stencil int) {}
func (n *nilRenderer) Draw(r image.Rectangle, o *Object, c *Camera) {
	o.Bounds()
	o.Lock()
	o.NativeObject = nilNativeObject{}
	o.Unlock()
}
func (n *nilRenderer) QueryWait() {}
func (n *nilRenderer) Render() {
	n.clock.Tick()
}

func (n *nilRenderer) LoadMesh(m *Mesh, done chan *Mesh) {
	m.Lock()
	m.Loaded = true
	m.ClearData()
	m.NativeMesh = nilNativeMesh{}
	m.Unlock()
	select {
	case done <- m:
	default:
	}
}
func (n *nilRenderer) LoadTexture(t *Texture, done chan *Texture) {
	t.Lock()
	t.Loaded = true
	t.ClearData()
	t.NativeTexture = nilNativeTexture{
		t.Format,
	}
	t.Unlock()
	select {
	case done <- t:
	default:
	}
}
func (n *nilRenderer) LoadShader(s *Shader, done chan *Shader) {
	s.Lock()
	s.Loaded = true
	s.ClearData()
	s.NativeShader = nilNativeShader{}
	s.Unlock()
	select {
	case done <- s:
	default:
	}
}

func (n *nilRenderer) RenderToTexture(cfg RTTConfig) Canvas {
	return nil
}

// Nil returns a renderer that does not actually render anything.
func Nil() Renderer {
	r := new(nilRenderer)
	r.precision = Precision{
		RedBits:     255,
		GreenBits:   255,
		BlueBits:    255,
		AlphaBits:   255,
		DepthBits:   255,
		StencilBits: 255,
	}
	r.msaa.enabled = true
	r.clock = clock.New()
	return r
}
