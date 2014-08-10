// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"runtime"
	"strings"

	"azul3d.org/gfx.v1"
	"azul3d.org/native/gl.v1"
)

// nativeShader is stored inside the *Shader.Native interface and stores GLSL
// shader IDs.
type nativeShader struct {
	program, vertex, fragment   uint32
	attribLookup, uniformLookup map[string]int32
	r                           *Renderer
}

func finalizeShader(n *nativeShader) {
	n.r.shadersToFree.Lock()
	n.r.shadersToFree.slice = append(n.r.shadersToFree.slice, n)
	n.r.shadersToFree.Unlock()
}

// Implements gfx.Destroyable interface.
func (n *nativeShader) Destroy() {
	finalizeShader(n)
}

func (r *Renderer) freeShaders() {
	// Lock the list.
	r.shadersToFree.Lock()

	// Free the shaders.
	for _, native := range r.shadersToFree.slice {
		// Delete shader objects (in practice we should be able to do this
		// directly after linking, but it would just leave the driver to
		// reference count anyway).
		r.loader.DeleteShader(native.vertex)
		r.loader.DeleteShader(native.fragment)

		// Delete program.
		r.loader.DeleteProgram(native.program)

		// Flush and execute OpenGL commands.
		r.loader.Flush()
		r.loader.Execute()
	}

	// Slice to zero, and unlock.
	r.shadersToFree.slice = r.shadersToFree.slice[:0]
	r.shadersToFree.Unlock()
}

// Implements gfx.Renderer interface.
func (r *Renderer) LoadShader(s *gfx.Shader, done chan *gfx.Shader) {
	// Lock the shader until we are done loading it.
	s.Lock()
	if s.Loaded || len(s.Error) > 0 {
		// Shader is already loaded or there was an error loading, signal
		// completion if needed and return after unlocking.
		s.Unlock()
		select {
		case done <- s:
		default:
		}
		return
	}

	f := func() {
		shaderCompilerLog := func(s uint32) (log []byte, compiled bool) {
			var (
				ok, logSize int32
			)
			r.loader.GetShaderiv(s, gl.COMPILE_STATUS, &ok)
			r.loader.Execute()

			// Shader compiler error
			r.loader.GetShaderiv(s, gl.INFO_LOG_LENGTH, &logSize)
			r.loader.Execute()

			if logSize > 0 {
				log = make([]byte, logSize)
				r.loader.GetShaderInfoLog(s, uint32(logSize), nil, &log[0])
				r.loader.Execute()

				// Strip null-termination byte.
				if log[len(log)-1] == 0 {
					log = log[:len(log)-1]
				}
			}
			return log, ok == 1
		}

		native := &nativeShader{
			attribLookup:  make(map[string]int32, 8),
			uniformLookup: make(map[string]int32, 8),
			r:             r,
		}

		// Handle the vertex shader now.
		if len(strings.TrimSpace(string(s.GLSLVert))) == 0 {
			// No source code in vertex shader (some drivers will crash in
			// this case).
			s.Error = append(s.Error, []byte(s.Name+" | Vertex shader with no source code.\n")...)

			// Log the error.
			r.logf("%s | Vertex shader with no source code.\n", s.Name)
		} else {
			// Compile vertex shader.
			native.vertex = r.loader.CreateShader(gl.VERTEX_SHADER)
			lengths := int32(len(s.GLSLVert))
			sources := &s.GLSLVert[0]
			r.loader.ShaderSource(native.vertex, 1, &sources, &lengths)
			r.loader.CompileShader(native.vertex)
			r.loader.Execute()

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
		if len(strings.TrimSpace(string(s.GLSLFrag))) == 0 {
			// No source code in fragment shader (some drivers will crash in
			// this case).
			s.Error = append(s.Error, []byte(s.Name+" | Fragment shader with no source code.\n")...)

			// Log the error.
			r.logf("%s | Fragment shader with no source code.\n", s.Name)
		} else {
			// Compile fragment shader.
			native.fragment = r.loader.CreateShader(gl.FRAGMENT_SHADER)
			lengths := int32(len(s.GLSLFrag))
			sources := &s.GLSLFrag[0]
			r.loader.ShaderSource(native.fragment, 1, &sources, &lengths)
			r.loader.CompileShader(native.fragment)
			r.loader.Execute()

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
			native.program = r.loader.CreateProgram()
			r.loader.AttachShader(native.program, native.vertex)
			r.loader.AttachShader(native.program, native.fragment)
			r.loader.LinkProgram(native.program)

			// Grab the linker's log.
			var (
				logSize int32
				log     []byte
			)
			r.loader.GetProgramiv(native.program, gl.INFO_LOG_LENGTH, &logSize)
			r.loader.Execute()

			if logSize > 0 {
				log = make([]byte, logSize)
				r.loader.GetProgramInfoLog(native.program, uint32(logSize), nil, &log[0])
				r.loader.Execute()

				// Strip null-termination byte.
				if log[len(log)-1] == 0 {
					log = log[:len(log)-1]
				}
			}

			// Check for linker errors.
			var ok int32
			r.loader.GetProgramiv(native.program, gl.LINK_STATUS, &ok)
			r.loader.Execute()
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
			s.Loaded = true
			s.NativeShader = native
			s.ClearData()

			// Attach a finalizer to the shader that will later free it.
			runtime.SetFinalizer(native, finalizeShader)
		}

		// Flush and execute OpenGL commands.
		r.loader.Flush()
		r.loader.Execute()

		// Unlock, signal completion, and return.
		s.Unlock()
		select {
		case done <- s:
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
