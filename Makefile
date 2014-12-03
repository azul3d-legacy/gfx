install:
	go get azul3d.org/clock.v1
	go get azul3d.org/lmath.v1
	go get azul3d.org/keyboard.v1
	go get azul3d.org/mouse.v1
	go test .
	go test ./gl2
	go test ./window
	go test ./internal/util
