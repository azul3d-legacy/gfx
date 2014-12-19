// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package clock

import (
	"time"
)

const (
	minDelta = 100 * time.Microsecond
)

var (
	programStart = time.Now()
)

// In here we simply fallback to the standard time package for systems that
// already support high resolution timers.
//
// Since this relies on system time and the user might change their time
// resulting in a negative time occuring, we enforce a positive delta duration
// of at least 100us.
func Time() time.Duration {
	s := time.Since(programStart)
	if s < minDelta {
		s = minDelta
	}
	return s
}
