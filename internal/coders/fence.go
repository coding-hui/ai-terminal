package coders

import (
	"fmt"
	"strings"
)

var fences = [][]string{
	{"``" + "`", "``" + "`"},
	{"<source>", "</source>"},
	{"<code>", "</code>"},
	{"<pre>", "</pre>"},
	{"<codeblock>", "</codeblock>"},
	{"<sourcecode>", "</sourcecode>"},
}

func wrapFence(rawContent string) string {
	openFence, closeFence := chooseBestFence(rawContent)
	return fmt.Sprintf("\n%s\n%s\n%s\n", openFence, rawContent, closeFence)
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
	open, close = fences[0][0], fences[0][1]

	return
}
