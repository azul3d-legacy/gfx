# Azul3D - gfx [![Build Status](https://travis-ci.org/azul3d/gfx.svg?branch=master)](https://travis-ci.org/azul3d/gfx)

This repository hosts packages for Azul3D's graphics core.

| Package | Description |
|---------|-------------|
| [azul3d.org/gfx.v2](https://azul3d.org/gfx.v2) | *A Go interface to modern GPU rendering API's.* |
| [azul3d.org/gfx.v2/window](https://azul3d.org/gfx.v2/window) | *Cross-platform window and graphics device management.* |
| [azul3d.org/gfx.v2/gl2](https://azul3d.org/gfx.v2/gl2) | *A OpenGL 2 based graphics device.* |

## Version 2

##### [gfx.v2](https://azul3d.org/gfx.v2) package:

* Added `Mesh.Append` method to append two meshes together (see [#21](https://github.com/azul3d/gfx/issues/21)).
* Added `MeshState` type to check if two meshes can append together perfectly (see [#21](https://github.com/azul3d/gfx/issues/21)).
* `TexCoord` and `Color` are now valid types for use in the `Shader.Input` map and as data to `VertexAttrib` (see [#23](https://github.com/azul3d/gfx/issues/23)).
* Added a convenience `Mesh.Normals` slice for storing the normals of a mesh (see [#11](https://github.com/azul3d/gfx/issues/11)).
* The TexWrap mode `BorderColor` is not always present, e.g. in OpenGL ES 2 (see [#56](https://github.com/azul3d/gfx/issues/56)).
* Clarify: Some renderers, e.g. OpenGL ES, only support boolean occlusion queries (see [#57](https://github.com/azul3d/gfx/issues/57)).
* Empty rectangles are no longer considered "the entire area" (see [#15](https://github.com/azul3d/gfx/issues/15)).
* `Renderer` is now `Device` (See [#60](https://github.com/azul3d/gfx/issues/60)).
* `GPUInfo` is now `DeviceInfo` (See [#60](https://github.com/azul3d/gfx/issues/60)).
* `Renderer.GPUInfo()` is now `Device.Info()` (See [#60](https://github.com/azul3d/gfx/issues/60)).
* Split OpenGL and GLSL information out of DeviceInfo struct (See [#62](https://github.com/azul3d/gfx/issues/62)).
* A seperate structure is used for GLSL shader sources now (See [#63](https://github.com/azul3d/gfx/issues/63)).
* Added support for drawing meshes as Lines and Points (See [#48](https://github.com/azul3d/gfx/issues/48)).

##### [gfx.v2/window](https://azul3d.org/gfx.v2/window) package:

* Moved to this repository as a sub-package (see [old repository](https://github.com/azul3d/gfx-window) and [issue 33](https://github.com/azul3d/issues/issues/33)).
* Better documentation (see [#49](https://github.com/azul3d/gfx/pull/49)).
* Added support for multiple windows (see [#38](https://github.com/azul3d/gfx/issues/38)).
* Exposed the main thread for clients that need it (see [#39](https://github.com/azul3d/gfx/issues/39)).
* Uses a 24/bpp framebuffer by default (see [#24](https://github.com/azul3d/gfx/issues/41)).
* The `gles2` build tag enables the use of the OpenGL ES 2 renderer on desktops (see [#43](https://github.com/azul3d/gfx/issues/43)).

##### [gfx.v2/gl2](https://azul3d.org/gfx.v2/gl2) package:

* Moved to this repository as a sub-package (see [old repository](https://github.com/azul3d/gfx-gl2) and [issue 33](https://github.com/azul3d/issues/issues/33)).
* Renderer now uses just one OpenGL context (see [#24](https://github.com/azul3d/gfx/issues/24)).
* Improved package documentation ([view commit](https://github.com/azul3d/gfx-gl2/commit/493f72dbb36547e394f2d4995ee7d74dbf7b86d4)).
* `gl2.Renderer` is now `gl2.Device` (See [#60](https://github.com/azul3d/gfx/issues/60)).
* `gl2.Renderer` is now an interface (See [#52](https://github.com/azul3d/gfx/issues/52)).
* `gl2.New` now takes option function parameters (See [#53](https://github.com/azul3d/gfx/issues/53)).
* Documented basic usage and window toolkit independence (See [#54](https://github.com/azul3d/gfx/issues/54)).
* Fix caching failure of shader uniform locations (See [#58](https://github.com/azul3d/gfx/issues/58)).
* Assets are now (optionally) shared across multiple renderers (See [#28](https://github.com/azul3d/gfx/issues/28)).

##### [gfx.v2/gfxsort](https://azul3d.org/gfx.v2/gfxsort) package:

* Sorting utilities from the `gfx` package moved here (See [#59](https://github.com/azul3d/gfx/issues/59)).
  * `ByDist` `ByState` and `InsertionSort`.

[Full v2 Changelog](https://github.com/azul3d/gfx/compare/v1.0.1...v2).

## Version 1.0.1

##### [gfx.v1](https://azul3d.org/gfx.v1) changes:

* Fixed a bug causing Transforms to be constantly recalculated (see [#16](https://github.com/azul3d/gfx/issues/16)).

[Full v1.0.1 Changelog](https://github.com/azul3d/gfx/compare/v1...v1.0.1).

## Version 1

##### [gfx.v1](https://azul3d.org/gfx.v1) changes:

* Initial implementation.

[Full v1 Changelog](https://github.com/azul3d/gfx/compare/24fcb440482034e45fba7fcbdd21fa9a7abbe6e6...v1).
