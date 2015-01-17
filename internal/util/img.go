package util

import (
	"image"
	"math"

	"azul3d.org/gfx.v2-unstable/internal/resize"
)

func nearestPOT(k int) int {
	// See:
	//
	// http://en.wikipedia.org/wiki/Power_of_two#Algorithm_to_convert_any_number_into_nearest_power_of_two_numbers
	return int(math.Pow(2, math.Ceil(math.Log(float64(k))/math.Log(2))))
}

// POT resamples/resizes the given image to the nearest power-of-two size (if
// it is not a power-of-two size already).
func POT(img image.Image) image.Image {
	bounds := img.Bounds()
	x, y := bounds.Dx(), bounds.Dy()
	potX, potY := nearestPOT(x), nearestPOT(y)
	if x == potX && y == potY {
		// Image is a power-of-two size already.
		return img
	}

	if potX < x && potY < y {
		// Resample is faster but only works for scaling down.
		return resize.Resample(img, bounds, potX, potY)
	}

	// Resize works in all cases.
	return resize.Resize(img, bounds, potX, potY)
}

// VerticalFlip flips the given image in-place, vertically.
func VerticalFlip(img *image.RGBA) {
	b := img.Bounds()
	rowCpy := make([]uint8, b.Dx()*4)
	for r := 0; r < (b.Dy() / 2); r++ {
		topRow := img.Pix[img.PixOffset(0, r):img.PixOffset(b.Dx(), r)]

		bottomR := b.Dy() - r - 1
		bottomRow := img.Pix[img.PixOffset(0, bottomR):img.PixOffset(b.Dx(), bottomR)]

		// Save bottom row.
		copy(rowCpy, bottomRow)

		// Copy top row to bottom row.
		copy(bottomRow, topRow)

		// Copy saved bottom row to top row.
		copy(topRow, rowCpy)
	}
}
