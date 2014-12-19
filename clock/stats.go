// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package clock

import (
	"sync"
	"time"
)

// Stats allows you to create linear performance samples using categories.
type Stats struct {
	access sync.RWMutex

	enabled        bool
	categories     map[string][]time.Duration
	startedSamples map[string]time.Duration
}

// SetEnabled specifies whether or not to enable taking samples, if off (false), then calls to
// Begin() and End() are no-op.
func (s *Stats) SetEnabled(enabled bool) {
	s.access.Lock()
	defer s.access.Unlock()

	s.enabled = enabled
}

// Enabled tells whether or not taking samples is currently enabled or not.
func (s *Stats) Enabled() bool {
	s.access.RLock()
	defer s.access.RUnlock()

	return s.enabled
}

// Add adds the specified category, after this operation samples may be created or retrieved.
//
// the samples parameter specifies how many samples will at max exist within the category.
func (s *Stats) Add(category string, samples uint) {
	s.access.Lock()
	defer s.access.Unlock()

	s.categories[category] = make([]time.Duration, 0)
}

// Remove removes the specified category and all samples associated with it.
func (s *Stats) Remove(category string) {
	s.access.Lock()
	defer s.access.Unlock()

	delete(s.categories, category)
}

// Has tells whether or not the specified category exists.
func (s *Stats) Has(category string) bool {
	s.access.RLock()
	defer s.access.RUnlock()

	_, ok := s.categories[category]
	return ok
}

// LastSample returns the last sample (or 0) of the specified category.
func (s *Stats) LastSample(category string) time.Duration {
	s.access.RLock()
	defer s.access.RUnlock()

	samples, ok := s.categories[category]
	if !ok {
		return 0
	}
	return samples[len(samples)-1]
}

// Samples returns all samples of the specified category, or nil if the category does not exist.
func (s *Stats) Samples(category string) []time.Duration {
	s.access.RLock()
	defer s.access.RUnlock()

	samples, ok := s.categories[category]
	if !ok {
		return nil
	}

	var samplesCopy []time.Duration
	copy(samplesCopy, samples)
	return samplesCopy
}

// Begin creates an new sample under the specified category, whose time will begin being measured
// at this very moment, if taking samples is enabled.
//
// If an sample already exists (and was not ended) for this category, then the previous sample will
// be ended, and an new one started.
//
// If the specified category does not exist, an panic occurs.
func (s *Stats) Begin(category string) {
	s.access.RLock()
	defer s.access.RUnlock()

	samples, ok := s.categories[category]
	if !ok {
		panic("Begin(): Category does not exist.")
	}

	if !s.enabled {
		return
	}

	sampleStart, ok := s.startedSamples[category]
	if ok {
		// An old already-started sample exists, end it.
		sample := Time() - sampleStart
		samples = append(samples, sample)
	}

	// Create new sample, starting right now.
	s.startedSamples[category] = Time()
}

// End stops the current sample under the specified category, if taking samples is enabled.
//
// If an sample for this category was not previously started, then this function is no-op.
//
// If the specified category does not exist, an panic occurs.
func (s *Stats) End(category string) {
	s.access.RLock()
	defer s.access.RUnlock()

	samples, ok := s.categories[category]
	if !ok {
		panic("End(): Category does not exist.")
	}

	if !s.enabled {
		return
	}

	sampleStart, ok := s.startedSamples[category]
	if !ok {
		return
	}

	// An old already-started sample exists, end it.
	sample := Time() - sampleStart
	samples = append(samples, sample)

	// Clear the old sample
	delete(s.startedSamples, category)
}

// NewStats returns an new initialized *Stats struct.
func NewStats() *Stats {
	s := new(Stats)
	s.categories = make(map[string][]time.Duration)
	s.startedSamples = make(map[string]time.Duration)
	return s
}
