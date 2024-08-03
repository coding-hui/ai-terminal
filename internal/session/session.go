package session

import (
	"context"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

// History is the interface for chat history session store.
type History interface {
	// AddAIMessage is a convenience method for adding an AI message string to
	// the store.
	AddAIMessage(ctx context.Context, sessionID, message string) error

	// AddUserMessage is a convenience method for adding a human message string
	// to the store.
	AddUserMessage(ctx context.Context, sessionID, message string) error

	// AddMessage adds a message to the store.
	AddMessage(ctx context.Context, sessionID string, message llms.ChatMessage) error

	// SetMessages replaces existing messages in the store
	SetMessages(ctx context.Context, sessionID string, messages []llms.ChatMessage) error

	// Messages retrieves all messages from the store
	Messages(ctx context.Context, sessionID string) ([]llms.ChatMessage, error)

	// Sessions retrieves all sessions id from the store
	Sessions(ctx context.Context) ([]string, error)

	// Clear removes all session from the store.
	Clear(ctx context.Context, sessionID string) error
}
