// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"io"
	"sync"
)

// Warner can warn developers of a potential problem. It writes it's warnings
// to the underlying writer.
//
// A map of previous warnings is used to identify previous warnings, and as
// such you should be concious of potential memory overhead (i.e. only use it
// for situations where warnings should not happen generally).
type Warner struct {
	sync.RWMutex
	W    io.Writer
	prev map[string]struct{}
}

// Warn writes a single warning message out to the underlying writer (w.W). It
// will only write the message once even if multiple calls are made, thus:
//
//  w.Warn("Something bad has happened!")
//  w.Warn("Something bad has happened!")
//  w.Warn("Something bad has happened (again)!")
//
// Will produce just the following output:
//
//  Something bad has happened!
//  Something bad has happened (again)!
//
func (w *Warner) Warn(msg string) {
	w.RLock()
	defer w.RUnlock()
	if _, warnedAlready := w.prev[msg]; warnedAlready {
		return
	}
	w.prev[msg] = struct{}{}
	fmt.Fprintf(w.W, msg)
}

// Warnf is short-handed for:
//
//  w.Warn(fmt.Sprintf(format, args...))
//
func (w *Warner) Warnf(format string, args ...interface{}) {
	w.Warn(fmt.Sprintf(format, args...))
}

// NewWarner initializes and returns a new warner, using the given writer.
func NewWarner(w io.Writer) *Warner {
	return &Warner{
		W:    w,
		prev: make(map[string]struct{}),
	}
}
