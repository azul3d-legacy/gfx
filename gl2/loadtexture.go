// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"fmt"
	"image"
	"image/draw"
	"runtime"
	"unsafe"

	"azul3d.org/gfx.v2-dev"
	"azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"
	"azul3d.org/gfx.v2-dev/internal/glutil"
	"azul3d.org/gfx.v2-dev/internal/util"
)

type nativeTexture struct {
	r              *device
	id             uint32
	internalFormat int32
	width, height  int
	*rttCanvas
	destroyHandler func(n *nativeTexture)
}

// Generates texture ID, binds, and sets BASE/MAX mipmap levels to zero.
//
// Used by both LoadTexture and RenderToTexture methods.
func newNativeTexture(r *device, internalFormat int32, width, height int) *nativeTexture {
	tex := &nativeTexture{
		r:              r,
		internalFormat: internalFormat,
		width:          width,
		height:         height,
		destroyHandler: finalizeTexture,
	}
	gl.GenTextures(1, &tex.id)
	//gl.Execute()

	gl.BindTexture(gl.TEXTURE_2D, tex.id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_BASE_LEVEL, 0)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAX_LEVEL, 1000)
	return tex
}

// Destroy implements the gfx.Destroyable interface.
func (n *nativeTexture) Destroy() {
	n.destroyHandler(n)
}

// ChosenFormat implements the gfx.NativeTexture interface.
func (n *nativeTexture) ChosenFormat() gfx.TexFormat {
	return unconvertTexFormat(n.internalFormat)
}

func finalizeTexture(n *nativeTexture) {
	n.r.texturesToFree.Lock()
	n.r.texturesToFree.slice = append(n.r.texturesToFree.slice, n.id)
	n.r.texturesToFree.Unlock()
}

