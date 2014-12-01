// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"image"
	"sync"

	"azul3d.org/gfx.v2-dev"
)

// BaseCanvas implements basic portions of a gfx.Canvas. Specifically the
// methods it implements are:
//
//  SetMSAA
//  MSAA
//  Precision
//  Bounds
//
type BaseCanvas struct {
	sync.RWMutex
	VMSAA      bool            // The MSAA state.
	VPrecision gfx.Precision   // The precision of this canvas.
	VBounds    image.Rectangle // The bounding rectangle of this canvas.
}

// Implements gfx.Canvas interface.
func (c *BaseCanvas) SetMSAA(msaa bool) {
	c.Lock()
	c.VMSAA = msaa
	c.Unlock()
}

// Implements gfx.Canvas interface.
func (c *BaseCanvas) MSAA() bool {
	c.RLock()
	msaa := c.VMSAA
	c.RUnlock()
	return msaa
}

// Implements gfx.Canvas interface.
func (c *BaseCanvas) Precision() gfx.Precision {
	c.RLock()
	precision := c.VPrecision
	c.RUnlock()
	return precision
}

// Implements the gfx.Canvas interface.
func (c *BaseCanvas) Bounds() image.Rectangle {
	c.RLock()
	bounds := c.VBounds
	c.RUnlock()
	return bounds
}

// UpdateBounds updates the bounds of this canvas, under c's write lock.
func (c *BaseCanvas) UpdateBounds(bounds image.Rectangle) {
	c.Lock()
	c.VBounds = bounds
	c.Unlock()
}
