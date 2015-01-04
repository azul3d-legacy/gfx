// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"runtime"
	"unsafe"

	"azul3d.org/gfx.v2-dev"
	"azul3d.org/gfx.v2-dev/internal/gl/2.0/gl"
	"azul3d.org/gfx.v2-dev/internal/glutil"
)

// nativeShader is stored inside the *Shader.Native interface and stores GLSL
// shader IDs.
type nativeShader struct {
	*glutil.LocationCache
	program, vertex, fragment uint32
	r                         *rsrcManager
}

func finalizeShader(n *nativeShader) {
	n.r.Lock()
	n.r.shaders = append(n.r.shaders, n)
	n.r.Unlock()
}

// Implements gfx.Destroyable interface.
func (n *nativeShader) Destroy() {
	finalizeShader(n)
}

func (r *rsrcManager) freeShaders() {
	// Lock the list.
	r.Lock()

	// Free the shaders.
	for _, native := range r.shaders {
		// Delete shader objects (in practice we should be able to do this
		// directly after linking, but it would just leave the driver to
		// reference count anyway).
		gl.DeleteShader(native.vertex)
		gl.DeleteShader(native.fragment)

		// Delete program.
		gl.DeleteProgram(native.program)

		// Flush OpenGL commands.
		gl.Flush()
	}

	// Slice to zero, and unlock.
	r.shaders = r.shaders[:0]
	r.Unlock()
}

func shaderCompilerLog(s uint32) (log []byte, compiled bool) {
	var (
		ok, logSize int32
	)
	gl.GetShaderiv(s, gl.COMPILE_STATUS, &ok)

	// Shader compiler error
	gl.GetShaderiv(s, gl.INFO_LOG_LENGTH, &logSize)

	if logSize > 0 {
		log = make([]byte, logSize)
		gl.GetShaderInfoLog(s, int32(logSize), nil, (*uint8)(unsafe.Pointer(&log[0])))

		// Strip null-termination byte.
		if log[len(log)-1] == 0 {
			log = log[:len(log)-1]
		}
	}
	return log, ok == 1
}

// LoadShader implements the gfx.Renderer interface.
func (r *device) LoadShader(s *gfx.Shader, done chan *gfx.Shader) {
	// If we are sharing assets with another renderer, allow it to load the
	// shader instead.
	r.shared.RLock()
	if r.shared.device != nil {
		r.shared.device.LoadShader(s, done)
		r.shared.RUnlock()
		return
	}
	r.shared.RUnlock()

	doLoad, err := glutil.PreLoadShader(s, done)
	if err != nil {
		r.logf("%v\n", err)
		return
	}
	if !doLoad {
		return
	}

	r.renderExec <- func() bool {
		native := &nativeShader{
			r: r.rsrcManager,
		}

		// Compile vertex shader.
		native.vertex = gl.CreateShader(gl.VERTEX_SHADER)
		lengths := int32(len(s.GLSL.Vertex))
		sources := &s.GLSL.Vertex[0]
		gl.ShaderSource(native.vertex, 1, (**uint8)(unsafe.Pointer(&sources)), &lengths)
		gl.CompileShader(native.vertex)

		// Check if the shader compiled or not.
		log, compiled := shaderCompilerLog(native.vertex)
		if !compiled {
			// Just for sanity.
			native.vertex = 0

			// Append the errors.
			s.Error = append(s.Error, []byte(s.Name+" | Vertex shader errors:\n")...)
			s.Error = append(s.Error, log...)
		}
		if len(log) > 0 {
			// Send the compiler log to the debug writer.
			r.logf("%s | Vertex shader errors:\n", s.Name)
			r.logf(string(log))
		}

		// Compile fragment shader.
		native.fragment = gl.CreateShader(gl.FRAGMENT_SHADER)
		lengths = int32(len(s.GLSL.Fragment))
		sources = &s.GLSL.Fragment[0]
		gl.ShaderSource(native.fragment, 1, (**uint8)(unsafe.Pointer(&sources)), &lengths)
		gl.CompileShader(native.fragment)

		// Check if the shader compiled or not.
		log, compiled = shaderCompilerLog(native.fragment)
		if !compiled {
			// Just for sanity.
			native.fragment = 0

			// Append the errors.
			s.Error = append(s.Error, []byte(s.Name+" | Fragment shader errors:\n")...)
			s.Error = append(s.Error, log...)
		}
		if len(log) > 0 {
			// Send the compiler log to the debug writer.
			r.logf("%s | Fragment shader errors:\n", s.Name)
			r.logf(string(log))
		}

		// Create the shader program if all went well with the vertex and
		// fragment shaders.
		if native.vertex != 0 && native.fragment != 0 {
			native.program = gl.CreateProgram()
			gl.AttachShader(native.program, native.vertex)
			gl.AttachShader(native.program, native.fragment)
			gl.LinkProgram(native.program)

			// Grab the linker's log.
			var (
				logSize int32
				log     []byte
			)
			gl.GetProgramiv(native.program, gl.INFO_LOG_LENGTH, &logSize)

			if logSize > 0 {
				log = make([]byte, logSize)
				gl.GetProgramInfoLog(native.program, int32(logSize), nil, (*uint8)(unsafe.Pointer(&log[0])))

				// Strip null-termination byte.
				if log[len(log)-1] == 0 {
					log = log[:len(log)-1]
				}
			}

			// Check for linker errors.
			var ok int32
			gl.GetProgramiv(native.program, gl.LINK_STATUS, &ok)
			if ok == 0 {
				// Just for sanity.
				native.program = 0

				// Append the errors.
				s.Error = append(s.Error, []byte(s.Name+" | Linker errors:\n")...)
				s.Error = append(s.Error, log...)
			}
			if len(log) > 0 {
				// Send the linker log to the debug writer.
				r.logf("%s | Linker errors:\n", s.Name)
				r.logf(string(log))
			}
		}

		// Mark the shader as loaded if there were no errors.
		if len(s.Error) == 0 {
			native.LocationCache = &glutil.LocationCache{
				GetAttribLocation: func(name string) int {
					return int(gl.GetAttribLocation(native.program, glStr(name)))
				},
				GetUniformLocation: func(name string) int {
					return int(gl.GetUniformLocation(native.program, glStr(name)))
				},
			}

			s.Loaded = true
			s.NativeShader = native
			s.ClearData()

			// Attach a finalizer to the shader that will later free it.
			runtime.SetFinalizer(native, finalizeShader)
		}

		// Flush OpenGL commands.
		gl.Flush()

		// Signal completion and return.
		select {
		case done <- s:
		default:
		}
		return false // no frame rendered.
	}
}
