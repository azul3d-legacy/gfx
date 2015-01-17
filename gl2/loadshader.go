// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"runtime"

	"azul3d.org/gfx.v2-unstable"
	"azul3d.org/gfx.v2-unstable/internal/gl/2.0/gl"
	"azul3d.org/gfx.v2-unstable/internal/glutil"
)

// nativeShader is stored inside the *Shader.Native interface and stores GLSL
// shader IDs.
type nativeShader struct {
	*glutil.LocationCache
	program, vertex, fragment uint32
	r                         *rsrcManager
}

// Implements gfx.Destroyable interface.
func (n *nativeShader) Destroy() {
	finalizeShader(n)
}

// finalizeShader is the finalizer called to free the native shader object. It
// must be free'd in the presence of the OpenGL context, and thus we queue it
// to be free'd at the next available time (next frame).
func finalizeShader(n *nativeShader) {
	n.r.Lock()

	// If the shader program is zero, it has already been free'd.
	if n.program == 0 {
		n.r.Unlock()
		return
	}
	n.r.shaders = append(n.r.shaders, n)
	n.r.Unlock()
}

// free literally frees the native shader object right now. It may only be
// called under the presence of the OpenGL context.
func (n *nativeShader) free() {
	// Delete shader objects (in practice we should be able to do this directly
	// after linking, but it would just leave the driver to reference count
	// them anyway).
	gl.DeleteShader(n.vertex)
	gl.DeleteShader(n.fragment)

	// Delete program.
	gl.DeleteProgram(n.program)

	// Zero-out the nativeShader structure, only keeping the rsrcManager around.
	*n = nativeShader{
		r: n.r,
	}
}

func shaderCompilerLog(s uint32) (log []byte, compiled bool) {
	var ok int32
	gl.GetShaderiv(s, gl.COMPILE_STATUS, &ok)
	if ok == 1 {
		return nil, true
	}

	// Shader compiler error
	var logSize int32
	gl.GetShaderiv(s, gl.INFO_LOG_LENGTH, &logSize)
	if logSize == 0 {
		return nil, ok == 1
	}

	log = make([]uint8, logSize)
	gl.GetShaderInfoLog(s, int32(logSize), nil, &log[0])

	// Strip the null-termination byte.
	log = log[:len(log)-1]
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

	// Perform pre-load checks on the shader.
	doLoad, err := glutil.PreLoadShader(s, done)
	if err != nil {
		r.warner.Warnf("%v\n", err)
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
		gl.ShaderSource(native.vertex, 1, &sources, &lengths)
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
			r.warner.Warnf("%s | Vertex shader errors:\n", s.Name)
			r.warner.Warnf(string(log))
		}

		// Compile fragment shader.
		native.fragment = gl.CreateShader(gl.FRAGMENT_SHADER)
		lengths = int32(len(s.GLSL.Fragment))
		sources = &s.GLSL.Fragment[0]
		gl.ShaderSource(native.fragment, 1, &sources, &lengths)
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
			r.warner.Warnf("%s | Fragment shader errors:\n", s.Name)
			r.warner.Warnf(string(log))
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
				gl.GetProgramInfoLog(native.program, int32(logSize), nil, &log[0])

				// Strip the null-termination byte.
				log = log[:len(log)-1]
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
				r.warner.Warnf("%s | Linker errors:\n", s.Name)
				r.warner.Warnf(string(log))
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
