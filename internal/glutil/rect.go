// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

import "image"

func ConvertRect(rect, bounds image.Rectangle) (x, y, width, height int32) {
	// We must flip the Y axis because image.Rectangle uses top-left as
	// the origin but OpenGL uses bottom-left as the origin.
	y = int32(bounds.Dy() - (rect.Min.Y + rect.Dy())) // bottom
	height = int32(rect.Dy())                         // top

	x = int32(rect.Min.X)
	width = int32(rect.Dx())
	return
}

func UnconvertRect(bounds image.Rectangle, x, y, width, height int32) (rect image.Rectangle) {
	// We must unflip the Y axis because image.Rectangle uses top-left as
	// the origin but OpenGL uses bottom-left as the origin.
	x0 := int(x)
	x1 := int(x + width)
	y0 := bounds.Dy() - int(y+height)
	y1 := y0 + int(height)
	return image.Rect(x0, y0, x1, y1)
}

