package term

import (
	"bufio"
	"io"
	"os"
	"strings"

	"k8s.io/klog/v2"
)

// ReadPipeInput reads input from a pipe and returns it as a string.
// It checks if the input is coming from a pipe and reads the content until EOF.
// Returns:
//   - A string containing the pipe input if successful.
//   - An empty string if there is an error or no input from the pipe.
func ReadPipeInput() string {
	// Check the status of standard input to determine if it's a pipe.
	stat, err := os.Stdin.Stat()
	if err != nil {
		klog.Warningf("Failed to start pipe input: %v", err)
		return ""
	}
	pipe := ""
	// Check if the input is coming from a pipe and not empty.
	if !(stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0) {
		reader := bufio.NewReader(os.Stdin)
		var builder strings.Builder

		// Read runes from the pipe until EOF is reached.
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

		// Trim any leading or trailing whitespace from the input.
		pipe = strings.TrimSpace(builder.String())
	}

	return pipe
}
