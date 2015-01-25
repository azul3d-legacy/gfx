package camera

import "azul3d.org/gfx.v2-unstable"

// A check for whether or not *camera.Camera implements gfx.Camera properly.
var _ gfx.Camera = New()
