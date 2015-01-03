// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"runtime"
	"strings"
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

func (r *device) freeShaders() {
	// Lock the list.
	r.rsrcManager.Lock()

	// Free the shaders.
	for _, native := range r.rsrcManager.shaders {
		// Delete shader objects (in practice we should be able to do this
		// directly after linking, but it would just leave the driver to
		// reference count anyway).
		gl.DeleteShader(native.vertex)
		gl.DeleteShader(native.fragment)

		// Delete program.
		gl.DeleteProgram(native.program)

		// Flush and execute OpenGL commands.
		gl.Flush()
		//gl.Execute()
	}

	// Slice to zero, and unlock.
	r.rsrcManager.shaders = r.rsrcManager.shaders[:0]
	r.rsrcManager.Unlock()
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

	if s.Loaded || len(s.Error) > 0 {
		// Shader is already loaded or there was an error loading, signal
		// completion if needed and return.
		select {
		case done <- s:
		default:
		}
		return
	}

	r.renderExec <- func() bool {
		shaderCompilerLog := func(s uint32) (log []byte, compiled bool) {
			var (
				ok, logSize int32
			)
			gl.GetShaderiv(s, gl.COMPILE_STATUS, &ok)
			//gl.Execute()

			// Shader compiler error
			gl.GetShaderiv(s, gl.INFO_LOG_LENGTH, &logSize)
			//gl.Execute()

			if logSize > 0 {
				log = make([]byte, logSize)
				gl.GetShaderInfoLog(s, int32(logSize), nil, (*int8)(unsafe.Pointer(&log[0])))
				//gl.Execute()

				// Strip null-termination byte.
				if log[len(log)-1] == 0 {
					log = log[:len(log)-1]
				}
			}
			return log, ok == 1
		}

		native := &nativeShader{
			r: r.rsrcManager,
		}

		// Handle the vertex shader now.
		if len(strings.TrimSpace(string(s.GLSL.Vertex))) == 0 {
			// No source code in vertex shader (some drivers will crash in
			// this case).
			s.Error = append(s.Error, []byte(s.Name+" | Vertex shader with no source code.\n")...)

			// Log the error.
			r.logf("%s | Vertex shader with no source code.\n", s.Name)
		} else {
			// Compile vertex shader.
			native.vertex = gl.CreateShader(gl.VERTEX_SHADER)
			lengths := int32(len(s.GLSL.Vertex))
			sources := &s.GLSL.Vertex[0]
			gl.ShaderSource(native.vertex, 1, (**int8)(unsafe.Pointer(&sources)), &lengths)
			gl.CompileShader(native.vertex)
			//gl.Execute()

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
		}

		// Handle the fragment shader now.
		if len(strings.TrimSpace(string(s.GLSL.Fragment))) == 0 {
			// No source code in fragment shader (some drivers will crash in
			// this case).
			s.Error = append(s.Error, []byte(s.Name+" | Fragment shader with no source code.\n")...)

			// Log the error.
			r.logf("%s | Fragment shader with no source code.\n", s.Name)
		} else {
			// Compile fragment shader.
			native.fragment = gl.CreateShader(gl.FRAGMENT_SHADER)
			lengths := int32(len(s.GLSL.Fragment))
			sources := &s.GLSL.Fragment[0]
			gl.ShaderSource(native.fragment, 1, (**int8)(unsafe.Pointer(&sources)), &lengths)
			gl.CompileShader(native.fragment)
			//gl.Execute()

			// Check if the shader compiled or not.
			log, compiled := shaderCompilerLog(native.fragment)
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
			//gl.Execute()

			if logSize > 0 {
				log = make([]byte, logSize)
				gl.GetProgramInfoLog(native.program, int32(logSize), nil, (*int8)(unsafe.Pointer(&log[0])))
				//gl.Execute()

				// Strip null-termination byte.
				if log[len(log)-1] == 0 {
					log = log[:len(log)-1]
				}
			}

			// Check for linker errors.
			var ok int32
			gl.GetProgramiv(native.program, gl.LINK_STATUS, &ok)
			//gl.Execute()
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

		// Flush and execute OpenGL commands.
		gl.Flush()
		//gl.Execute()

		// Signal completion and return.
		select {
		case done <- s:
		default:
		}
		return false // no frame rendered.
	}
}
