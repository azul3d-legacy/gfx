// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"azul3d.org/gfx.v1"
	"azul3d.org/lmath.v1"
	"azul3d.org/native/gl.v1"
	"fmt"
	"image"
)

var (
	// Get an matrix which will translate our matrix from ZUpRight to YUpRight
	zUpRightToYUpRight = lmath.CoordSysZUpRight.ConvertMat4(lmath.CoordSysYUpRight)
)

var (
	textureNames  = make([]string, 32)
	texCoordNames = make([]string, 32)
)

func init() {
	for i := 0; i < len(textureNames); i++ {
		textureNames[i] = fmt.Sprintf("Texture%d", i)
	}

	for i := 0; i < len(texCoordNames); i++ {
		texCoordNames[i] = fmt.Sprintf("TexCoord%d", i)
	}
}

func textureName(i int) string {
	if i < len(textureNames) {
		return textureNames[i]
	}
	n := fmt.Sprintf("Texture%d", i)
	textureNames = append(textureNames, n)
	return n
}

func texCoordName(i int) string {
	if i < len(texCoordNames) {
		return texCoordNames[i]
	}
	n := fmt.Sprintf("TexCoord%d", i)
	texCoordNames = append(texCoordNames, n)
	return n
}

// Used as the *gfx.Object.NativeObject interface value.
type nativeObject struct {
	// The graphics object's last-known transform, if they are not equal then
	// the matrices must be recalculated.
	Transform lmath.Mat4

	// The last-known camera transform and projection.
	CameraTransform lmath.Mat4
	Projection      gfx.Mat4

	// Cached pre-calculated matrices to feed into shaders, this way we don't
	// recalculate matrices every single frame but instead only when they
	// actually change.
	model, view, projection, mvp gfx.Mat4

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

func (n nativeObject) needRebuild(o *gfx.Object, c *gfx.Camera) bool {
	if o.Transform.Mat4() != n.Transform {
		return true
	}
	if c.Object.Transform.Mat4() != n.CameraTransform {
		return true
	}
	if c.Projection != n.Projection {
		return true
	}
	return false
}

func (n nativeObject) rebuild(o *gfx.Object, c *gfx.Camera) nativeObject {
	// The "Model" matrix is the Object's transformation matrix, we feed it
	// directly in.
	n.model = gfx.ConvertMat4(o.Transform.Mat4())

	// The "View" matrix is the coordinate system conversion, multiplied
	// against the camera object's transformation matrix
	view := zUpRightToYUpRight
	if c != nil {
		// Apply inverse of camera object transformation.
		camInverse, _ := c.Object.Transform.Mat4().Inverse()
		view = camInverse.Mul(view)
	}
	n.view = gfx.ConvertMat4(view)

	// The "Projection" matrix is the camera's projection matrix.
	projection := lmath.Mat4Identity
	if c != nil {
		projection = c.Projection.Mat4()
	}
	n.projection = gfx.ConvertMat4(projection)

	// The "MVP" matrix is Model * View * Projection matrix.
	mvp := o.Transform.Mat4()
	mvp = mvp.Mul(view)
	mvp = mvp.Mul(projection)
	n.mvp = gfx.ConvertMat4(mvp)
	return n
}

// Implements gfx.Renderer interface.
func (r *Renderer) Draw(rect image.Rectangle, o *gfx.Object, c *gfx.Camera) {
	lock := func() {
		o.Lock()
		if c != nil {
			c.Lock()
		}
	}

	unlock := func() {
		o.Unlock()
		if c != nil {
			c.Unlock()
		}
	}

	// Lock the object until we are completely done drawing it.
	lock()

	var (
		shaderLoaded   chan *gfx.Shader
		meshesLoaded   []chan *gfx.Mesh
		texturesLoaded []chan *gfx.Texture
	)

	// Begin loading shader.
	if o.Shader == nil {
		// Can't draw.
		unlock()
		return
	}
	o.Shader.RLock()
	shaderNeedLoad := !o.Shader.Loaded
	shaderHasError := len(o.Shader.Error) > 0
	o.Shader.RUnlock()
	if shaderHasError {
		// Can't draw.
		unlock()
		return
	}
	if shaderNeedLoad {
		shaderLoaded = make(chan *gfx.Shader, 1)
		r.LoadShader(o.Shader, shaderLoaded)
	}

	// Begin loading meshes.
	if len(o.Meshes) == 0 {
		// Can't draw.
		unlock()
		return
	}

	for _, m := range o.Meshes {
		m.RLock()
		meshNeedLoad := !m.Loaded || m.HasChanged()
		meshEmpty := !m.Loaded && len(m.Vertices) == 0
		m.RUnlock()
		if meshEmpty {
			// Can't draw.
			unlock()
			return
		}
		if meshNeedLoad {
			ch := make(chan *gfx.Mesh, 1)
			r.LoadMesh(m, ch)
			meshesLoaded = append(meshesLoaded, ch)
		}
	}

	// Begin loading textures.
	for _, t := range o.Textures {
		t.RLock()
		texNeedLoad := !t.Loaded
		t.RUnlock()
		if texNeedLoad {
			ch := make(chan *gfx.Texture, 1)
			r.LoadTexture(t, ch)
			texturesLoaded = append(texturesLoaded, ch)
		}
	}

	// Wait for shader, meshes, and textures to finish loading.
	if shaderLoaded != nil {
		<-shaderLoaded
	}
	for _, load := range meshesLoaded {
		<-load
	}
	for _, load := range texturesLoaded {
		<-load
	}

	// Check if the now-loaded shader might have errors.
	o.Shader.RLock()
	shaderHasError = len(o.Shader.Error) > 0
	o.Shader.RUnlock()
	if shaderHasError {
		// Can't draw.
		unlock()
		return
	}

	// Must set at least an empty native object before Draw() returns.
	o.NativeObject = nativeObject{}

	// Ask the render loop to perform drawing.
	r.RenderExec <- func() bool {
		// Set global GL state.
		r.setGlobalState()

		// Update the scissor region (effects drawing).
		r.stateScissor(r.render, r.Bounds(), rect)

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

		// Unlock the object now that we are done drawing it.
		unlock()

		// Yield for occlusion query results, if any are available.
		r.queryYield()

		return false
	}
}

func (r *Renderer) findAttribLocation(native *nativeShader, name string) (uint32, bool) {
	location, ok := native.attribLookup[name]
	if ok {
		return uint32(location), true
	}
	bts := []byte(name)
	bts = append(bts, 0)
	location = r.render.GetAttribLocation(native.program, &bts[0])
	if location < 0 {
		return 0, false
	}
	return uint32(location), true
}

func (r *Renderer) findUniformLocation(native *nativeShader, name string) int32 {
	location, ok := native.uniformLookup[name]
	if ok {
		return location
	}
	bts := []byte(name)
	bts = append(bts, 0)
	location = r.render.GetUniformLocation(native.program, &bts[0])
	if location < 0 {
		// Just for sanity.
		return -1
	}
	return location
}

type texSlot int32

func (r *Renderer) updateUniform(native *nativeShader, name string, value interface{}) {
	location := r.findUniformLocation(native, name)
	if location == -1 {
		// The uniform is not used by the shader program and should just be
		// dropped.
		return
	}

	switch v := value.(type) {
	case texSlot:
		// Special case: Texture input uniform.
		r.render.Uniform1i(location, int32(v))

	case bool:
		var intBool int32
		if v {
			intBool = 1
		}
		r.render.Uniform1iv(location, 1, &intBool)

	case float32:
		r.render.Uniform1fv(location, 1, &v)

	case []float32:
		if len(v) > 0 {
			r.render.Uniform1fv(location, uint32(len(v)), &v[0])
		}

	case gfx.Vec3:
		r.render.Uniform3fv(location, 1, &v.X)

	case []gfx.Vec3:
		if len(v) > 0 {
			r.render.Uniform3fv(location, uint32(len(v)), &v[0].X)
		}

	case gfx.Mat4:
		r.render.UniformMatrix4fv(location, 1, gl.GLBool(false), &v[0][0])

	case []gfx.Mat4:
		if len(v) > 0 {
			r.render.UniformMatrix4fv(location, uint32(len(v)), gl.GLBool(false), &v[0][0][0])
		}

	default:
		// We don't know of the type at all, ignore it.
	}
}

func (r *Renderer) beginQuery(o *gfx.Object, n nativeObject) nativeObject {
	if r.glArbOcclusionQuery && o.OcclusionTest {
		r.render.GenQueries(1, &n.pendingQuery)
		r.render.Execute()
		r.render.BeginQuery(gl.SAMPLES_PASSED, n.pendingQuery)
		r.render.Execute()

		// Add the pending query.
		r.pending.Lock()
		r.pending.queries = append(r.pending.queries, pendingQuery{n.pendingQuery, o})
		r.pending.Unlock()
	}
	return n
}

func (r *Renderer) endQuery(o *gfx.Object, n nativeObject) nativeObject {
	if r.glArbOcclusionQuery && o.OcclusionTest {
		r.render.EndQuery(gl.SAMPLES_PASSED)
		r.render.Execute()
	}
	return n
}

func (r *Renderer) useState(ns *nativeShader, obj *gfx.Object, c *gfx.Camera) {
	// Use object state.
	r.stateColorWrite(r.render, [4]bool{obj.WriteRed, obj.WriteGreen, obj.WriteBlue, obj.WriteAlpha})
	r.stateDithering(r.render, obj.Dithering)
	r.stateStencilTest(r.render, obj.StencilTest)
	r.stateStencilOp(r.render, obj.StencilFront, obj.StencilBack)
	r.stateStencilFunc(r.render, obj.StencilFront, obj.StencilBack)
	r.stateStencilMask(r.render, obj.StencilFront.WriteMask, obj.StencilBack.WriteMask)
	r.stateDepthFunc(r.render, obj.DepthCmp)
	r.stateDepthTest(r.render, obj.DepthTest)
	r.stateDepthWrite(r.render, obj.DepthWrite)
	r.stateFaceCulling(r.render, obj.FaceCulling)

	// Begin using the shader.
	shader := obj.Shader
	if r.lastShader != shader {
		r.lastShader = shader

		r.stateProgram(r.render, ns.program)

		// Update shader inputs.
		for name := range shader.Inputs {
			value := shader.Inputs[name]
			r.updateUniform(ns, name, value)
		}
	}

	// Consider rebuilding the object's cached matrices, if needed.
	nativeObj := obj.NativeObject.(nativeObject)
	if nativeObj.needRebuild(obj, c) {
		// Rebuild cached matrices.
		nativeObj = nativeObj.rebuild(obj, c)
	}
	obj.NativeObject = nativeObj

	// Add the matrix inputs for the object.
	r.updateUniform(ns, "Model", nativeObj.model)
	r.updateUniform(ns, "View", nativeObj.view)
	r.updateUniform(ns, "Projection", nativeObj.projection)
	r.updateUniform(ns, "MVP", nativeObj.mvp)

	// Set alpha mode.
	r.stateAlphaToCoverage(r.render, &r.gpuInfo, obj.AlphaMode == gfx.AlphaToCoverage)
	r.stateBlend(r.render, obj.AlphaMode == gfx.AlphaBlend)
	if obj.AlphaMode == gfx.AlphaBlend {
		r.stateBlendColor(r.render, obj.Blend.Color)
		r.stateBlendFuncSeparate(r.render, obj.Blend)
		r.stateBlendEquationSeparate(r.render, obj.Blend)
	}

	switch obj.AlphaMode {
	case gfx.NoAlpha, gfx.AlphaBlend:
		r.updateUniform(ns, "BinaryAlpha", false)

	case gfx.BinaryAlpha, gfx.AlphaToCoverage:
		r.updateUniform(ns, "BinaryAlpha", true)
	}

	// Bind each texture.
	for i, t := range obj.Textures {
		nt := t.NativeTexture.(*nativeTexture)

		r.render.ActiveTexture(gl.TEXTURE0 + int32(i))
		r.render.BindTexture(gl.TEXTURE_2D, nt.id)

		// Load wrap mode.
		uWrap := convertWrap(t.WrapU)
		vWrap := convertWrap(t.WrapV)
		if t.WrapU == gfx.BorderColor || t.WrapV == gfx.BorderColor {
			// We must specify the actual border color then.
			r.render.TexParameterfv(gl.TEXTURE_2D, gl.TEXTURE_BORDER_COLOR, &t.BorderColor.R)
			r.render.Execute()
		}
		r.render.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, uWrap)
		r.render.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, vWrap)

		// Load filter.
		r.render.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, convertFilter(t.MinFilter))
		r.render.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, convertFilter(t.MagFilter))

		// Enable mipmap generation if either filter is mipmapped.
		if t.MinFilter.Mipmapped() || t.MagFilter.Mipmapped() {
			r.render.TexParameteri(gl.TEXTURE_2D, gl.GENERATE_MIPMAP, int32(gl.TRUE))
		}

		// Add uniform input.
		r.updateUniform(ns, textureName(i), texSlot(i))
	}

	// Begin occlusion query.
	obj.NativeObject = r.beginQuery(obj, nativeObj)
}

