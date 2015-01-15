// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"image"
	"sync"

	"azul3d.org/lmath.v1"
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
	*Object

	// The projection matrix of the camera, which is responsible for projecting
	// world coordinates into device coordinates.
	Projection Mat4

	fov, near, far float64
}

// SetOrtho sets this camera's Projection matrix to an orthographic one.
//
// The view parameter is the viewing rectangle for the orthographic
// projection in window coordinates.
//
// The near and far parameters describe the minimum closest and maximum
// furthest clipping points of the view frustum.
//
// Clients who need advanced control over how the orthographic viewing frustum
// is set up may use this method's source as a reference (e.g. to change the
// center point, which this method sets at the bottom-left).
func (c *Camera) SetOrtho(view image.Rectangle, near, far float64) {
	w := float64(view.Dx())
	w = float64(int((w / 2.0)) * 2)
	h := float64(view.Dy())
	h = float64(int((h / 2.0)) * 2)
	m := lmath.Mat4Ortho(0, w, 0, h, near, far)
	c.Projection = ConvertMat4(m)

	c.fov = -1
	c.near = near
	c.far = far
}

// SetPersp sets this camera's Projection matrix to an perspective one.
//
// The view parameter is the viewing rectangle for the orthographic
// projection in window coordinates.
//
// The fov parameter is the Y axis field of view (e.g. some games use 75) to
// use.
//
// The near and far parameters describe the minimum closest and maximum
// furthest clipping points of the view frustum.
//
// Clients who need advanced control over how the perspective viewing frustum
// is set up may use this method's source as a reference (e.g. to change the
// center point, which this method sets at the center).
func (c *Camera) SetPersp(view image.Rectangle, fov, near, far float64) {
	aspectRatio := float64(view.Dx()) / float64(view.Dy())
	m := lmath.Mat4Perspective(fov, aspectRatio, near, far)
	c.Projection = ConvertMat4(m)

	c.fov = fov
	c.near = near
	c.far = far
}

// Project returns a 2D point in normalized device space coordinates given a 3D
// point in the world.
//
// If ok=false is returned then the point is outside of the camera's view and
// the returned point may not be meaningful.
func (c *Camera) Project(p3 lmath.Vec3) (p2 lmath.Vec2, ok bool) {
	cameraInv, _ := c.Object.Transform.Mat4().Inverse()
	cameraInv = cameraInv.Mul(zUpRightToYUpRight)

	projection := c.Projection.Mat4()
	vp := cameraInv.Mul(projection)

	p2, ok = vp.Project(p3)
	return
}

// FOV returns the field of view of the camera. For an orthographic camera, this
// will always be -1.
func (c *Camera) FOV() float64 {
	return c.fov
}

// Near returns the minimum closest clipping point of the view frustum.
func (c *Camera) Near() float64 {
	return c.near
}

// Near returns the maximum furthest clipping point of the view frustum.
func (c *Camera) Far() float64 {
	return c.far
}

// Copy returns a new copy of this Camera.
func (c *Camera) Copy() *Camera {
	return &Camera{
		Object:     c.Object.Copy(),
		Projection: c.Projection,
	}
}

// Reset resets this camera to it's default (NewCamera) state.
func (c *Camera) Reset() {
	c.Object.Reset()
	c.Projection = ConvertMat4(lmath.Mat4Identity)
	c.fov = 0.0
	c.near = 0.0
	c.far = 0.0
}

// Destroy destroys this camera for use by other callees to NewCamera. You must
// not use it after calling this method. This makes an implicit call to
// c.Object.Destroy.
func (c *Camera) Destroy() {
	c.Object.Destroy()
	c.Reset()
	camPool.Put(c)
}

var camPool = sync.Pool{
	New: func() interface{} {
		return &Camera{
			NewObject(),
			ConvertMat4(lmath.Mat4Identity),
			0.0, 0.0, 0.0,
		}
	},
}

// NewCamera returns a new *Camera with the default values.
func NewCamera() *Camera {
	return camPool.Get().(*Camera)
}
