install:
	go get azul3d.org/gfx.v2-dev/...
	go test .
	go test ./gl2
	go test ./window
	go test ./internal/util
