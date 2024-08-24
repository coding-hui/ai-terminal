package coders

import (
	"fmt"
	"html"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrompts(t *testing.T) {
	t.Run("editBlockCoderPrompt", testEditBlockCoderPrompt)
}

func testEditBlockCoderPrompt(t *testing.T) {
	tpl, err := promptBaseCoder.FormatPrompt(map[string]any{
		userQuestionKey: "add comment",
		addedFilesKey:   "test",
		openFenceKey:    "```",
		closeFenceKey:   "```",
		lazyPromptKey:   lazyPrompt,
	})
	require.NoError(t, err)
	fmt.Println(html.UnescapeString(tpl.String()))
}
