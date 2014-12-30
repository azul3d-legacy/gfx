package webgl

import "io"

type device struct{}

func (d *device) SetDebugOutput(w io.Writer) {
}

func newDevice(ctx interface{}, opts ...Option) (Device, error) {
	return nil, nil
}
