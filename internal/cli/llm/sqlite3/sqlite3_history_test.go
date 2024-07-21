package sqlite3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/cli/llm/sqlite3"
)

func TestSqliteChatMessageHistory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	h := sqlite3.NewSqliteChatMessageHistory(sqlite3.WithContext(ctx))

	err := h.AddAIMessage(ctx, "foo")
	require.NoError(t, err)

	err = h.AddUserMessage(ctx, "bar")
	require.NoError(t, err)

	messages, err := h.Messages(ctx)
	require.NoError(t, err)

	assert.Equal(t, []llms.ChatMessage{
		llms.AIChatMessage{Content: "foo"},
		llms.HumanChatMessage{Content: "bar"},
	}, messages)

	h = sqlite3.NewSqliteChatMessageHistory(
		sqlite3.WithContext(ctx),
		sqlite3.WithOverwrite(),
	)

	err = h.SetMessages(ctx,
		[]llms.ChatMessage{
			llms.AIChatMessage{Content: "foo"},
			llms.SystemChatMessage{Content: "bar"},
		})
	require.NoError(t, err)

	err = h.AddUserMessage(ctx, "zoo")
	require.NoError(t, err)

	messages, err = h.Messages(ctx)
	require.NoError(t, err)

	assert.Equal(t, []llms.ChatMessage{
		llms.AIChatMessage{Content: "foo"},
		llms.SystemChatMessage{Content: "bar"},
		llms.HumanChatMessage{Content: "zoo"},
	}, messages)
}