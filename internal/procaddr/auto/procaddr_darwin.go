// +build !egl,!noauto

package auto

import "azul3d.org/gfx.v2-dev/gl2/internal/procaddr/darwin"

func init() {
	GetProcAddress = darwin.GetProcAddress
}
