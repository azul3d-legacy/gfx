// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

import "image"

func ConvertRect(rect, bounds image.Rectangle) (x, y, width, height int) {
	// We must flip the Y axis because image.Rectangle uses top-left as
	// the origin but OpenGL uses bottom-left as the origin.
	y = bounds.Dy() - (rect.Min.Y + rect.Dy()) // bottom
	height = rect.Dy()                         // top

	x = rect.Min.X
	width = rect.Dx()
	return
}

func UnconvertRect(bounds image.Rectangle, x, y, width, height int) (rect image.Rectangle) {
	// We must unflip the Y axis because image.Rectangle uses top-left as
	// the origin but OpenGL uses bottom-left as the origin.
	x0 := x
	x1 := x + width
	y0 := bounds.Dy() - (y + height)
	y1 := y0 + height
	return image.Rect(x0, y0, x1, y1)
}