func (r *Renderer) clearState(ns *nativeShader, obj *gfx.Object) {
	// End occlusion query.
	obj.NativeObject = r.endQuery(obj, obj.NativeObject.(nativeObject))

	// Use no texture.
	r.render.BindTexture(gl.TEXTURE_2D, 0)
	r.render.ActiveTexture(gl.TEXTURE0)
}

func (r *Renderer) drawMesh(ns *nativeShader, m *gfx.Mesh) {
	// Grab the native mesh.
	native := m.NativeMesh.(*nativeMesh)

	// Use vertices data.
	location, ok := r.findAttribLocation(ns, "Vertex")
	if ok {
		r.render.BindBuffer(gl.ARRAY_BUFFER, native.vertices)
		r.render.EnableVertexAttribArray(location)
		defer r.render.DisableVertexAttribArray(location)
		r.render.VertexAttribPointer(location, 3, gl.FLOAT, gl.GLBool(false), 0, nil)
	}

	if native.colors != 0 {
		// Use colors data.
		location, ok = r.findAttribLocation(ns, "Color")
		if ok {
			r.render.BindBuffer(gl.ARRAY_BUFFER, native.colors)
			r.render.EnableVertexAttribArray(location)
			defer r.render.DisableVertexAttribArray(location)
			r.render.VertexAttribPointer(location, 4, gl.FLOAT, gl.GLBool(false), 0, nil)
		}
	}

	if native.bary != 0 {
		// Use bary data.
		location, ok = r.findAttribLocation(ns, "Bary")
		if ok {
			r.render.BindBuffer(gl.ARRAY_BUFFER, native.bary)
			r.render.EnableVertexAttribArray(location)
			defer r.render.DisableVertexAttribArray(location)
			r.render.VertexAttribPointer(location, 3, gl.FLOAT, gl.GLBool(false), 0, nil)
		}
	}

	// Use each texture coordinate set data.
	for index, texCoords := range native.texCoords {
		name := texCoordName(index)
		location, ok = r.findAttribLocation(ns, name)
		if ok {
			r.render.BindBuffer(gl.ARRAY_BUFFER, texCoords)
			r.render.EnableVertexAttribArray(location)
			defer r.render.DisableVertexAttribArray(location)
			r.render.VertexAttribPointer(location, 2, gl.FLOAT, gl.GLBool(false), 0, nil)
		}
	}

	if native.indicesCount > 0 {
		// Draw indexed mesh.
		r.render.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, native.indices)
		r.render.DrawElements(gl.TRIANGLES, native.indicesCount, gl.UNSIGNED_INT, nil)
	} else {
		// Draw regular mesh.
		r.render.DrawArrays(gl.TRIANGLES, 0, native.verticesCount)
	}

	// Unbind buffer to avoid carrying OpenGL state.
	r.render.BindBuffer(gl.ARRAY_BUFFER, 0)
}
