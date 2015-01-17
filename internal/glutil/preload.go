// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

import (
	"fmt"
	"strings"

	"azul3d.org/gfx.v2-unstable"
)

// PreLoadShader implements the precusor to gfx.Canvas.LoadShader; it returns a
// boolean value whether or not loading of the given shader should continue,
// and a (warning) error, if any.
func PreLoadShader(s *gfx.Shader, done chan *gfx.Shader) (doLoad bool, err error) {
	// signal is used to signal completion to the done channel in a non
	// blocking way.
	signal := func() {
		select {
		case done <- s:
		default:
		}
	}

	// If the shader is already loaded or there was previously an error loading
	// it then signal completion and perform no further loading.
	if s.Loaded || len(s.Error) > 0 {
		signal()
		return false, nil
	}

	// A vertex or fragment shader with no code at all causes an undefined
	// behavior and can cause some drivers to crash. It is an error and as such
	// no further loading of the shader should occur.
	if len(strings.TrimSpace(string(s.GLSL.Vertex))) == 0 {
		err = fmt.Errorf("%s | Vertex shader with no source code.", s.Name)
		s.Error = append(s.Error, []byte(err.Error())...)
		signal()
		return false, err
	}
	if len(strings.TrimSpace(string(s.GLSL.Fragment))) == 0 {
		err = fmt.Errorf("%s | Fragment shader with no source code.", s.Name)
		s.Error = append(s.Error, []byte(err.Error())...)
		signal()
		return false, err
	}
	return true, nil
}
