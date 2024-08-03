package simple

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

func TestChatMessageHistory(t *testing.T) {
	t.Parallel()

	sessionID := "test"

	h := NewChatMessageHistory()
	err := h.AddAIMessage(context.Background(), sessionID, "foo")
	require.NoError(t, err)
	err = h.AddUserMessage(context.Background(), sessionID, "bar")
	require.NoError(t, err)

	messages, err := h.Messages(context.Background(), sessionID)
	require.NoError(t, err)

	assert.Equal(t, []llms.ChatMessage{
		llms.AIChatMessage{Content: "foo"},
		llms.HumanChatMessage{Content: "bar"},
	}, messages)

	h = NewChatMessageHistory(
		WithPreviousMessages(sessionID, []llms.ChatMessage{
			llms.AIChatMessage{Content: "foo"},
			llms.SystemChatMessage{Content: "bar"},
		}),
	)
	err = h.AddUserMessage(context.Background(), sessionID, "zoo")
	require.NoError(t, err)

	messages, err = h.Messages(context.Background(), sessionID)
	require.NoError(t, err)

	assert.Equal(t, []llms.ChatMessage{
		llms.AIChatMessage{Content: "foo"},
		llms.SystemChatMessage{Content: "bar"},
		llms.HumanChatMessage{Content: "zoo"},
	}, messages)

	sessions, err := h.Sessions(context.Background())
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, sessionID, sessions[0])

}
