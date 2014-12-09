// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package window

import "sync"

// notifierEntry is a event channel and it's associated event mask. It is the
// pair passed into the Window interface's Notify method.
type notifierEntry struct {
	ch chan<- Event
	EventMask
}

// notifier implements the Window interface's Notify method.
type notifier struct {
	sync.RWMutex
	entries []notifierEntry
}

// Implements the Window interface.
func (n *notifier) Notify(ch chan<- Event, m EventMask) {
	n.Lock()
	if m == NoEvents {
		n.deleteEntries(ch)
	} else {
		n.entries = append(n.entries, notifierEntry{ch, m})
	}
	n.Unlock()
}

// findEntry searches for the entry associated with ch and returns it's slice
// index or -1.
//
// n.Lock must be held for it to operate safely.
func (n *notifier) findEntry(ch chan<- Event) int {
	for index, ev := range n.entries {
		if ev.ch == ch {
			return index
		}
	}
	return -1
}

// deleteEntries deletes all entries associated with ch.
func (n *notifier) deleteEntries(ch chan<- Event) {
	s := n.entries
	idx := n.findEntry(ch)
	for idx != -1 {
		s = append(s[:idx], s[idx+1:]...)
		idx = n.findEntry(ch)
	}
	n.entries = s
}

// sendEvent sends the given event to all of the notifier entries whose bitmask
// matches with m.
func (n *notifier) sendEvent(ev Event, m EventMask) {
	n.RLock()
	for _, nf := range n.entries {
		if (nf.EventMask & m) != 0 {
			select {
			case nf.ch <- ev:
			default:
			}
		}
	}
	n.RUnlock()
}
