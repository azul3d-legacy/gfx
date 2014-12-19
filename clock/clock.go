// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clock

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// Clock is a high resolution clock for measuring real-time application
// statistics.
type Clock struct {
	access sync.RWMutex

	delta, maxDelta, fixedDelta, startTime, lastFrameTime, maxFrameRateSleep time.Duration
	frameCount, frameRateFrames                                              uint64

	averageFrameSamples                                           []float64
	frameRate, maxFrameRate, averageFrameRate, frameRateDeviation float64
}

// FrameRate returns the number of frames per second according to this Clock.
func (c *Clock) FrameRate() float64 {
	c.access.RLock()
	defer c.access.RUnlock()

	return c.frameRate
}

// FrameRateDeviation returns the standard deviation of the frame times that
// have occured over the last AverageFrameRateSamples() frames.
func (c *Clock) FrameRateDeviation() float64 {
	c.access.RLock()
	defer c.access.RUnlock()

	return c.frameRateDeviation
}

// AverageFrameRate returns the average number of frames per second that have
// occured over the last Clock.AverageFrameRateSamples() frames.
func (c *Clock) AverageFrameRate() float64 {
	c.access.RLock()
	defer c.access.RUnlock()

	return c.averageFrameRate
}

// SetAverageFrameRateSamples specifies the number of previous frames to sample
// each frame to determine the average frame rate.
//
// Note: This means allocating an []float64 of size n, so be thoughtful.
func (c *Clock) SetAverageFrameRateSamples(n int) {
	c.access.Lock()
	defer c.access.Unlock()

	c.averageFrameSamples = make([]float64, n)
}

// AverageFrameRateSamples returns the number of previous frames that are
// sampled each frame to determine the average frame rate.
func (c *Clock) AverageFrameRateSamples() int {
	c.access.RLock()
	defer c.access.RUnlock()

	return len(c.averageFrameSamples)
}

// SetFrameCount specifies the current number of frames that have rendered.
func (c *Clock) SetFrameCount(count uint64) {
	c.access.Lock()
	defer c.access.RUnlock()

	c.frameCount = count
}

// ResetFrameCount resets the frame counter of this Clock to zero.
//
// Short hand for Clock.SetFrameCount(0)
func (c *Clock) ResetFrameCount() {
	c.SetFrameCount(0)
}

// FrameCount returns the number of frames that have rendered.
func (c *Clock) FrameCount() uint64 {
	c.access.RLock()
	defer c.access.RUnlock()

	return c.frameCount
}

// SetMaxDelta specifies an duration which will serve as the maximum duration
// returned by Clock.Delta().
//
// Zero is considered "no maximum delta".
func (c *Clock) SetMaxDelta(max time.Duration) {
	c.access.Lock()
	defer c.access.Unlock()

	c.maxDelta = max
}

// MaxDelta returns the duration which serves as the maximum duration returned
// by Clock.Delta()
//
// Zero is considered "no maximum delta".
func (c *Clock) MaxDelta() time.Duration {
	c.access.RLock()
	defer c.access.RUnlock()

	return c.maxDelta
}

// SetMaxFrameRate specifies an maximum frame rate, calls to Clock.Tick() will
// block for whatever time is significant enough to ensure that the frame rate
// is at max this number.
//
// If max is zero, it is considered "no maximum frame rate".
//
// If max is less than zero, an panic occurs.
func (c *Clock) SetMaxFrameRate(max float64) {
	c.access.Lock()
	defer c.access.Unlock()

	if max < 0 {
		panic("Clock.SetMaxFrameRate(): Maximum frame rate cannot be less than zero!")
	}
	c.maxFrameRate = max
}

// MaxFrameRate returns the maximum frame rate of this Clock, as it was set
// previously by Clock.SetMaxFrameRate().
func (c *Clock) MaxFrameRate() float64 {
	c.access.RLock()
	defer c.access.RUnlock()

	return c.maxFrameRate
}

// SetFixedDelta specifies an duration to be handed out via Clock.Delta()
// instead of the actual calculated delta.
func (c *Clock) SetFixedDelta(delta time.Duration) {
	c.access.Lock()
	defer c.access.Unlock()

	c.fixedDelta = delta
}

// FixedDelta returns the duration which is to be handed out via Clock.Delta()
// instead of the actual calculated delta.
//
// If time.Duration(0) is returned, then there is no fixed delta specified
// currently.
func (c *Clock) FixedDelta() time.Duration {
	c.access.RLock()
	defer c.access.RUnlock()

	return c.fixedDelta
}

