// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clock

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32                      = syscall.MustLoadDLL("kernel32.dll")
	procQueryPerformanceCounter   = kernel32.MustFindProc("QueryPerformanceCounter")
	procQueryPerformanceFrequency = kernel32.MustFindProc("QueryPerformanceFrequency")
)

func queryPerformanceCounter(performanceCount unsafe.Pointer) uintptr {
	ret, _, _ := procQueryPerformanceCounter.Call(
		uintptr(performanceCount),
	)
	return ret
}

func queryPerformanceFrequency(performanceFrequency unsafe.Pointer) uintptr {
	ret, _, _ := procQueryPerformanceFrequency.Call(
		uintptr(performanceFrequency),
	)
	return ret
}

const (
	// Note: Due to some BIOS issues in multi-core CPU's, the value returned
	// from QueryPerformanceCounter() cannot always be trusted, and may be
	// either smaller or larger.
	//
	// We solve this issue here by comparing with the time modules (less
	// precise) values.

	// We want to keep the precision QueryPerformanceCounter() gives us, and at
	// the same time keep the safety time.Now() gives us, so this allowance
	// determines how much QueryPerformanceCounter values are allowed to differ
	// from time.Now() values, before being considered invalid.
	hrAllowance = 1 * time.Millisecond

	// If the value QueryPerformanceCounter() returns is considered outside the
	// allowance, then we will attempt to re-query using
	// QueryPerformanceCounter() this many times, before settling with the
	// value time.Now() returns.
	//
	// Our hope with this is that multiple calls to QueryPerformanceCounter on
	// AMD dual core CPU's with bugged BIOS will eventually give us the correct
	// processor value after one of these attempts.
	hrAttempts = 4

	minDelta = 100 * time.Microsecond
)

var (
	start          uint64
	freqNs         float64
	doFallback     bool
	lastQueryValue time.Duration
	programStart   = time.Now()
)

func highResTimeFallback() time.Duration {
	s := time.Since(programStart)
	if s < minDelta {
		s = minDelta
	}
	return s
}

func init() {
	if queryPerformanceCounter(unsafe.Pointer(&start)) == 0 {
		doFallback = true
		return
	}

	var freq uint64
	if queryPerformanceFrequency(unsafe.Pointer(&freq)) == 0 {
		doFallback = true
		return
	}
	if freq <= 0 {
		doFallback = true
		return
	}
	freqNs = float64(freq) / 1e9
}

// getTime returns the number of milliseconds that have elapsed since the
// program started
func getTime() time.Duration {
	if doFallback {
		return highResTimeFallback()
	}

	lowResolution := highResTimeFallback()

	var greaterThanAttempt, lessThanAttempt, lastQueryAttempt int

attempt:
	var now uint64
	if queryPerformanceCounter(unsafe.Pointer(&now)) == 0 {
		doFallback = true
		return highResTimeFallback()
	}

	highResolution := time.Duration(float64(now-start) / freqNs)

	// Handle the case that QueryPerformanceCounter() gave us an number greater
	// than our maximum allowance, by attempting to query hrAttempts times
	// again, and then falling back to the low resolution time
	for highResolution > lowResolution+hrAllowance && greaterThanAttempt < hrAttempts {
		greaterThanAttempt += 1
		goto attempt
	}
	if greaterThanAttempt == hrAttempts {
		// Fallback to low resolution since we got some very innacurate results
		highResolution = lowResolution
	}

	// Handle the case that QueryPerformanceCounter() gave us an number smaller
	// than our maximum allowance, by attempting to query hrAttempts times
	// again, and then falling back to the low resolution time
	for highResolution < lowResolution-hrAllowance && lessThanAttempt < hrAttempts {
		lessThanAttempt += 1
		goto attempt
	}
	if lessThanAttempt == hrAttempts {
		// Fallback to low resolution since we got some very innacurate results
		highResolution = lowResolution
	}

	// Handle the case that QueryPerformanceCounter() gave us an number smaller
	// than the previous time we called it, by querying hrAttempts times again,
	// and then falling back to the low resolution timer.
	if lastQueryValue >= 0 {
		for highResolution < lastQueryValue && lastQueryAttempt < hrAttempts {
			lastQueryAttempt += 1
			goto attempt
		}
		if lastQueryAttempt == hrAttempts {
			highResolution = lowResolution
			if highResolution < lastQueryValue {
				highResolution = lastQueryValue
			}
		}
	}

	lastQueryValue = highResolution
	return highResolution
}
