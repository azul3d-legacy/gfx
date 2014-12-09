// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

// StencilState represents the state to use when the stencil test occurs for a
// front or back facing pixel of an object during drawing. If written in Go
// it would look something like:
//
//  if (s.Reference & s.ReadMask) s.Cmp (stencilValue & s.ReadMask) {
//      if depthTestFailed {
//          stencilValue = s.DepthFail() & s.WriteMask
//      } else {
//          stencilValue = s.DepthPass() & s.WriteMask
//      }
//  } else {
//      stencilValue = s.Fail() & s.WriteMask
//  }
//
type StencilState struct {
	// A mask that will be AND'd with each pixel to be written to the stencil
	// buffer, e.g. 0xFFFF would allow writing to the full range of every pixel
	// in the stencil buffer when drawing the object.
	WriteMask uint

	// A mask that will be AND'd with each pixel to be read/compared to the
	// existing value in the stencil buffer, e.g. 0xFFFF would disable the use
	// of the mask altogether.
	ReadMask uint

	// The reference value that will be used to compare existing values in the
	// stencil buffer against, e.g. if s.Reference == 2 and if s.Func ==
	// GreaterOrEqual, then any value below 2 would not be affected.
	Reference uint

	// Fail specifies what stencil operation should occur when the stencil test
	// fails.
	//
	// Any predefined StencilOp constant is accepted.
	Fail StencilOp

	// DepthFail specifies what stencil operation should occur when the stencil
	// test passes but the depth test fails.
	//
	// Any predefined StencilOp constant is accepted.
	DepthFail StencilOp

	// DepthPass specifies what stencil operation should occur when the stencil
	// test passes and the depth test passes.
	//
	// Any predefined StencilOp constant is accepted.
	DepthPass StencilOp

	// Cmp specifies the comparison operator to use when comparing stencil data
	// with existing data in the stencil buffer.
	//
	// Any predefined Cmp constant is accepted.
	Cmp Cmp
}

// Compare compares this state against the other one using DefaultStencilState
// as a reference when inequality occurs and returns whether or not this state
// should sort before the other one for purposes of state sorting.
func (s StencilState) Compare(other StencilState) bool {
	if s == other {
		return true
	}
	if s.WriteMask != other.WriteMask {
		return s.WriteMask == DefaultStencilState.WriteMask
	}
	if s.ReadMask != other.ReadMask {
		return s.ReadMask == DefaultStencilState.ReadMask
	}
	if s.Reference != other.Reference {
		return s.Reference == DefaultStencilState.Reference
	}
	if s.Fail != other.Fail {
		return s.Fail == DefaultStencilState.Fail
	}
	if s.DepthFail != other.DepthFail {
		return s.DepthFail == DefaultStencilState.DepthFail
	}
	if s.DepthPass != other.DepthPass {
		return s.DepthPass == DefaultStencilState.DepthPass
	}
	if s.Cmp != other.Cmp {
		return s.Cmp == DefaultStencilState.Cmp
	}
	return true
}

// The default stencil state that should be used for graphics objects.
var DefaultStencilState = StencilState{
	WriteMask: 0xFFFF,
	Fail:      SKeep,
	DepthFail: SKeep,
	DepthPass: SKeep,
	Cmp:       Always,
}

// StencilOp represents a single stencil operation to occur when the stencil
// function passes, like SKeep, SReplace, etc.
type StencilOp uint8

const (
	// SKeep keeps the existing stencil data.
	SKeep StencilOp = iota

	// SZero sets the stencil data to zero.
	SZero

	// SReplace replaces the existing stencil data with the stencil reference
	// value.
	SReplace

	// SIncr increments the stencil value by one and clamps the result.
	SIncr

	// SIncrWrap increments the stencil value by 1 and wraps the result if
	// necessary.
	SIncrWrap

	// SDecr decrements the stencil value by one and clamps the result.
	SDecr

	// SDecrWrap decrements the stencil value by 1 and wraps the result if
	// necessary.
	SDecrWrap

	// SInvert inverts the stencil data.
	SInvert
)
