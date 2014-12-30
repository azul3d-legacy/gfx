## Overview

This folder has vendored packages utilized by the various graphics packages (namely the renderers). They are internal packages and should not be used by anyone else.

| Package         | Description                                            |
|-----------------|--------------------------------------------------------|
| gl/2.0/gl       | OpenGL 2.0 wrappers generated using Glow.              |
| gles2/2.0/gles2 | OpenGL ES 2.0 wrappers generated using Glow.           |
| restrict.json   | Glow symbol restriction JSON file.                     |
| procaddr        | Build-tagged version of github.com/go-gl/glow/procaddr |
| resize          | Appengine image resizing package.                      |
| util            | Common gfx.Device utilities.                           |
| glutil          | Standard OpenGL device utilities.                      |
| tag             | Simply exposes a few build tags.                       |

## Glow

Glow (the OpenGL wrapper generator) can be found [on GitHub](http://github.com/go-gl/glow).

## Regenerating

First change directory to the internal folder and then run Glow:

```
cd azul3d.org/gfx.v2-dev/internal
glow download

# OpenGL ES 2 bindings:
glow generate -api=gles2 -version=2.0 -restrict=./restrict.json

# OpenGL 2 bindings:
glow generate -api=gl -version=2.0 -restrict=./restrict.json
```

## Contributing back

If you intend to contribute the regenerated bindings back, you'll need to slightly modify these files:

```
debug.go
conversions.go
package.go
```

`git diff` is your friend in order to see what is missing. But specifically:

Add the build tag line to each file:

```
# OpenGL ES 2
// +build arm gles2

# OpenGL 2
// +build 386 amd64
```

And change the imports (example only):

```
OLD:
	"github.com/go-gl/glow/procaddr"
	"github.com/go-gl/glow/procaddr/auto"
NEW:
	"azul3d.org/gfx.v2-dev/internal/procaddr"
	"azul3d.org/gfx.v2-dev/internal/procaddr/auto"
```
