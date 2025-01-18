package sqlite3

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/convo"
)

func TestSqliteStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	convoID := convo.NewConversationID()
	h := NewSqliteStore(
		WithConversation(convoID),
		WithContext(ctx),
		WithDataPath(t.TempDir()),
	)

	t.Run("Save and Get conversation", func(t *testing.T) {
		err := h.SaveConversation(ctx, convoID, "foo", "test")
		require.NoError(t, err)

		convo, err := h.GetConversation(ctx, convoID)
		require.NoError(t, err)
		assert.Equal(t, "foo", convo.Title)
		assert.False(t, convo.UpdatedAt.IsZero())
	})

	t.Run("Get non-existent conversation", func(t *testing.T) {
		_, err := h.GetConversation(ctx, "nonexistent")
		assert.True(t, errors.Is(err, errNoMatches))
	})

	t.Run("Delete conversation", func(t *testing.T) {
		require.NoError(t, h.DeleteConversation(ctx, convoID))

		ok, err := h.ConversationExists(ctx, convoID)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("List conversations", func(t *testing.T) {
		// Create multiple conversations
		ids := []string{
			convo.NewConversationID(),
			convo.NewConversationID(),
		}
		require.NoError(t, h.SaveConversation(ctx, ids[0], "first", "test"))
		require.NoError(t, h.SaveConversation(ctx, ids[1], "second", "test"))

		convos, err := h.ListConversations(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(convos), 2)
	})

	t.Run("ListOlderThan", func(t *testing.T) {
		oldID := convo.NewConversationID()
		require.NoError(t, h.SaveConversation(ctx, oldID, "old", "test"))

		// Update timestamp to be old
		_, err := h.DB.ExecContext(ctx, `
			UPDATE conversations 
			SET updated_at = datetime('now', '-1 hour')
			WHERE id = ?
		`, oldID)
		require.NoError(t, err)

		convos, err := h.ListConversationsOlderThan(ctx, 30*time.Minute)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(convos), 1)
	})

	t.Run("Clear all conversations", func(t *testing.T) {
		require.NoError(t, h.ClearConversations(ctx))

		convos, err := h.ListConversations(ctx)
		require.NoError(t, err)
		assert.Empty(t, convos)
	})
}

func TestSqliteChatMessageHistory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	convoID := convo.NewConversationID()
	h := NewSqliteStore(
		WithContext(ctx),
		WithDataPath(t.TempDir()),
	)

	t.Run("Add and get messages", func(t *testing.T) {
		err := h.AddAIMessage(ctx, convoID, "foo")
		require.NoError(t, err)

		err = h.AddUserMessage(ctx, convoID, "bar")
		require.NoError(t, err)

		messages, err := h.Messages(ctx, convoID)
		require.NoError(t, err)

		assert.Equal(t, []llms.ChatMessage{
			llms.AIChatMessage{Content: "foo"},
			llms.HumanChatMessage{Content: "bar"},
		}, messages)
	})

	t.Run("Set and add messages", func(t *testing.T) {
		h = NewSqliteStore(
			WithContext(ctx),
			WithDataPath(t.TempDir()),
		)

		err := h.SetMessages(ctx,
			convoID,
			[]llms.ChatMessage{
				llms.AIChatMessage{Content: "foo"},
				llms.SystemChatMessage{Content: "bar"},
			})
		require.NoError(t, err)

		err = h.AddUserMessage(ctx, convoID, "zoo")
		require.NoError(t, err)

		messages, err := h.Messages(ctx, convoID)
		require.NoError(t, err)

		assert.Equal(t, []llms.ChatMessage{
			llms.AIChatMessage{Content: "foo"},
			llms.SystemChatMessage{Content: "bar"},
			llms.HumanChatMessage{Content: "zoo"},
		}, messages)
	})

	t.Run("Get messages from non-existent conversation", func(t *testing.T) {
		messages, err := h.Messages(ctx, "nonexistent")
		assert.NoError(t, err)
		assert.Len(t, messages, 0)
	})
}
