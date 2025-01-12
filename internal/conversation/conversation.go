package conversation

import (
	"context"
	"time"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

type Conversation struct {
	ID        string    `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Model     *string   `db:"model" json:"model"`
}

type ChatMessageHistory interface {
	// AddAIMessage is a convenience method for adding an AI message string to
	// the store.
	AddAIMessage(ctx context.Context, convoID, message string) error

	// AddUserMessage is a convenience method for adding a human message string
	// to the store.
	AddUserMessage(ctx context.Context, convoID, message string) error

	// AddMessage adds a message to the store.
	AddMessage(ctx context.Context, convoID string, message llms.ChatMessage) error

	// SetMessages replaces existing messages in the store
	SetMessages(ctx context.Context, convoID string, messages []llms.ChatMessage) error

	// Messages retrieves all messages from the store
	Messages(ctx context.Context, convoID string) ([]llms.ChatMessage, error)
}

// Store is the interface for chat history conversation store.
type Store interface {
	ChatMessageHistory

	// Latest returns the last message in the chat history.
	Latest(ctx context.Context) (*Conversation, error)

	// Get retrieves a conversation from the store
	Get(ctx context.Context, convoID string) (*Conversation, error)

	// List retrieves all conversation id from the store
	List(ctx context.Context) ([]Conversation, error)

	// ListOlderThan retrieves all conversation id from the store that are older than the given time.
	ListOlderThan(ctx context.Context, t time.Duration) ([]Conversation, error)

	// Save saves a conversation to the store
	Save(ctx context.Context, id, title, model string) error

	// Delete removes a conversation from the store
	Delete(ctx context.Context, convoID string) error

	// Clear removes all conversation from the store.
	Clear(ctx context.Context) error

	//Exists checks if the given chat conversation exists.
	Exists(ctx context.Context, sessionID string) (bool, error)
}
