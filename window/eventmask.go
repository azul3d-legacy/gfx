// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package window

import "math"

// EventMask is a bitmask of event types. They can be combined, for instance:
//
//  mask := GenericEvents
//  mask |= MouseEvents
//  mask |= KeyboardEvents
//
// would select generic, mouse, and keyboard events.
type EventMask uint32

const (
	// NoEvents is a event mask matching no events at all.
	NoEvents EventMask = 0

	// CloseEvents is a event mask matching window.Close events.
	CloseEvents EventMask = 1 << 0

	// DamagedEvents is a event mask matching window.Damaged events.
	DamagedEvents EventMask = 1 << 1

	// CursorMovedEvents is a event mask matching window.CursorMoved events.
	CursorMovedEvents EventMask = 1 << 2

	// CursorEnterEvents is a event mask matching window.CursorEnter events.
	CursorEnterEvents EventMask = 1 << 3

	// CursorExitEvents is a event mask matching window.CursorExit events.
	CursorExitEvents EventMask = 1 << 4

	// MinimizedEvents is a event mask matching window.Minimized events.
	MinimizedEvents EventMask = 1 << 5

	// RestoredEvents is a event mask matching window.Restored events.
	RestoredEvents EventMask = 1 << 6

	// GainedFocusEvents is a event mask matching window.GainedFocus events.
	GainedFocusEvents EventMask = 1 << 7

	// LostFocusEvents is a event mask matching window.LostFocus events.
	LostFocusEvents EventMask = 1 << 8

	// MovedEvents is a event mask matching window.Moved events.
	MovedEvents EventMask = 1 << 9

	// ResizedEvents is a event mask matching window.Resized events.
	ResizedEvents EventMask = 1 << 10

	// FramebufferResizedEvents is a event mask matching
	// window.FramebufferResized events.
	FramebufferResizedEvents EventMask = 1 << 11

	// ItemsDroppedEvents is a event mask matching window.ItemsDropped events.
	ItemsDroppedEvents EventMask = 1 << 12

	// MouseEvents is a event mask matching mouse.Event events.
	MouseEvents EventMask = 1 << 13

	// MouseScrolledEvents is a event mask matching mouse.Scrolled events.
	MouseScrolledEvents EventMask = 1 << 14

	// KeyboardTypedEvents is a event mask matching keyboard.TypedEvent events.
	KeyboardTypedEvents EventMask = 1 << 15

	// KeyboardStateEvents is a event mask matching keyboard.StateEvent events.
	KeyboardStateEvents EventMask = 1 << 16

	// AllEvents is a event mask matching all possible events.
	AllEvents = EventMask(math.MaxUint32)
)

