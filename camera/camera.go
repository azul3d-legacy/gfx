// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package camera

import (
	"image"
	"sync"

	"azul3d.org/lmath.v1"
	"azul3d.org/gfx.v2-unstable"
)

var (
	// Get an matrix which will translate our matrix from ZUpRight to YUpRight
	zUpRightToYUpRight = lmath.CoordSysZUpRight.ConvertMat4(lmath.CoordSysYUpRight)
)

// Camera represents a camera object, it may be moved in 3D space using the
// objects transform and the viewing frustum controls how the camera views
// things. Since a camera is in itself also an object it may also have visible
// meshes attatched to it, etc.
//
// A camera and it's methods are not safe for access from multiple goroutines
// concurrently.
type Camera struct {
	*gfx.Object

	// View is the viewing rectangle of the camera, e.g. the rectangle it will
	// draw to on the screen.
	View image.Rectangle

	// Near and far values of the camera's viewing frustum, e.g. Near=0.01,
	// Far=1000.
	Near, Far float64

	// FOV is the Y axis field-of-view (e.g. 75) for the camera's lens. For an
	// orthographic (2D) camera, this field is unused.
	FOV float64

	// Ortho is whether or not the camera is orthographic (2D), if false then
	// it is a projection (3D) camera.
	Ortho bool

	// P is the calculated projection matrix of the camera, as returned by the
	// Projection method.
	P gfx.Mat4

	// Debug causes the camera to attach a wireframe mesh and shader to itself
	// each time Update is called, for debugging purposes.
	Debug bool
}

// Projection implements the gfx.Camera interface.
func (c *Camera) Projection() gfx.Mat4 {
	return c.P
}

// Project returns a 2D point in normalized device space coordinates given a 3D
// point in the world.
//
// If ok=false is returned then the point is outside of the camera's view and
// the returned point may not be meaningful.
func (c *Camera) Project(p3 lmath.Vec3) (p2 lmath.Vec2, ok bool) {
	cameraInv, _ := c.Object.Transform.Mat4().Inverse()
	cameraInv = cameraInv.Mul(zUpRightToYUpRight)

	vp := cameraInv.Mul(c.P.Mat4())

	p2, ok = vp.Project(p3)
	return
}

// Copy returns a new copy of this Camera.
func (c *Camera) Copy() *Camera {
	cpy := *c
	cpy.Object = c.Object.Copy()
	return &cpy
}

// Update updates the camera's projection matrix to account for the given
// viewing rectangle which may have changed (e.g. upon window resize).
func (c *Camera) Update(view image.Rectangle) {
	c.View = view

	if c.Debug {
		c.debugUpdate()
	}

	if c.Ortho {
		// An orthographic camera projection.
		w := float64(c.View.Dx())
		w = float64(int((w / 2.0)) * 2)
		h := float64(c.View.Dy())
		h = float64(int((h / 2.0)) * 2)
		m := lmath.Mat4Ortho(0, w, 0, h, c.Near, c.Far)
		c.P = gfx.ConvertMat4(m)
		return
	}

	// An perspective camera projection.
	aspectRatio := float64(c.View.Dx()) / float64(c.View.Dy())
	m := lmath.Mat4Perspective(c.FOV, aspectRatio, c.Near, c.Far)
	c.P = gfx.ConvertMat4(m)
}

// Destroy destroys this camera for use by other callees to New. You must not
// use it after calling this method. This makes an implicit call to destroy the
// gfx.Object as well (if non-nil).
func (c *Camera) Destroy() {
	if c.Object != nil {
		c.Object.Destroy()
	}
	camPool.Put(c)
}

var camPool = sync.Pool{
	New: func() interface{} {
		return &Camera{
			Object: gfx.NewObject(),
		}
	},
}

// New returns a new perspective (3D) camera updated with the given viewing
// rectangle. The returned camera has the following properties:
//
//   FOV = 75
//   Near = 0.1
//   Far = 1000
//   Ortho = false
//
func New(view image.Rectangle) *Camera {
	c := camPool.Get().(*Camera)
	c.Near = 0.1
	c.Far = 1000
	c.FOV = 75
	c.Ortho = false
	c.Update(view)
	return c
}

// NewOrtho returns a new orthographic (2D) camera updated with the given
// viewing rectangle. The returned camera has the following properties:
//
//   Near = 0.1
//   Far = 1000
//   FOV = 75
//   Ortho = true
//
func NewOrtho(view image.Rectangle) *Camera {
	c := camPool.Get().(*Camera)
	c.Near = 0.1
	c.Far = 1000
	c.FOV = 75
	c.Ortho = true
	c.Update(view)
	return c
}
