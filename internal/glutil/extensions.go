// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

import "strings"

// Extensions represents a set of OpenGL extensions. If an extension is present
// it is in the map.
type Extensions map[string]struct{}

// Present tells if the given extension string is present or not.
func (e Extensions) Present(s string) bool {
	_, ok := e[s]
	return ok
}

// Slice returns the extensions as a slice of strings.
func (e Extensions) Slice() []string {
	s := make([]string, len(e))
	i := 0
	for ext := range e {
		s[i] = ext
		i++
	}
	return s
}

// ParseExtensions parses the space-seperated list of OpenGL extensions and
// returned them.
func ParseExtensions(s string) Extensions {
	if len(s) == 0 {
		return nil
	}
	e := make(Extensions, 32) // 32 extensions, at least.
	for _, ext := range strings.Split(s, " ") {
		if len(ext) == 0 {
			continue
		}
		e[ext] = struct{}{}
	}
	return e
}
