// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webgl

import (
	"image"
	"io"

	"azul3d.org/gfx.v2-dev"
)

// Device is a WebGL graphics device.
type Device interface {
	gfx.Device

	// Exec returns the device's execution channel.
	//
	// Whenever the device needs to perform a OpenGL task of any sort it is
	// done through this execution channel.
	//
	// If a function returns true, it is effectively a signal that the device's
	// canvas had it's Render() method called. Thus the frame is complete and
	// has been fully rendered, and you should now swap the window's buffers.
	//
	// The functions sent to this channel must be executed under the presence
	// of an OpenGL context.
	Exec() chan func() bool

	// UpdateBounds updates the effective bounding rectangle of this device.
	//
	// It must be called whenever the OpenGL framebuffer should change (e.g. on
	// window resize).
	UpdateBounds(bounds image.Rectangle)

	// SetDebugOutput sets the writer, w, to write debug output to. It will
	// mostly contain just shader debug information, but other information may
	// be written in future versions as well.
	SetDebugOutput(w io.Writer)

	// Destroy immediately destroys this device and it's associated assets.
	Destroy()
}

// Option represents a single option function.
type Option func(d *device)

// DebugOutput specifies the writer, w, as the destination for the device to
// write debug output to.
//
// It will mostly contain just shader debug information, but other information
// may be written in future versions as well.
func DebugOutput(w io.Writer) Option {
	return func(d *device) {
		d.SetDebugOutput(w)
	}
}

// New returns a new WebGL graphics device. If any error occurs it is returned
// along with a nil device.
//
// The ctx argument must be a JavaScript WebGLRenderingContext object.
func New(ctx interface{}, opts ...Option) (Device, error) {
	d, err := newDevice(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return d, nil
}
