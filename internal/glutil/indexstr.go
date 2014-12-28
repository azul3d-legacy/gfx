// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

import "strconv"

const indexStrCacheSize = 64

// IndexStr generates index strings with the given prefix, for example:
//
//  Texture0
//  Texture1
//  Texture2
//
type IndexStr struct {
	prefix string
	cache  []string
}

// Name returns the index string for the given index.
func (i IndexStr) Name(n int) string {
	if n >= 0 && n < 64 {
		return i.cache[n]
	}
	return i.prefix + strconv.Itoa(n)
}

// NewIndexStr returns a new index string generator with the given prefix.
func NewIndexStr(prefix string) *IndexStr {
	i := &IndexStr{
		prefix: prefix,
	}
	i.cache = make([]string, indexStrCacheSize)
	for k := 0; k < indexStrCacheSize; k++ {
		i.cache[k] = i.prefix + strconv.Itoa(k)
	}
	return i
}
