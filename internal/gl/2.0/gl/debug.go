// Glow automatically generated OpenGL binding: http://github.com/go-gl/glow
// +build 386,!gles2 amd64,!gles2

package gl

import "C"
import "unsafe"

type DebugProc func(
	source uint32,
	gltype uint32,
	id uint32,
	severity uint32,
	length int32,
	message string,
	userParam unsafe.Pointer)

var userDebugCallback DebugProc

//export glowDebugCallback_gl20
func glowDebugCallback_gl20(
	source uint32,
	gltype uint32,
	id uint32,
	severity uint32,
	length int32,
	message *uint8,
	userParam unsafe.Pointer) {
	if userDebugCallback != nil {
		userDebugCallback(source, gltype, id, severity, length, GoStr(message), userParam)
	}
}
