package git

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_GitDir(t *testing.T) {
	g := New()
	dir, err := g.GitDir()
	require.NoError(t, err)
	assert.NotEmpty(t, dir)
	assert.True(t, strings.HasSuffix(strings.TrimSpace(dir), ".git"))
}
