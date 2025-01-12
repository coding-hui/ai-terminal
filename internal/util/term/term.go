// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package term provides structures and helper functions to work with
// terminal (state, sizes).
package term

import (
	"io"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/mattn/go-isatty"
)

// TTY helps invoke a function and preserve the state of the terminal, even if the process is
// terminated during execution. It also provides support for terminal resizing for remote command
// execution/attachment.
type TTY struct {
	// In is a reader representing stdin. It is a required field.
	In io.Reader
	// Out is a writer representing stdout. It must be set to support terminal resizing. It is an
	// optional field.
	Out io.Writer
	// Raw is true if the terminal should be set raw.
	Raw bool
	// TryDev indicates the TTY should try to open /dev/tty if the provided input
	// is not a file descriptor.
	TryDev bool
}

var IsInputTTY = sync.OnceValue(func() bool {
	return isatty.IsTerminal(os.Stdin.Fd())
})

var IsOutputTTY = sync.OnceValue(func() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
})

// SanitizeFilename cleans up a filename to make it safe for use in file systems
// by removing or replacing invalid characters.
func SanitizeFilename(filename string) string {
	// Remove any path components
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")

	// Remove other potentially problematic characters
	reg := regexp.MustCompile(`[<>:"|?*]`)
	filename = reg.ReplaceAllString(filename, "_")

	// Trim spaces and dots from start/end
	filename = strings.TrimSpace(filename)
	filename = strings.Trim(filename, ".")

	// Ensure filename is not empty
	if filename == "" {
		filename = "unnamed"
	}

	// Limit length to 255 characters (common filesystem limit)
	if len(filename) > 64 {
		filename = filename[:64]
	}

	return filename
}
