// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package clock measures frame-based application performance.
//
// Typically you create a single Clock object for each "renderer" and invoke
// Tick on it at the start of each frame.
//
// When using a maximum frame rate, Tick blocks just long enough to ensure that
// the application is at max running at MaxFrameRate.
package clock
