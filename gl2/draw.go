// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"fmt"
	"image"
	"reflect"

	"azul3d.org/gfx.v2-dev"
	"azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"
	"azul3d.org/gfx.v2-dev/internal/glutil"
	"azul3d.org/gfx.v2-dev/internal/util"
)

var (
	textureIndex  = glutil.NewIndexStr("Texture")
	texCoordIndex = glutil.NewIndexStr("TexCoord")
)

// Used as the *gfx.Object.NativeObject interface value.
type nativeObject struct {
	*glutil.MVPCache

	// The pending occlusion query ID.
	pendingQuery uint32

	// The sample count of the object the last time it was drawn.
	sampleCount int
}

// Implements the gfx.NativeObject interface.
func (n nativeObject) SampleCount() int {
	return n.sampleCount
}

// Implements the gfx.Destroyable interface.
func (n nativeObject) Destroy() {}

func (r *device) hookedDraw(rect image.Rectangle, o *gfx.Object, c *gfx.Camera, pre, post func()) {
	doDraw, err := util.PreDraw(r, rect, o, c)
	if err != nil {
		r.logf("%v\n", err)
		return
	}
	if !doDraw {
		return
	}

	// Ask the render loop to perform drawing.
	r.renderExec <- func() bool {
		// Give the object a native object.
		o.NativeObject = nativeObject{
			MVPCache: &glutil.MVPCache{},
		}

		if pre != nil {
			pre()
		}

		// Set global GL state.
		r.graphicsState.Begin(r)

		// Update the scissor region (effects drawing).
		r.performScissor(rect)

		var ns *nativeShader
		if o.NativeShader != nil {
			ns = o.NativeShader.(*nativeShader)
		}

		// Use the object's state.
		r.useState(ns, o, c)

		// Draw each mesh.
		for _, m := range o.Meshes {
			r.drawMesh(ns, m)
		}

		// Clear the object's state.
		r.clearState(ns, o)

		// Yield for occlusion query results, if any are available.
		r.queryYield()

		if post != nil {
			post()
		}
		return false
	}
}

type texSlot int32

func (r *device) updateUniform(native *nativeShader, name string, value interface{}) {
	location := int32(native.LocationCache.FindUniform(name))
	if location == -1 {
		// The uniform is not used by the shader program and should just be
		// dropped.
		return
	}

	switch v := value.(type) {
	case texSlot:
		// Special case: Texture input uniform.
		gl.Uniform1i(location, int32(v))

	case bool:
		var intBool int32
		if v {
			intBool = 1
		}
		gl.Uniform1iv(location, 1, &intBool)

	case float32:
		gl.Uniform1fv(location, 1, &v)

	case []float32:
		if len(v) > 0 {
			gl.Uniform1fv(location, int32(len(v)), &v[0])
		}

	case gfx.TexCoord:
		gl.Uniform2fv(location, 1, &v.U)

	case []gfx.TexCoord:
		if len(v) > 0 {
			gl.Uniform2fv(location, int32(len(v)), &v[0].U)
		}

	case gfx.Vec3:
		gl.Uniform3fv(location, 1, &v.X)

	case []gfx.Vec3:
		if len(v) > 0 {
			gl.Uniform3fv(location, int32(len(v)), &v[0].X)
		}

	case gfx.Vec4:
		gl.Uniform4fv(location, 1, &v.X)

	case []gfx.Vec4:
		if len(v) > 0 {
			gl.Uniform4fv(location, int32(len(v)), &v[0].X)
		}

	case gfx.Color:
		gl.Uniform4fv(location, 1, &v.R)

	case []gfx.Color:
		if len(v) > 0 {
			gl.Uniform4fv(location, int32(len(v)), &v[0].R)
		}

	case gfx.Mat4:
		gl.UniformMatrix4fv(location, 1, false, &v[0][0])

	case []gfx.Mat4:
		if len(v) > 0 {
			gl.UniformMatrix4fv(location, int32(len(v)), false, &v[0][0][0])
		}

	default:
		r.logf("Shader input %q uses an invalid shader input data type %q, ignoring.\n", name, reflect.TypeOf(value))
		// We don't know of the type at all, ignore it.
	}
}

