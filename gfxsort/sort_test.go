// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfxsort

import (
	"math/rand"
	"sort"
	"testing"

	"azul3d.org/gfx.v2-dev"
	"azul3d.org/lmath.v1"
)

func TestSortByDist(t *testing.T) {
	a := gfx.NewObject()
	a.Transform.SetPos(lmath.Vec3{10, 10, 10})

	b := gfx.NewObject()
	b.Transform.SetPos(lmath.Vec3{-10, 2, 2})

	c := gfx.NewObject()
	c.Transform.SetPos(lmath.Vec3{0, 6, 5})

	byDist := ByDist{
		Objects: []*gfx.Object{a, b, c, a, b, c, b, c, a},
		Target:  lmath.Vec3{0, 0, 0},
	}
	sort.Sort(byDist)

	for i := 0; i < 3; i++ {
		p := byDist.Objects[i].Transform.Pos()
		if p != a.Pos() {
			t.Fail()
		}
	}

	for i := 3; i < 6; i++ {
		p := byDist.Objects[i].Transform.Pos()
		if p != b.Pos() {
			t.Fail()
		}
	}

	for i := 6; i < 9; i++ {
		p := byDist.Objects[i].Transform.Pos()
		if p != c.Pos() {
			t.Fail()
		}
	}
}

func sortByDist(shifts, amount int, b *testing.B, standard bool) {
	b.StopTimer()
	byDist := ByDist{
		Objects: make([]*gfx.Object, amount),
		Target: lmath.Vec3{
			rand.Float64(),
			rand.Float64(),
			rand.Float64(),
		},
	}
	for i := 0; i < amount; i++ {
		byDist.Objects[i] = gfx.NewObject()
	}

	for _, o := range byDist.Objects {
		o.Transform.SetPos(lmath.Vec3{
			rand.Float64(),
			rand.Float64(),
			rand.Float64(),
		})
	}
	b.StartTimer()

	if standard {
		sort.Sort(byDist)
	} else {
		InsertionSort(byDist)
	}

	for i := 0; i < shifts; i++ {
		// Test that the sorting algorithm exploits temporal coherence by shifting
		// a eighth of the objects by a random small amount.
		b.StopTimer()
		for _, o := range byDist.Objects[:len(byDist.Objects)/8] {
			offset := lmath.Vec3{
				rand.Float64() * 0.1,
				rand.Float64() * 0.1,
				rand.Float64() * 0.1,
			}
			o.Transform.SetPos(o.Transform.Pos().Add(offset))
		}
		b.StartTimer()

		if standard {
			sort.Sort(byDist)
		} else {
			InsertionSort(byDist)
		}
	}
}

func BenchmarkDistSortOpt250(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByDist(250, 250, b, false)
	}
}
func BenchmarkDistSortOpt500(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByDist(250, 500, b, false)
	}
}
func BenchmarkDistSortOpt1k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByDist(250, 1000, b, false)
	}
}
func BenchmarkDistSortOpt5k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByDist(250, 5000, b, false)
	}
}
func BenchmarkDistSortStd250(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByDist(250, 250, b, true)
	}
}
func BenchmarkDistSortStd500(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByDist(250, 500, b, true)
	}
}
func BenchmarkDistSortStd1k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByDist(250, 1000, b, true)
	}
}
func BenchmarkDistSortStd5k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByDist(250, 5000, b, true)
	}
}

func TestSortByState(t *testing.T) {
	a := gfx.NewObject()
	a.State.Dithering = true
	a.State.DepthTest = true
	a.State.DepthWrite = true
	a.State.WriteRed = false
	a.State.WriteGreen = true
	a.State.WriteBlue = false
	a.State.WriteAlpha = true

	b := gfx.NewObject()
	b.State.Dithering = false
	b.State.DepthTest = true
	b.State.DepthWrite = false
	a.State.WriteRed = false
	a.State.WriteGreen = false
	a.State.WriteBlue = true
	a.State.WriteAlpha = true

	var l = []*gfx.Object{a, b, a, b, a, a, b, b, a, a, a, a, b, b, b, b}
	sort.Sort(ByState(l))

	for i := 0; i < 8; i++ {
		s := l[i].State
		if !s.Dithering || !s.DepthTest || !s.DepthWrite {
			t.Fail()
		}
	}

	for i := 8; i < 16; i++ {
		s := l[i].State
		if s.Dithering || !s.DepthTest || s.DepthWrite {
			t.Fail()
		}
	}
}

func sortByState(amount int, b *testing.B) {
	b.StopTimer()
	objs := make([]*gfx.Object, amount)
	for i := 0; i < amount; i++ {
		objs[i] = gfx.NewObject()
	}

	randBool := func() bool {
		return (rand.Int() % 2) == 0
	}

	for _, o := range objs {
		o.State = gfx.State{
			WriteRed:    randBool(),
			WriteGreen:  randBool(),
			WriteBlue:   randBool(),
			WriteAlpha:  randBool(),
			Dithering:   randBool(),
			DepthTest:   randBool(),
			DepthWrite:  randBool(),
			StencilTest: randBool(),
		}
	}
	b.StartTimer()

	sort.Sort(ByState(objs))
}

func BenchmarkStateSort250(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByState(250, b)
	}
}

func BenchmarkStateSort500(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByState(500, b)
	}
}

func BenchmarkStateSort1k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByState(1000, b)
	}
}

func BenchmarkStateSort5k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortByState(5000, b)
	}
}
