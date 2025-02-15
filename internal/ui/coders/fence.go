package coders

import (
	"fmt"
	"path/filepath"
	"strings"
)

var fences = [][]string{
	{"``" + "`", "``" + "`"},
	{"<code>", "</code>"},
	{"<source>", "</source>"},
	{"<pre>", "</pre>"},
	{"<codeblock>", "</codeblock>"},
	{"<sourcecode>", "</sourcecode>"},
}

func wrapFenceWithType(rawContent, filename string) string {
	fileExt := strings.TrimLeft(filepath.Ext(filename), ".")
	openFence, closeFence := chooseBestFence(rawContent)
	return fmt.Sprintf("\n%s%s\n%s\n%s\n", openFence, fileExt, rawContent, closeFence)
}

func defaultBestFence() (open string, close string) {
	return fences[0][0], fences[0][1]
}

func chooseExistingFence(rawContent string) (open string, close string) {
	for _, fence := range fences {
		if strings.Contains(rawContent, fence[0]) && strings.Contains(rawContent, fence[1]) {
			open, close = fence[0], fence[1]
			return
		}
	}
	// Unable to find a fencing strategy!
	return defaultBestFence()
}

func chooseBestFence(rawContent string) (open string, close string) {
	for _, fence := range fences {
		if strings.Contains(rawContent, fence[0]) || strings.Contains(rawContent, fence[1]) {
			continue
		}
		open, close = fence[0], fence[1]
		return
	}

	// Unable to find a fencing strategy!
	return defaultBestFence()
}
