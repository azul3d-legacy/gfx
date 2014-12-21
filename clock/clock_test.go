// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clock

import (
	"testing"
	"time"

	"azul3d.org/lmath.v1"
)

func TestHighResolutionTime(t *testing.T) {
	lrStart := time.Now()
	hrStart := getTime()

	var diffTotal time.Duration
	for i := 0; i < 10; i++ {
		lrDiff := time.Since(lrStart)
		hrDiff := getTime() - hrStart

		diffTotal += hrDiff
		t.Logf("%d.\ttime.Since()=%d\tgetTime()=%d", i, lrDiff, hrDiff)

		lrStart = time.Now()
		hrStart = getTime()
	}

	if diffTotal <= 0 {
		t.Fail()
	}
}

func TestFrameRateLimit(t *testing.T) {
	c := New()
	c.SetMaxFrameRate(100)
	c.SetAverageFrameRateSamples(100)
	for i := 0; i < c.AverageFrameRateSamples(); i++ {
		c.Tick()
	}
	avg := c.AverageFrameRate()
	if !lmath.AlmostEqual(avg, 100, 0.05) {
		t.Log("got avg", avg)
		t.Fatal("expected avg near", 100)
	}
}
