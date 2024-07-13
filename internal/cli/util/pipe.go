package util

import (
	"bufio"
	"io"
	"os"
	"strings"

	"k8s.io/klog/v2"
)

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
