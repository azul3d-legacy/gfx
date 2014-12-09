// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import "fmt"

// AlphaMode describes a single alpha transparency mode that can be used for
// drawing transparent parts of objects.
type AlphaMode uint8

// String returns a string representation of this alpha transparency mode.
// e.g. NoAlpha -> "NoAlpha"
func (m AlphaMode) String() string {
	switch m {
	case NoAlpha:
		return "NoAlpha"
	case AlphaBlend:
		return "AlphaBlend"
	case BinaryAlpha:
		return "BinaryAlpha"
	case AlphaToCoverage:
		return "AlphaToCoverage"
	}
	return fmt.Sprintf("AlphaMode(%d)", m)
}

const (
	// NoAlpha means the object should be drawn without transparency. Parts
	// of the object that would normally be drawn as transparent will be drawn
	// as an opaque black color instead.
	NoAlpha AlphaMode = iota

	// AlphaBlend means the object should be drawn with alpha-blended
	// transparency. This type of transparency works well for most (but not
	// all) cases.
	//
	// Pros:
	//     Pixels can be semi-transparent.
	//
	// Cons:
	//     Draw order dependant. Opaque objects mut be drawn before
	//     transparent ones due to the way alpha blending works.
	//
	//     Does not work well with self-occluding transparent objects (e.g. a
	//     cube where all faces are semi-transparent) because individual faces
	//     would have to be sorted for correct order -- which is not feasible
	//     in realtime applications.
	//
	AlphaBlend

	// BinaryAlpha means the object should be drawn with binary transparency,
	// this causes transparency to be thought of as a 'binary' decision, where
	// each pixel is either fully transparent or opaque.
	//
	// Pixels with an alpha value of less than 0.5 are considered fully
	// transparent (invisible), and likewise pixels with an alpha value of
	// greater than or equal to 0.5 are considered fully opaque (solid,
	// non-transparent).
	//
	// Pros:
	//     Draw order independent. Regardless of the order objects are drawn
	//     the result will look the same (unlike AlphaBlend).
	//
	// Cons:
	//     Jagged-looking edges because pixels may not be semi-transparent.
	//
	BinaryAlpha

	// AlphaToCoverage means the object should be drawn using alpha-to-coverage
	// with special multisample bits.
	//
	// Pros:
	//     Draw order independent. Regardless of the order objects are drawn
	//     the result will look the same (unlike AlphaBlend).
	//
	//     No jagged-looking edges, pixels may be semi-transparent (unlike
	//     BinaryAlpha).
	//
	// Cons:
	//     Only some newer hardware supports it (in the event that hardware
	//     does not support it the fallback used is BinaryAlpha because it also
	//     does not suffer from draw ordering issues, although it does cause
	//     jagged-looking edges).
	AlphaToCoverage
)
