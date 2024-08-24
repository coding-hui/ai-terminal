package util

import (
	"bufio"
	"io"
	"os"
	"strings"

	"k8s.io/klog/v2"
)

// ReadPipeInput reads input from a pipe and returns it as a string.
// It checks if the input is coming from a pipe and reads it rune by rune.
// If the input is not from a pipe or there's an error, it logs a warning and returns an empty string.
//
// Returns:
//   - A string containing the input from the pipe, trimmed of leading and trailing whitespace.
//   - An empty string if there's no input from the pipe or an error occurs.
func ReadPipeInput() string {
	stat, err := os.Stdin.Stat()
	if err != nil {
		klog.Warningf("Failed to start pipe input: %v", err)
		return ""
	}
	pipe := ""
	if !(stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0) {
		reader := bufio.NewReader(os.Stdin)
		var builder strings.Builder

		for {
			r, _, err := reader.ReadRune()
			if err != nil && err == io.EOF {
				break
			}
			_, err = builder.WriteRune(r)
			if err != nil {
				klog.Warningf("Failed to getting pipe input: %v", err)
				return ""
			}
		}

		pipe = strings.TrimSpace(builder.String())
	}

	return pipe
}
