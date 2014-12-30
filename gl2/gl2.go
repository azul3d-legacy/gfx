// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"errors"
	"image"
	"io"

	"azul3d.org/gfx.v2-dev"
)

// ErrInvalidVersion is returned by New when attempting to create a OpenGL 2
// device in a lesser version OpenGL context.
var ErrInvalidVersion = errors.New("invalid OpenGL version; must be at least OpenGL 2.0")

// Device is a OpenGL 2 based graphics device.
//
// It runs independant of the window management library being used (GLFW, SDL,
// QT, etc), all it needs is a valid OpenGL 2 context.
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
	//
	// This function must be called under the presence of an OpenGL context.
	Destroy()
}

// Option represents a single option function.
type Option func(d *device)

// KeepState is an option that specifies whether or not the existing OpenGL
// graphics state should be kept between frames.
//
// If this option is present, the device will save and restore the OpenGL
// graphics state before and after rendering each frame. This is needed when
// trying to cooperate with another renderer in the same OpenGL context (such
// as rendering into a QT5 user interface).
//
// If not specified, the device is able to carry OpenGL state across multiple
// frames and avoid needlessly setting OpenGL state, which is more optimal for
// performance.
//
// Do not specify this option unless you're sure that you need it.
func KeepState() Option {
	return func(d *device) {
		d.keepState = true
	}
}

// Share is an option that specifies that this device should request the other
// device to perform loading of all assets.
//
// The given other device must be from this package specifically, or else a
// panic will occur.
func Share(other Device) Option {
	return func(d *device) {
		d.shared.device = other.(*device)
	}
}

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

// New returns a new OpenGL 2 graphics device. If any error occurs it is
// returned along with a nil device.
//
// It is only safe to call this function under the presence of an OpenGL 2
// feature level context.
func New(opts ...Option) (Device, error) {
	d, err := newDevice(opts...)
	if err != nil {
		return nil, err
	}
	return d, nil
}