func (r *device) beginQuery(o *gfx.Object, n nativeObject) nativeObject {
	if r.glArbOcclusionQuery && o.OcclusionTest {
		gl.GenQueries(1, &n.pendingQuery)
		//gl.Execute()
		gl.BeginQuery(gl.SAMPLES_PASSED, n.pendingQuery)
		//gl.Execute()

		// Add the pending query.
		r.pending.Lock()
		r.pending.queries = append(r.pending.queries, pendingQuery{n.pendingQuery, o})
		r.pending.Unlock()
	}
	return n
}

func (r *device) endQuery(o *gfx.Object, n nativeObject) nativeObject {
	if r.glArbOcclusionQuery && o.OcclusionTest {
		gl.EndQuery(gl.SAMPLES_PASSED)
		//gl.Execute()
	}
	return n
}

func (r *device) useState(ns *nativeShader, obj *gfx.Object, c *gfx.Camera) {
	// Use object state.
	r.graphicsState.ColorWrite(obj.WriteRed, obj.WriteGreen, obj.WriteBlue, obj.WriteAlpha)
	r.graphicsState.Dithering(obj.Dithering)
	r.graphicsState.StencilTest(obj.StencilTest)
	r.graphicsState.StencilOpSeparate(obj.StencilFront, obj.StencilBack)
	r.graphicsState.stencilFuncSeparate(obj.StencilFront, obj.StencilBack)
	r.graphicsState.stencilMaskSeparate(obj.StencilFront.WriteMask, obj.StencilBack.WriteMask)
	if r.gpuInfo.DepthClamp {
		r.graphicsState.depthClamp(obj.DepthClamp)
	}
	r.graphicsState.DepthCmp(obj.DepthCmp)
	r.graphicsState.DepthTest(obj.DepthTest)
	r.graphicsState.DepthWrite(obj.DepthWrite)
	r.graphicsState.FaceCulling(obj.FaceCulling)

	// Begin using the shader.
	shader := obj.Shader
	r.graphicsState.useProgram(ns.program)

	// Update shader inputs.
	for name := range shader.Inputs {
		value := shader.Inputs[name]
		r.updateUniform(ns, name, value)
	}

	// Update the object's MVP cache, if needed.
	nativeObj := obj.NativeObject.(nativeObject)
	nativeObj.MVPCache.Update(obj, c)

	// Add the matrix inputs for the object.
	r.updateUniform(ns, "Model", nativeObj.MVPCache.Model)
	r.updateUniform(ns, "View", nativeObj.MVPCache.View)
	r.updateUniform(ns, "Projection", nativeObj.MVPCache.Projection)
	r.updateUniform(ns, "MVP", nativeObj.MVPCache.MVP)

	// Set alpha mode.
	if r.gpuInfo.AlphaToCoverage {
		r.graphicsState.SampleAlphaToCoverage(obj.AlphaMode == gfx.AlphaToCoverage)
	}
	r.graphicsState.Blend(obj.AlphaMode == gfx.AlphaBlend)
	if obj.AlphaMode == gfx.AlphaBlend {
		r.graphicsState.BlendColor(obj.Blend.Color)
		r.graphicsState.BlendFuncSeparate(obj.Blend)
		r.graphicsState.BlendEquationSeparate(obj.Blend)
	}

	switch obj.AlphaMode {
	case gfx.NoAlpha, gfx.AlphaBlend:
		r.updateUniform(ns, "BinaryAlpha", false)

	case gfx.BinaryAlpha, gfx.AlphaToCoverage:
		r.updateUniform(ns, "BinaryAlpha", true)
	}

	// Bind each texture.
	for i, t := range obj.Textures {
		// Ensure there are no feedback loops if we are rendering to a texture.
		if r.rttCanvas != nil {
			cfg := r.rttCanvas.cfg
			color := cfg.Color.NativeTexture
			depth := cfg.Color.NativeTexture
			stencil := cfg.Color.NativeTexture
			native := t.NativeTexture
			if native != nil && (native == color || native == depth || native == stencil) {
				panic("Feedback Loop - Object cannot use the texture that is being drawn to.")
			}
		}

		nt := t.NativeTexture.(*nativeTexture)

		gl.ActiveTexture(gl.TEXTURE0 + uint32(i))
		gl.BindTexture(gl.TEXTURE_2D, nt.id)

		// Load wrap mode.
		uWrap := convertWrap(t.WrapU)
		vWrap := convertWrap(t.WrapV)
		if t.WrapU == gfx.BorderColor || t.WrapV == gfx.BorderColor {
			// We must specify the actual border color then.
			gl.TexParameterfv(gl.TEXTURE_2D, gl.TEXTURE_BORDER_COLOR, &t.BorderColor.R)
			//gl.Execute()
		}
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, uWrap)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, vWrap)

		// Load filter.
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, convertFilter(t.MinFilter))
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, convertFilter(t.MagFilter))

		// If we do not want mipmapping, turn it off. Note that only the
		// minification filter can be mipmapped (mag filter can never be).
		if t.MinFilter.Mipmapped() {
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_BASE_LEVEL, 0)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAX_LEVEL, 1000)
		} else {
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_BASE_LEVEL, 0)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAX_LEVEL, 0)
		}

		// Add uniform input.
		r.updateUniform(ns, textureIndex.Name(i), texSlot(i))
	}

	// Begin occlusion query.
	obj.NativeObject = r.beginQuery(obj, nativeObj)
}

