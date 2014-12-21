// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package clock measures frame-based application performance.
//
// A Time function is provided, which returns the time since the program
// started, it can be very useful on it's own, as on Windows this package uses
// cgo to call QueryPerformanceFrequency and QueryPerformanceCounter, which
// provides a higher resolution time than Go's time package.
//
// An the case of an visual application, an typical use case would be creating
// an single Clock for each "renderer", and then invoke Clock.Tick() at the
// start of every frame before it is rendered.
//
// When using a maximum frame rate, Clock.Tick blocks for an little while to
// ensure running at least under Clock.MaxFrameRate(), if you simply ignore
// this blocking or push it to another goroutine, you'll lose the whole point
// of having an maximum frame rate specified.
package clock
