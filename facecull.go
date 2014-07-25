// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import "fmt"

// FaceCullMode represents a single face culling mode. BackFaceCulling is the
// default (zero value).
type FaceCullMode uint8

// String returns a string representation of this FaceCullMode.
// e.g. BackFaceCulling -> "BackFaceCulling"
func (f FaceCullMode) String() string {
	switch f {
	case BackFaceCulling:
		return "BackFaceCulling"
	case FrontFaceCulling:
		return "FrontFaceCulling"
	case NoFaceCulling:
		return "NoFaceCulling"
	}
	return fmt.Sprintf("FaceCullMode(%d)", f)
}

const (
	// Culls only back faces (i.e. only the front side is rendered).
	BackFaceCulling FaceCullMode = iota

	// Culls only front faces (i.e. only the back side is rendered).
	FrontFaceCulling

	// Does not cull any faces (i.e. both sides are rendered).
	NoFaceCulling
)