// TODO(slimsag): move to internal/glc ?
func fbErrorString(err uint32) string {
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

// Download implements the gfx.Downloadable interface.
func (n *nativeTexture) Download(rect image.Rectangle, complete chan image.Image) {
	if !n.r.glArbFramebufferObject {
		// We don't have GL_ARB_framebuffer_object extension, we can't do this
		// at all.
		n.r.logf("Download(): GL_ARB_framebuffer_object not supported; returning nil\n")
		complete <- nil
		return
	}

	if n.internalFormat != gl.RGBA {
		n.r.logf("Download(): invalid (non-RGBA) texture format; returning nil\n")
		complete <- nil
		return
	}

	n.r.renderExec <- func() bool {
		// Create a FBO, bind it now.
		var fbo uint32
		gl.GenFramebuffers(1, &fbo)
		//gl.Execute()
		gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

		gl.BindTexture(gl.TEXTURE_2D, n.id)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.BindTexture(gl.TEXTURE_2D, 0)

		// Attach the texture to the FBO.
		gl.FramebufferTexture2D(
			gl.FRAMEBUFFER,
			gl.COLOR_ATTACHMENT0,
			gl.TEXTURE_2D,
			n.id, // texture ID
			0,    // level
		)

		// Intersect the rectangle with the texture's bounds.
		bounds := image.Rect(0, 0, n.width, n.height)
		rect = bounds.Intersect(rect)

		status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
		if status != gl.FRAMEBUFFER_COMPLETE {
			// Log the error.
			n.r.logf("Download(): glCheckFramebufferStatus() failed! Status == %s.\n", fbErrorString(status))
			complete <- nil
			return false // no frame rendered.
		}

		// Read texture pixels.
		img := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
		x, y, w, h := glutil.ConvertRect(rect, bounds)
		gl.ReadPixels(
			int32(x), int32(y), int32(w), int32(h),
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			unsafe.Pointer(&img.Pix[0]),
		)

		// Delete the FBO.
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.DeleteFramebuffers(1, &fbo)

		// Flush and execute.
		gl.Flush()
		//gl.Execute()

		complete <- img
		return false // no frame rendered.
	}
}

func prepareImage(npot bool, img image.Image) *image.RGBA {
	bounds := img.Bounds()

	if !npot {
		// Convert the image to a power-of-two size if it's not already.
		img = util.POT(img)

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

// Download implements the gfx.Downloadable interface.
func (r *device) Download(rect image.Rectangle, complete chan image.Image) {
	r.hookedDownload(rect, complete, nil, nil)
}

// Implements gfx.Downloadable interface.
func (r *device) hookedDownload(rect image.Rectangle, complete chan image.Image, pre, post func()) {
	r.renderExec <- func() bool {
		if pre != nil {
			pre()
		}

		// Intersect the rectangle with the renderer's bounds.
		bounds := r.Bounds()
		rect = bounds.Intersect(rect)

		img := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
		x, y, w, h := glutil.ConvertRect(rect, bounds)
		gl.ReadPixels(
			int32(x), int32(y), int32(w), int32(h),
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			unsafe.Pointer(&img.Pix[0]),
		)

		if post != nil {
			post()
		}

		// Flush and execute.
		gl.Flush()
		//gl.Execute()

		// We must vertically flip the image.
		util.VerticalFlip(img)

		// Yield for occlusion query results, if any are available.
		r.queryYield()

		complete <- img
		return false
	}
}

// TODO(slimsag): move to internal/glc
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

// TODO(slimsag): move to internal/glc
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

func (r *device) freeTextures() {
	// Lock the list.
	r.texturesToFree.Lock()

	if len(r.texturesToFree.slice) > 0 {
		// Free the textures.
		gl.DeleteTextures(int32(len(r.texturesToFree.slice)), &r.texturesToFree.slice[0])

		// Flush and execute OpenGL commands.
		gl.Flush()
		//gl.Execute()
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

// TODO(slimsag): move to internal/glc ?
func convertTexFormat(f gfx.TexFormat) int32 {
	switch f {
	case gfx.RGBA:
		return gl.RGBA8
	case gfx.RGB:
		return gl.RGB8
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

// TODO(slimsag): move to internal/glc ?
func unconvertTexFormat(f int32) gfx.TexFormat {
	switch f {
	case gl.RGBA8:
		return gfx.RGBA
	case gl.RGB8:
		return gfx.RGB
	case glCOMPRESSED_RGB_S3TC_DXT1_EXT:
		return gfx.DXT1
	case glCOMPRESSED_RGBA_S3TC_DXT1_EXT:
		return gfx.DXT1RGBA
	case glCOMPRESSED_RGBA_S3TC_DXT3_EXT:
		return gfx.DXT3
	case glCOMPRESSED_RGBA_S3TC_DXT5_EXT:
		return gfx.DXT5
	}
	panic("unknown format")
}

// LoadTexture implements the gfx.Renderer interface.
func (r *device) LoadTexture(t *gfx.Texture, done chan *gfx.Texture) {
	// If we are sharing assets with another renderer, allow it to load the
	// texture instead.
	r.shared.RLock()
	if r.shared.device != nil {
		r.shared.device.LoadTexture(t, done)
		r.shared.RUnlock()
		return
	}
	r.shared.RUnlock()

	if !t.Loaded && t.Source == nil {
		panic("LoadTexture(): Texture has a nil source!")
	}
	if t.Loaded {
		// Texture is already loaded, signal completion if needed and return.
		select {
		case done <- t:
		default:
		}
		return
	}

	// Prepare the image for uploading.
	src := prepareImage(r.devInfo.NPOT, t.Source)

	r.renderExec <- func() bool {
		// Determine appropriate internal image format.
		targetFormat := convertTexFormat(t.Format)
		internalFormat := int32(gl.RGBA)
		for _, format := range r.compressedTextureFormats {
			if format == targetFormat {
				internalFormat = format
				break
			}
		}

		// Initialize native texture.
		bounds := src.Bounds()
		native := newNativeTexture(
			r,
			internalFormat,
			bounds.Dx(),
			bounds.Dy(),
		)

		if t.MinFilter.Mipmapped() {
			gl.TexParameteri(gl.TEXTURE_2D, gl.GENERATE_MIPMAP, int32(gl.TRUE))
		}

		// Upload the image.
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			internalFormat,
			int32(bounds.Dx()),
			int32(bounds.Dy()),
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			unsafe.Pointer(&src.Pix[0]),
		)

		// Unbind texture to avoid carrying OpenGL state.
		gl.BindTexture(gl.TEXTURE_2D, 0)

		// Flush, Finish and execute OpenGL commands.
		gl.Flush()
		// Use Finish() to avoid accessing the texture before upload has completed, see:
		//  http://higherorderfun.com/blog/2011/05/26/multi-thread-opengl-texture-loading/
		gl.Finish()
		//gl.Execute()

		// Mark the texture as loaded.
		t.Loaded = true
		t.NativeTexture = native
		t.ClearData()

		// Attach a finalizer to the texture that will later free it.
		runtime.SetFinalizer(native, finalizeTexture)

		// Signal completion and return.
		select {
		case done <- t:
		default:
		}
		return false // no frame rendered.
	}
}
