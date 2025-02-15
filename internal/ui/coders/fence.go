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

// chooseExistingFence finds and returns the first existing code fence pair in the rawContent
// where the open fence is followed by "<<<<<<< SEARCH" and the close fence is preceded by ">>>>>>> REPLACE".
// It returns the open and close fence strings if found, otherwise empty strings.
func chooseExistingFence(rawContent string) (open string, close string) {
	// Iterate through all supported fence pairs
	for _, fence := range fences {
		openFence, closeFence := fence[0], fence[1]

		// Find the open fence
		openIndex := strings.Index(rawContent, openFence)
		if openIndex == -1 {
			continue
		}

		// Find the close fence
		closeIndex := strings.Index(rawContent[openIndex+len(openFence):], closeFence)
		if closeIndex == -1 {
			continue
		}
		closeIndex += openIndex + len(openFence)

		// Check if the previous line before the close fence is ">>>>>>> REPLACE"
		beforeClose := rawContent[:closeIndex]
		if !strings.HasSuffix(strings.TrimSpace(beforeClose), UPDATED) {
			continue
		}

		// Check if the next tew line after the open fence is "<<<<<<< SEARCH"
		afterOpen := rawContent[openIndex+len(openFence) : closeIndex]
		if !strings.Contains(strings.TrimSpace(afterOpen), HEAD) {
			continue
		}

		// If both conditions are met, return the fence pair
		return openFence, closeFence
	}

	// If no matching fence pair is found, return default fence pair
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
