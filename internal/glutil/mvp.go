// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

import (
	"azul3d.org/gfx.v2-unstable"
	"azul3d.org/gfx.v2-unstable/camera"
	"azul3d.org/lmath.v1"
)

// MVPCache caches a gfx.Object's Model, View, Projection and MVP matrices.
//
// Each one is precalculated such that they can be fed directly into shaders
// without a recalculation each frame.
type MVPCache struct {
	// LastTransform is the object's last-known transform. If the object's
	// current transform and this one are not identical, the matrices must be
	// recalculated.
	lastTransform lmath.Mat4

	// The last-known camera transform and projection.
	lastCameraTransform lmath.Mat4
	lastProjection      lmath.Mat4

	// The cached pre-calculated matrices to feed directly into shaders.
	Model, View, Projection, MVP gfx.Mat4
}

// needUpdate tells if the cached matrices need to be updated to account for
// the given object and camera.
func (m *MVPCache) needUpdate(o *gfx.Object, c gfx.Camera) bool {
	if o.Transform.Mat4() != m.lastTransform {
		return true
	}
	if m.camMat(c) != m.lastCameraTransform {
		return true
	}
	if m.camProj(c) != m.lastProjection {
		return true
	}
	return false
}

// Update updates the cache (if needed) to account for changes to the object's
// transform, or the camera's transform/projection changing. It should be
// called before drawing the object each frame.
func (m *MVPCache) Update(o *gfx.Object, c gfx.Camera) {
	if !m.needUpdate(o, c) {
		return
	}

	objMat := o.Transform.Mat4()
	m.lastTransform = objMat

	// The "Model" matrix is the object's transformation matrix, completely
	// untouched.
	m.Model = gfx.ConvertMat4(objMat)

	// The "View" matrix is the coordinate system conversion, multiplied
	// against the camera object's transformation matrix.
	m.lastCameraTransform = m.camMat(c)
	view := CoordSys
	if c != nil {
		// Apply inverse of camera object transformation.
		camInverse, _ := m.lastCameraTransform.Inverse()
		view = camInverse.Mul(view)
	}
	m.View = gfx.ConvertMat4(view)

	// The "Projection" matrix is the camera's projection matrix, completely
	// untouched.
	m.lastProjection = m.camProj(c)
	m.Projection = gfx.ConvertMat4(m.lastProjection)

	// The "MVP" matrix is Model * View * Projection matrix.
	mvp := objMat
	mvp = mvp.Mul(view)
	mvp = mvp.Mul(m.lastProjection)
	m.MVP = gfx.ConvertMat4(mvp)
}

// camMat returns the camera's transformation matrix, or the identity matrix if
// the camera is nil.
func (m *MVPCache) camMat(c gfx.Camera) lmath.Mat4 {
	if c != nil {
		// TODO(slimsag): incorrect type assumption!
		return c.(*camera.Camera).Object.Transform.Mat4()
	}
	return lmath.Mat4Identity
}

// camProj returns the camera's projection matrix, or the identity matrix if
// the camera is nil.
func (m *MVPCache) camProj(c gfx.Camera) lmath.Mat4 {
	if c != nil {
		return c.Projection().Mat4()
	}
	return lmath.Mat4Identity
}
