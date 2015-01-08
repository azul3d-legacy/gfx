// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package window

import "math"

// EventMask is a bitmask of event types. They can be combined, for instance:
//
//  mask := window.CloseEvents
//  mask |= window.MinimizedEvents
//  mask |= window.LostFocusEvents
//
// would select events of the following types:
//
//  window.Close
//  window.Minimized
//  window.LostFocus
//
type EventMask uint32

const (
	// CloseEvents is a event mask matching window.Close events.
	CloseEvents EventMask = 1 << iota

	// DamagedEvents is a event mask matching window.Damaged events.
	DamagedEvents

	// CursorMovedEvents is a event mask matching window.CursorMoved events.
	CursorMovedEvents

	// CursorEnterEvents is a event mask matching window.CursorEnter events.
	CursorEnterEvents

	// CursorExitEvents is a event mask matching window.CursorExit events.
	CursorExitEvents

	// MinimizedEvents is a event mask matching window.Minimized events.
	MinimizedEvents

	// RestoredEvents is a event mask matching window.Restored events.
	RestoredEvents

	// GainedFocusEvents is a event mask matching window.GainedFocus events.
	GainedFocusEvents

	// LostFocusEvents is a event mask matching window.LostFocus events.
	LostFocusEvents

	// MovedEvents is a event mask matching window.Moved events.
	MovedEvents

	// ResizedEvents is a event mask matching window.Resized events.
	ResizedEvents

	// FramebufferResizedEvents is a event mask matching
	// window.FramebufferResized events.
	FramebufferResizedEvents

	// ItemsDroppedEvents is a event mask matching window.ItemsDropped events.
	ItemsDroppedEvents

	// MouseButtonEvents is a event mask matching mouse.ButtonEvent's.
	MouseButtonEvents

	// MouseScrolledEvents is a event mask matching mouse.Scrolled events.
	MouseScrolledEvents

	// KeyboardTypedEvents is a event mask matching keyboard.Typed events.
	KeyboardTypedEvents

	// KeyboardButtonEvents is a event mask matching keyboard.ButtonEvent's.
	KeyboardButtonEvents

	// NoEvents is a event mask matching no events at all.
	NoEvents EventMask = 0

	// AllEvents is a event mask matching all possible events.
	AllEvents EventMask = math.MaxUint32
)

const (
	// CursorEvents is an event mask that selects all cursor events:
	//
	//  window.CursorMoved
	//  window.CursorEnter
	//  window.CursorExit
	//
	CursorEvents EventMask = CursorMovedEvents | CursorEnterEvents | CursorExitEvents

	// MouseEvents is an event mask that selects all mouse events:
	//
	//  mouse.ButtonEvent
	//  mouse.Scrolled
	//
	MouseEvents EventMask = MouseButtonEvents | MouseScrolledEvents

	// KeyboardEvents is an event mask that selects all keyboard events:
	//
	//  keyboard.ButtonEvent
	//  keyboard.Typed
	//
	KeyboardEvents EventMask = KeyboardButtonEvents | KeyboardTypedEvents
)