func (r *device) clearState(ns *nativeShader, obj *gfx.Object) {
	// End occlusion query.
	obj.NativeObject = r.endQuery(obj, obj.NativeObject.(nativeObject))

	// Use no texture.
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.ActiveTexture(gl.TEXTURE0)
}

func (r *device) drawMesh(ns *nativeShader, m *gfx.Mesh) {
	// Grab the native mesh.
	native := m.NativeMesh.(*nativeMesh)

	// Use vertices data.
	location := ns.LocationCache.FindAttrib("Vertex")
	if location != -1 {
		gl.BindBuffer(gl.ARRAY_BUFFER, native.vertices)
		gl.EnableVertexAttribArray(uint32(location))
		defer gl.DisableVertexAttribArray(uint32(location))
		gl.VertexAttribPointer(uint32(location), 3, gl.FLOAT, false, 0, nil)
	}

	// Use each texture coordinate set data.
	for index, texCoords := range native.texCoords {
		location = ns.LocationCache.FindAttrib(texCoordIndex.Name(index))
		if location != -1 {
			gl.BindBuffer(gl.ARRAY_BUFFER, texCoords)
			gl.EnableVertexAttribArray(uint32(location))
			defer gl.DisableVertexAttribArray(uint32(location))
			gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, nil)
		}
	}

	// Use each custom vertex data set.
	for name, attrib := range native.attribs {
		for i, vbo := range attrib.vbos {
			// Determine name.
			indexName := name
			if len(attrib.vbos) > 1 {
				indexName = fmt.Sprintf("%s%d", name, i)
			}

			// Find input location.
			location = ns.LocationCache.FindAttrib(indexName)
			if location == -1 {
				continue
			}

			// Bind the buffer, send each row.
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
			for row := uint32(0); row < attrib.rows; row++ {
				l := uint32(location) + row
				gl.EnableVertexAttribArray(l)
				defer gl.DisableVertexAttribArray(l)
				gl.VertexAttribPointer(l, attrib.size, gl.FLOAT, false, 0, nil)
			}
		}
	}

	if native.indicesCount > 0 {
		// Draw indexed mesh.
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, native.indices)
		gl.DrawElements(uint32(r.Common.ConvertPrimitive(m.Primitive)), native.indicesCount, gl.UNSIGNED_INT, nil)
	} else {
		// Draw regular mesh.
		gl.DrawArrays(uint32(r.Common.ConvertPrimitive(m.Primitive)), 0, native.verticesCount)
	}

	// Unbind buffer to avoid carrying OpenGL state.
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}
