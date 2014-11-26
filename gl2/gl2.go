// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import(
	"errors"
	"image"
	"io"

	"azul3d.org/gfx.v2-dev"
)

// Used when attempting to create an OpenGL 2.0 renderer in a lesser OpenGL
// context.
var ErrInvalidVersion = errors.New("invalid OpenGL version; must be at least OpenGL 2.0")

// Renderer is an OpenGL 2 based graphics renderer.
//
// It runs independant of the window management library being used (GLFW, SDL,
// QT, etc), all it needs is a valid OpenGL 2 context.
//
// The renderer primarily uses two independant OpenGL contexts, one is used for
// rendering and one is used for managing resources like meshes, textures, and
// shaders which allows for asynchronous loading (although it is also possible
// to use only a single OpenGL context for windowing libraries that do not
// support multiple, but this will inheritly disable asynchronous loading).
type Renderer interface {
	gfx.Renderer

	// RenderExec returns the renderer execution channel.
	RenderExec() chan func() bool

	// UpdateBounds updates the effective bounding rectangle of this renderer. It
	// must be called whenever the OpenGL canvas size should change (e.g. on window
	// resize).
	UpdateBounds(bounds image.Rectangle)

	// SetDebugOutput sets the writer, w, to write debug output to. It will mostly
	// contain just shader debug information, but other information may be written
	// in future versions as well.
	SetDebugOutput(w io.Writer)
}
