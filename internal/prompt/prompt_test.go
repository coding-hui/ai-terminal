package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPromptStringByTemplate(t *testing.T) {
	promptTemplate := `{{ .summarize_prefix }}: {{ .summarize_title }}

{{ .summarize_message }}
`
	p, err := GetPromptStringByTemplate(promptTemplate, map[string]any{
		"summarize_prefix":  "feat",
		"summarize_title":   "ut test",
		"summarize_message": "test test test",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, p)
	assert.Contains(t, p, "feat")
}
