// Copyright 2020 ActiveState Software. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file

package stacktrace

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// Stacktrace represents a stacktrace
type Stacktrace struct {
	Frames []Frame
}

// Frame is a single frame in a stacktrace
type Frame struct {
	// Func contains a function name.
	Func string
	// Line contains a line number.
	Line int
	// Path contains a file path.
	Path string
	// Package is the package name for this frame
	Package string
}

// FrameCap is a default cap for frames array.
// It can be changed to number of expected frames
// for purpose of performance optimisation.
var FrameCap = 20

// String returns a string representation of a stacktrace
func (t *Stacktrace) String() string {
	result := []string{}
	for _, frame := range t.Frames {
		result = append(result, fmt.Sprintf(`%s:%s:%d`, frame.Path, frame.Func, frame.Line))
	}
	return strings.Join(result, "\n")
}

// Get returns a stacktrace
func Get() *Stacktrace {
	stacktrace := &Stacktrace{}
	pc := make([]uintptr, FrameCap)
	n := runtime.Callers(1, pc)
	if n == 0 {
		return stacktrace
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	var skipFile, skipPkg string
	for {
		frame, more := frames.Next()
		pkg := strings.Split(filepath.Base(frame.Func.Name()), ".")[0]

		// Skip our own path
		if skipFile == "" {
			skipFile = filepath.Dir(frame.File)
			skipPkg = pkg
		}
		if strings.Contains(frame.File, skipFile) && pkg == skipPkg {
			continue
		}

		stacktrace.Frames = append(stacktrace.Frames, Frame{
			Func:    frame.Func.Name(),
			Line:    frame.Line,
			Path:    frame.File,
			Package: pkg,
		})

		if !more {
			break
		}
	}

	return stacktrace
}
