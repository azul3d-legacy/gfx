// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"azul3d.org/gfx.v1"
	"azul3d.org/gfx/gl2.v1/internal/resize"
	"azul3d.org/native/gl.v1"
	"fmt"
	"image"
	"image/draw"
	"math"
	"runtime"
	"unsafe"
)

type nativeTexture struct {
	r              *Renderer
	id             uint32
	internalFormat int32
	width, height  int
}

func (n *nativeTexture) Destroy() {
	finalizeTexture(n)
}

func finalizeTexture(n *nativeTexture) {
	n.r.texturesToFree.Lock()
	n.r.texturesToFree.slice = append(n.r.texturesToFree.slice, n.id)
	n.r.texturesToFree.Unlock()
}

func fbErrorString(err int32) string {
	switch err {
	case gl.FRAMEBUFFER_INCOMPLETE_ATTACHMENT:
		return "GL_FRAMEBUFFER_INCOMPLETE_ATTACHMENT"
	case 36057: //gl.FRAMEBUFFER_INCOMPLETE_DIMENSIONS:
		return "GL_FRAMEBUFFER_INCOMPLETE_DIMENSIONS"
	case gl.FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT:
		return "GL_FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT"
	case gl.FRAMEBUFFER_UNSUPPORTED:
		return "GL_FRAMEBUFFER_UNSUPPORTED"
	}
	return fmt.Sprintf("%d", err)
}

func (n *nativeTexture) Download(rect image.Rectangle, complete chan image.Image) {
	if !n.r.glArbFramebufferObject {
		// We don't have GL_ARB_framebuffer_object extensions, we can't do
		// this at all.
		n.r.logf("Download(): GL_ARB_framebuffer_object not supported; returning nil\n")
		complete <- nil
		return
	}

	if n.internalFormat != gl.RGBA {
		n.r.logf("Download(): invalid (non-RGBA) texture format; returning nil\n")
		complete <- nil
		return
	}

	n.r.LoaderExec <- func() {
		// Create a FBO, bind it now.
		var fbo uint32
		n.r.loader.GenFramebuffers(1, &fbo)
		n.r.loader.Execute()
		n.r.loader.BindFramebuffer(gl.FRAMEBUFFER, fbo)

		n.r.loader.BindTexture(gl.TEXTURE_2D, n.id)
		n.r.loader.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		n.r.loader.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		n.r.loader.BindTexture(gl.TEXTURE_2D, 0)

		// Attach the texture to the FBO.
		n.r.loader.FramebufferTexture2D(
			gl.FRAMEBUFFER,
			gl.COLOR_ATTACHMENT0,
			gl.TEXTURE_2D,
			n.id, // texture ID
			0,    // level
		)

		// If the rectangle is empty use the entire area.
		bounds := image.Rect(0, 0, n.width, n.height)
		if rect.Empty() {
			rect = bounds
		} else {
			// Intersect the rectangle with the texture's bounds.
			rect = bounds.Intersect(rect)
		}

		status := n.r.loader.CheckFramebufferStatus(gl.FRAMEBUFFER)
		if status != gl.FRAMEBUFFER_COMPLETE {
			// Log the error.
			n.r.logf("Download(): glCheckFramebufferStatus() failed! Status == %s.\n", fbErrorString(status))
			complete <- nil
			return
		}

		// Read texture pixels.
		img := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
		x, y, w, h := convertRect(rect, bounds)
		n.r.loader.ReadPixels(
			x, y, w, h,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			unsafe.Pointer(&img.Pix[0]),
		)

		// Delete the FBO.
		n.r.loader.BindFramebuffer(gl.FRAMEBUFFER, 0)
		n.r.loader.DeleteFramebuffers(1, &fbo)

		// Flush and execute.
		n.r.loader.Flush()
		n.r.loader.Execute()

		complete <- img
	}
}

// Implements gfx.Renderer interface.
func (r *Renderer) RenderToTexture(t *gfx.Texture) gfx.Canvas {
	return r
}

