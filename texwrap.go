// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

// TexWrap represents a single way that the extra area of a texture wraps
// around a mesh.
type TexWrap uint8

const (
	// Repeat repeats the extra area of the texture into infinity.
	Repeat TexWrap = iota

	// Clamp clamps the extra area of the texture by stretching the edge pixels
	// out into infinity.
	Clamp

	// BorderColor represents the extra area of the texture by the border color
	// specified on the Texture object.
	//
	// Not all devices support BorderColor (See DeviceInfo.TexWrapBorderColor
	// for more information). To check for support:
	//
	//  if devInfo.TexWrapBorderColor {
	//      // Have BorderColor support.
	//  }
	//
	// Devices fall back to the Clamp TexWrap mode in the event that you
	// attempt to use it and the device does not support it.
	BorderColor

	// Mirror represents the extra area of the texture by mirroring itself into
	// infinity.
	Mirror
)
