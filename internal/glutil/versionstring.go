// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glutil

import (
	"strconv"
	"strings"
)

// ParseVersionString parses a OpenGL version string and returns it's
// components.
//
// The string returned may be 'major.minor' or 'major.minor.release' and may
// be followed by a space and any vendor specific information. For more
// information see:
//
// http://www.opengl.org/sdk/docs/man/xhtml/glGetString.xml
//
func ParseVersionString(ver string) (major, minor, release int, vendor string) {
	if len(ver) == 0 {
		// Version string must not be empty
		return
	}

	// First locate a proper version string without vendor specific
	// information.
	var (
		versionString string
		err           error
	)
	if strings.Contains(ver, " ") {
		// It must have vendor information
		split := strings.Split(ver, " ")
		if len(split) > 0 || len(split[0]) > 0 {
			// Everything looks good.
			versionString = split[0]
		} else {
			// Something must be wrong with their vendor string.
			return
		}

		// Store the vendor version information.
		vendor = ver[len(versionString):]
	} else {
		// No vendor information.
		versionString = ver
	}

	// We have a proper version string now without vendor information.
	dots := strings.Count(versionString, ".")
	if dots == 1 {
		// It's a 'major.minor' style string
		versions := strings.Split(versionString, ".")
		if len(versions) == 2 {
			major, err = strconv.Atoi(versions[0])
			if err != nil {
				return
			}

			minor, err = strconv.Atoi(versions[1])
			if err != nil {
				return
			}

		} else {
			return
		}

	} else if dots == 2 {
		// It's a 'major.minor.release' style string
		versions := strings.Split(versionString, ".")
		if len(versions) == 3 {
			major, err = strconv.Atoi(versions[0])
			if err != nil {
				return
			}

			minor, err = strconv.Atoi(versions[1])
			if err != nil {
				return
			}

			release, err = strconv.Atoi(versions[2])
			if err != nil {
				return
			}
		} else {
			return
		}
	}
	return
}
