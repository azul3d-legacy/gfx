// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gl2 provides an OpenGL 2 based graphics device.
//
// The behavior of the device is fully defined in the gfx package, and as such
// the following documentation only makes note of strictly OpenGL related
// caveats (like initialization, etc).
//
// Window Toolkit
//
// The device operates independently of the window toolkit in use. It can work
// well with any window toolkit that provides an OpenGL 2 feature-level context
// (GLFW, SDL, QT, etc).
//
// If you just want to open a window and draw graphics to it and don't care
// about the specifics, the simpler gfx/window package should be used instead:
//
// http://azul3d.org/gfx.v2/window
//
// Feedback Loops
//
// When performing render-to-texture (RTT), feedback loops are explicitly
// prohibited.
//
// This means that the device will panic if you attempt to draw an object to a
// RTT canvas when the object uses the literal RTT texture in itself.
//
// That is, drawing an object with a texture that of the final render
// destination will panic. Such recursive drawing is prohibited by OpenGL, and
// as such is now allowed.
//
// Mipmapping
//
// The gfx package allows turning on and off mipmapping of a loaded texture
// dynamically by setting it's minification filter to a mipmapped or
// non-mipmapped one. This OpenGL device reflects this behavior, but with a
// small caveat:
//
// Mipmapping can only be toggled post-load in the event that the texture was
// first loaded with a mipmapped filter, or else the request is ignored and the
// texture filter is not changed.
//
// This is done for performance reasons, turning on mipmapping post-load time
// would require a full texture reload (and having it on by default would use
// more memory due to mipmaps always being generated).
//
// Uniforms
//
// A gfx.Shader will have all of it's inputs (from the Shader.Inputs map)
// mapped by name to uniforms within the fragment and vertex shader programs.
//
// If you do not intend to utilize a uniform value, then omit it in your GLSL
// code and it will not be fed into the GLSL program.
//
// The default uniforms are:
//
//  uniform mat4 Model;       -> Model matrix from gfx.Object.Transform
//  uniform mat4 View;        -> View matrix from gfx.Camera.Transform
//  uniform mat4 Projection;  -> Projection matrix from gfx.Camera.Projection
//  uniform mat4 MVP;         -> Premultiplied Model/View/Projection matrix.
//  uniform bool BinaryAlpha; -> See below.
//
// BinaryAlpha is a boolean uniform value that informs the shader of the chosen
// alpha transparency mode of an object. It is set to true if the gfx.Object
// being drawn has a gfx.State.AlphaMode of gfx.BinaryAlpha or if the alpha
// mode is gfx.AlphaToCoverage but the GPU does not support it.
//
// Vertex Attributes
//
// A mesh will have all of it's attributes (from the Mesh.Attribs map) mapped
// by name to attributes within the GLSL fragment and vertex shader programs.
//
// Like uniforms, if you do not intend to utilize a attribute value, then omit
// it in your GLSL code and it will simply not be fed into the GLSL program.
//
// The default vertex attributes are:
//
//  attribute vec3 Vertex;      -> from gfx.Mesh.Vertices and gfx.Mesh.Indices
//  attribute vec4 Color;       -> from gfx.Mesh.Colors
//  attribute vec3 Bary;        -> from gfx.Mesh.Bary
//  attribute vec2 TexCoord[N]; -> [N] is the nth index of gfx.Mesh.TexCoords
//
// Uniform And Attribute Types
//
// In both the case of uniforms as well as attributes, data types from the gfx
// package are mapped directly to their GLSL equivilent:
//
//  gfx.Vec4     -> vec4
//  gfx.Vec3     -> vec3
//  gfx.Color    -> vec4 (GLSL does not have a dedicated color type)
//  gfx.TexCoord -> vec2 (GLSL does not have a dedicated texture coordinate type)
//
// Slices are mapped directly to GLSL arrays, which can be fixed or dynamically
// sized, standard GLSL restrictions apply (such as a lack of dynamic indexing
// on dynamically sized arrays, etc).
//
// Basic Usage
//
// Functions recieved over the device's execution channel should be executed
// under the presence of the OpenGL context, when a device function returns
// true gfx.Device.Canvas.Render has been called meaning the frame is finished
// and the window's buffers should be swapped.
//
//  func main() {
//      // Always make proper use of LockOSThread! The OpenGL context is tied
//      // to the OS thread.
//      runtime.LockOSThread()
//
//      // Initialize OpenGL, through your window toolkit of choice.
//      window.GLInit()
//
//      // Initialize the device, under the OpenGL context.
//      d, err := gl2.New()
//      handle(err)
//
//      // Main loop, execute device functions.
//      exec := device.Exec()
//      for {
//          if window.Closed() {
//              // Cleanup the device.
//              device.Destroy()
//              break
//          }
//
//          // Wait for a task from the device to execute.
//          f := <-exec
//          if f() {
//              // A frame was rendered, swap the buffers.
//              window.GLSwapBuffers()
//          }
//      }
//  }
//
// Debugging
//
// User applications (i.e. not this device itself) can be debugged using the
// SetDebugOutput method of the Device interface exposed by this package. This
// has little to no overhead.
//
// More extensive debugging can be enabled via the "gfxdebug" build tag, which
// enables debugging of the OpenGL device itself (not user code). Do not use
// this build tag except when debugging, as it's performance is purely driver
// implementation dependant. This form of debugging should be used with a debug
// OpenGL context.
//
// Examples
//
// The examples repository contains several examples which utilize the gfx core
// packages. Please see:
//
// https://azul3d.org/examples.v1
//
package gl2 // import "azul3d.org/gfx.v2-dev/gl2"
