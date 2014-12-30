// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glc

import "errors"

var (
	InvalidEnum                 = errors.New("GL_INVALID_ENUM")
	InvalidValue                = errors.New("GL_INVALID_VALUE")
	InvalidOperation            = errors.New("GL_INVALID_OPERATION")
	InvalidFramebufferOperation = errors.New("GL_INVALID_FRAMEBUFFER_OPERATION")
	OutOfMemory                 = errors.New("GL_OUT_OF_MEMORY")
	StackUnderflow              = errors.New("GL_STACK_UNDERFLOW")
	StackOverflow               = errors.New("GL_STACK_OVERFLOW")
)