// Delta returns the time between the start of the current frame and the start
// of the last frame.
//
// If Clock.FixedDelta() returns non-zero, then this function returns that
// value instead.
//
// The value returned will be clamped to clock.MaxDelta(), regardless if the
// value returned would otherwise be larger.
//
// The duration returned will never be less than zero as long as clock.Tick()
// has been called at least once previously.
func (c *Clock) Delta() time.Duration {
	c.access.RLock()
	defer c.access.RUnlock()

	if c.fixedDelta != 0 {
		return c.fixedDelta
	}

	if c.maxDelta > 0 {
		if c.delta > c.maxDelta {
			return c.maxDelta
		}
	}
	return c.delta
}

// Dt is short-hand for:
//
//  dt := float64(c.Delta()) / float64(time.Second)
//
// which is useful for applying movement over time.
func (c *Clock) Dt() float64 {
	return float64(c.Delta()) / float64(time.Second)
}

// LastFrame returns the time at which the last frame began, in time since the
// program started.
func (c *Clock) LastFrame() time.Duration {
	c.access.RLock()
	defer c.access.RUnlock()

	return c.lastFrameTime
}

// ResetLastFrame resets this Clock's last frame time to the current real time,
// as if the frame had just begun.
func (c *Clock) ResetLastFrame() {
	c.lastFrameTime = getTime()
}

// Tick signals to this Clock that an new frame has just begun.
func (c *Clock) Tick() {
	c.access.Lock()
	defer c.access.Unlock()

	firstFrame := false
	if c.frameCount == 0 {
		firstFrame = true
		c.frameCount = 1
	}

	frameStartTime := getTime()

	// Frames per second
	calcFrameRate := func() {
		// Calculate time difference between this frame and the last frame
		c.delta = getTime() - c.lastFrameTime
		if c.delta > 0 {
			c.frameRate = float64(time.Second / c.delta)
		}
	}

	// We need to calculate the frame rate and delta times right now, because
	// we use them when considering how long to sleep for in the event we have
	// an maximum frame rate.
	calcFrameRate()

	if c.maxFrameRate > 0 {
		if false {
			fmt.Println(c.frameRate, c.maxFrameRate, c.delta)
		}
		if c.frameRate > c.maxFrameRate {
			// Sleep long enough that we stay under the max frame rate.
			timeToSleep := time.Duration(float64(time.Second)/c.maxFrameRate) - c.delta
			time.Sleep(timeToSleep)

			// Calculate frame rate now that we're done using time.Sleep() for
			// sure.
			calcFrameRate()
		}
	}

	// Update the average samples
	for i, sample := range c.averageFrameSamples {
		if i-1 >= 0 {
			c.averageFrameSamples[i-1] = sample
		}
	}
	c.averageFrameSamples[len(c.averageFrameSamples)-1] = c.delta.Seconds()

	// Calculate the average frame rate.
	c.averageFrameRate = 0
	for _, sample := range c.averageFrameSamples {
		c.averageFrameRate += sample
	}
	c.averageFrameRate /= float64(len(c.averageFrameSamples))

	// Store for calculation deviation further down
	averageFrameRateDelta := c.averageFrameRate

	// Convert to frames per second
	c.averageFrameRate = 1.0 / c.averageFrameRate

	// Calculate the standard deviation of frame times
	variance := 0.0
	for i, sample := range c.averageFrameSamples {
		if i < len(c.averageFrameSamples)-1 {
			diff := sample - averageFrameRateDelta
			variance += (diff * diff)
		}
	}
	c.frameRateDeviation = math.Sqrt(variance / float64(len(c.averageFrameSamples)))

	c.lastFrameTime = frameStartTime

	if !firstFrame {
		c.frameCount += 1
	}
}

// Time returns the duration of time that has passed since this clock started
// or was last reset.
func (c *Clock) Time() time.Duration {
	c.access.RLock()
	defer c.access.RUnlock()

	return getTime() - c.startTime
}

// Reset resets this clock's starting time, as if it had just been created.
func (c *Clock) Reset() {
	c.startTime = getTime()
}

// New returns an new *Clock, with:
//
// It's start time set to the current time (via Clock.Reset).
//
// It's maximum frame rate set to 75 (Note: This is good practice because not
// all computers have working support for high resolution clocks, by setting an
// maximum frame rate, you ensure that you will never get Clock.Delta() values
// equal to zero).
//
// It's number of average frame rate samples set to 120 (via
// Clock.SetAverageFrameRateSamples).
func New() *Clock {
	c := new(Clock)
	c.Reset()
	c.SetMaxFrameRate(75)
	c.SetAverageFrameRateSamples(120)
	return c
}
