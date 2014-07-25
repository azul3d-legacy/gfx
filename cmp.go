// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

// Cmp represents a single comparison operator, like Less, Never, etc.
type Cmp uint8

const (
	// Always is like Go's 'true', for example:
	//  if true {
	//      ...
	//  }
	Always Cmp = iota

	// Never is like Go's 'false', for example:
	//  if false {
	//      ...
	//  }
	Never

	// Less is like Go's '<', for example:
	//  if a < b {
	//      ...
	//  }
	Less

	// LessOrEqual is like Go's '<=', for example:
	//  if a <= b {
	//      ...
	//  }
	LessOrEqual

	// Greater is like Go's '>', for example:
	//  if a > b {
	//      ...
	//  }
	Greater

	// GreaterOrEqual is like Go's '>=', for example:
	//  if a >= b {
	//      ...
	//  }
	GreaterOrEqual

	// Equal is like Go's '==', for example:
	//  if a == b {
	//      ...
	//  }
	Equal

	// NotEqual is like Go's '!=', for example:
	//  if a != b {
	//      ...
	//  }
	NotEqual
)
