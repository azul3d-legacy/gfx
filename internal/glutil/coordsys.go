// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

import "azul3d.org/lmath.v1"

var (
	// CoordSys is a coordinate system matrix converting from a right-handed Z
	// up coordinate system to a right-handed Y up coordinate system.
	CoordSys = lmath.CoordSysZUpRight.ConvertMat4(lmath.CoordSysYUpRight)
)

