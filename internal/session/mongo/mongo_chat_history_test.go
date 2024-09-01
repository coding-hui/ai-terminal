package mongo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

func runTestContainer() (string, error) {
	ctx := context.Background()

	mongoContainer, err := mongodb.Run(
		ctx,
		"mongo:7.0.8",
		mongodb.WithUsername("test"),
		mongodb.WithPassword("test"),
	)
	if err != nil {
		return "", err
	}

	url, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		return "", err
	}
	return url, nil
}

func TestMongoDBChatMessageHistory(t *testing.T) {
	t.Parallel()

	sessionID := "test"
	chatEngineMode := "chat"

	url, err := runTestContainer()
	require.NoError(t, err)

	ctx := context.Background()
	_, err = NewMongoDBChatMessageSession(ctx)
	assert.Equal(t, errMongoInvalidURL, err)

	_, err = NewMongoDBChatMessageSession(ctx, WithConnectionURL(url))
	assert.Equal(t, errMongoInvalidChatEngineMode, err)

	history, err := NewMongoDBChatMessageSession(ctx, WithConnectionURL(url), WithChatEngineMode(chatEngineMode))
	require.NoError(t, err)

	err = history.AddAIMessage(ctx, sessionID, "Hi")
	require.NoError(t, err)

	err = history.AddUserMessage(ctx, sessionID, "Hello")
	require.NoError(t, err)

	messages, err := history.Messages(ctx, sessionID)
	require.NoError(t, err)

	assert.Len(t, messages, 2)
	assert.Equal(t, llms.ChatMessageTypeAI, messages[0].GetType())
	assert.Equal(t, "Hi", messages[0].GetContent())
	assert.Equal(t, llms.ChatMessageTypeHuman, messages[1].GetType())
	assert.Equal(t, "Hello", messages[1].GetContent())

	sessions, err := history.Sessions(ctx)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, sessionID, sessions[0])

	t.Cleanup(func() {
		require.NoError(t, history.Clear(ctx, sessionID))
	})
}
