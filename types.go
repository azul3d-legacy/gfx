// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"azul3d.org/lmath.v1"
)

// TexCoord represents a 2D texture coordinate with U and V components.
type TexCoord struct {
	U, V float32
}

// Mat4 represents a 32-bit floating point 4x4 matrix for compatability with
// graphics hardware.
// math.Mat4 should be used anywhere that an explicit 32-bit type is not
// needed.
type Mat4 [4][4]float32

// Mat4 converts this 32-bit Mat4 to a 64-bit math.Mat4 matrix.
func (m Mat4) Mat4() math.Mat4 {
	return math.Mat4{
		[4]float64{float64(m[0][0]), float64(m[0][1]), float64(m[0][2]), float64(m[0][3])},
		[4]float64{float64(m[1][0]), float64(m[1][1]), float64(m[1][2]), float64(m[1][3])},
		[4]float64{float64(m[2][0]), float64(m[2][1]), float64(m[2][2]), float64(m[2][3])},
		[4]float64{float64(m[3][0]), float64(m[3][1]), float64(m[3][2]), float64(m[3][3])},
	}
}

// ConvertMat4 converts the 64-bit math.Mat4 to a 32-bit Mat4 matrix.
func ConvertMat4(m math.Mat4) Mat4 {
	return Mat4{
		[4]float32{float32(m[0][0]), float32(m[0][1]), float32(m[0][2]), float32(m[0][3])},
		[4]float32{float32(m[1][0]), float32(m[1][1]), float32(m[1][2]), float32(m[1][3])},
		[4]float32{float32(m[2][0]), float32(m[2][1]), float32(m[2][2]), float32(m[2][3])},
		[4]float32{float32(m[3][0]), float32(m[3][1]), float32(m[3][2]), float32(m[3][3])},
	}
}

// Vec3 represents a 32-bit floating point three-component vector for
// compatability with graphics hardware.
// math.Vec3 should be used anywhere that an explicit 32-bit type is not
// needed.
type Vec3 struct {
	X, Y, Z float32
}

// Vec3 converts this 32-bit Vec3 to a 64-bit math.Vec3 vector.
func (v Vec3) Vec3() math.Vec3 {
	return math.Vec3{float64(v.X), float64(v.Y), float64(v.Z)}
}

// ConvertVec3 converts the 64-bit math.Vec3 to a 32-bit Vec3 vector.
func ConvertVec3(v math.Vec3) Vec3 {
	return Vec3{float32(v.X), float32(v.Y), float32(v.Z)}
}
