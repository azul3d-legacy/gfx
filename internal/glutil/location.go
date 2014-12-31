// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

type location struct {
	name     string
	location int
}

// LocationCache caches GLSL uniform and attribute locations.
type LocationCache struct {
	GetAttribLocation, GetUniformLocation func(name string) int

	attribs  []location
	uniforms []location
}

// FindAttrib finds the named attribute location in the cache, if it's not in
// the cache it is queried from l.GetAttribLocation directly and cached for
// later.
func (l *LocationCache) FindAttrib(name string) int {
	// Scan through the cache for it now.
	for _, a := range l.attribs {
		if a.name != name {
			continue
		}
		return a.location
	}

	// Query directly and store in the cache.
	i := l.GetAttribLocation(name)
	if i < 0 {
		i = -1
	}
	l.attribs = append(l.attribs, location{
		name:     name,
		location: i,
	})
	return i
}

// FindUniform finds the named uniform location in the cache, if it's not in
// the cache it is queried from l.GetUniformLocation directly and cached for
// later.
func (l *LocationCache) FindUniform(name string) int {
	// Scan through the cache for it now.
	for _, u := range l.uniforms {
		if u.name != name {
			continue
		}
		return u.location
	}

	// Query directly and store in the cache.
	i := l.GetUniformLocation(name)
	if i < 0 {
		i = -1
	}
	l.uniforms = append(l.uniforms, location{
		name:     name,
		location: i,
	})
	return i
}