func verticalFlip(img *image.RGBA) {
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

func nearestPOT(k int) int {
	// See:
	//
	// http://en.wikipedia.org/wiki/Power_of_two#Algorithm_to_convert_any_number_into_nearest_power_of_two_numbers
	return int(math.Pow(2, math.Ceil(math.Log(float64(k))/math.Log(2))))
}

func prepareImage(npot bool, img image.Image) *image.RGBA {
	bounds := img.Bounds()

	if !npot {
		// Convert the image to a power-of-two size if it's not already.
		x, y := bounds.Dx(), bounds.Dy()
		potX, potY := nearestPOT(x), nearestPOT(y)
		if x != potX || y != potY {
			if potX < x && potY < y {
				// Resample is faster but only works for scaling down.
				img = resize.Resample(img, bounds, potX, potY)
			} else {
				// Resize works in all cases.
				img = resize.Resize(img, bounds, potX, potY)
			}
		}

		// Update known bounds.
		bounds = img.Bounds()
	}

	// Currently, images must be RGBA format. Convert now if needed.
	rgba, ok := img.(*image.RGBA)
	if !ok {
		// Convert the image to RGBA.
		rgba = image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
		draw.Draw(rgba, rgba.Bounds(), img, bounds.Min, draw.Src)
	}
	return rgba
}

// Implements gfx.Downloadable interface.
func (r *Renderer) Download(rect image.Rectangle, complete chan image.Image) {
	r.RenderExec <- func() bool {
		bounds := r.Bounds()

		// If the rectangle is empty use the entire area.
		if rect.Empty() {
			rect = bounds
		} else {
			// Intersect the rectangle with the renderer's bounds.
			rect = bounds.Intersect(rect)
		}

		img := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
		x, y, w, h := convertRect(rect, bounds)
		r.render.ReadPixels(
			x, y, w, h,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			unsafe.Pointer(&img.Pix[0]),
		)

		// Flush and execute.
		r.render.Flush()
		r.render.Execute()

		// We must vertically flip the image.
		verticalFlip(img)

		// Yield for occlusion query results, if any are available.
		r.queryYield()

		complete <- img
		return false
	}
}

func convertWrap(w gfx.TexWrap) int32 {
	switch w {
	case gfx.Repeat:
		return gl.REPEAT
	case gfx.Clamp:
		return gl.CLAMP_TO_EDGE
	case gfx.BorderColor:
		return gl.CLAMP_TO_BORDER
	case gfx.Mirror:
		return gl.MIRRORED_REPEAT
	}
	panic("Invalid wrap mode")
}

func convertFilter(f gfx.TexFilter) int32 {
	switch f {
	case gfx.Nearest:
		return gl.NEAREST
	case gfx.Linear:
		return gl.LINEAR
	case gfx.NearestMipmapNearest:
		return gl.NEAREST_MIPMAP_NEAREST
	case gfx.LinearMipmapNearest:
		return gl.LINEAR_MIPMAP_NEAREST
	case gfx.NearestMipmapLinear:
		return gl.NEAREST_MIPMAP_LINEAR
	case gfx.LinearMipmapLinear:
		return gl.LINEAR_MIPMAP_LINEAR
	}
	panic("invalid filter.")
}

func (r *Renderer) freeTextures() {
	// Lock the list.
	r.texturesToFree.Lock()

	if len(r.texturesToFree.slice) > 0 {
		// Free the textures.
		r.loader.DeleteTextures(uint32(len(r.texturesToFree.slice)), &r.texturesToFree.slice[0])

		// Flush and execute OpenGL commands.
		r.loader.Flush()
		r.loader.Execute()
	}

	// Slice to zero, and unlock.
	r.texturesToFree.slice = r.texturesToFree.slice[:0]
	r.texturesToFree.Unlock()
}

const (
	// We really should try to get our GL bindings to wrap extensions..
	// See: http://www.opengl.org/registry/specs/EXT/texture_compression_s3tc.txt
	glCOMPRESSED_RGB_S3TC_DXT1_EXT  = 0x83F0
	glCOMPRESSED_RGBA_S3TC_DXT1_EXT = 0x83F1
	glCOMPRESSED_RGBA_S3TC_DXT3_EXT = 0x83F2
	glCOMPRESSED_RGBA_S3TC_DXT5_EXT = 0x83F3
)

func convertTexFormat(f gfx.TexFormat) int32 {
	switch f {
	case gfx.RGBA:
		return gl.RGBA
	case gfx.RGB:
		return gl.RGB
	case gfx.DXT1:
		return glCOMPRESSED_RGB_S3TC_DXT1_EXT
	case gfx.DXT1RGBA:
		return glCOMPRESSED_RGBA_S3TC_DXT1_EXT
	case gfx.DXT3:
		return glCOMPRESSED_RGBA_S3TC_DXT3_EXT
	case gfx.DXT5:
		return glCOMPRESSED_RGBA_S3TC_DXT5_EXT
	}
	panic("unknown format")
}

// Implements gfx.Renderer interface.
func (r *Renderer) LoadTexture(t *gfx.Texture, done chan *gfx.Texture) {
	// Lock the texture until we are done loading it.
	t.Lock()
	if t.Loaded {
		// Texture is already loaded, signal completion if needed and return
		// after unlocking.
		t.Unlock()
		select {
		case done <- t:
		default:
		}
		return
	}

	f := func() {
		// Create texture ID.
		native := &nativeTexture{
			r: r,
		}
		r.loader.GenTextures(1, &native.id)
		r.loader.Execute()

		// Bind the texture.
		r.loader.BindTexture(gl.TEXTURE_2D, native.id)

		// Set wrap mode.
		r.loader.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_BASE_LEVEL, 0)
		r.loader.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAX_LEVEL, 0)

		// Determine appropriate internal image format.
		targetFormat := convertTexFormat(t.Format)
		internalFormat := gl.RGBA
		for _, format := range r.compressedTextureFormats {
			if format == targetFormat {
				internalFormat = format
				break
			}
		}
		native.internalFormat = internalFormat

		// Upload the image.
		src := prepareImage(r.gpuInfo.NPOT, t.Source)
		bounds := src.Bounds()
		r.loader.TexImage2D(
			gl.TEXTURE_2D,
			0,
			internalFormat,
			uint32(bounds.Dx()),
			uint32(bounds.Dy()),
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			unsafe.Pointer(&src.Pix[0]),
		)
		native.width = bounds.Dx()
		native.height = bounds.Dy()

		// Unbind texture to avoid carrying OpenGL state.
		r.loader.BindTexture(gl.TEXTURE_2D, 0)

		// Flush, Finish and execute OpenGL commands.
		r.loader.Flush()
		// Use Finish() to avoid accessing the texture before upload has completed, see:
		//  http://higherorderfun.com/blog/2011/05/26/multi-thread-opengl-texture-loading/
		r.loader.Finish()
		r.loader.Execute()

		// Mark the texture as loaded.
		t.Loaded = true
		t.NativeTexture = native
		t.ClearData()

		// Attach a finalizer to the texture that will later free it.
		runtime.SetFinalizer(native, finalizeTexture)

		// Unlock, signal completion, and return.
		t.Unlock()
		select {
		case done <- t:
		default:
		}
	}

	select {
	case r.LoaderExec <- f:
	default:
		go func() {
			r.LoaderExec <- f
		}()
	}
}
