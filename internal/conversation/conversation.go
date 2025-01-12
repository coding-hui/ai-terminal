package conversation

import (
	"context"
	"time"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

// ContentType defines the type of content being loaded into the conversation.
// It can be one of: file, URL, or plain text.
type ContentType string

const (
	// ContentTypeFile represents content loaded from a local file
	ContentTypeFile ContentType = "file"

	// ContentTypeURL represents content loaded from a remote URL
	ContentTypeURL ContentType = "url"

	// ContentTypeText represents direct text content
	ContentTypeText ContentType = "text"
)

// LoadContext contains metadata and content information for loaded resources.
// It tracks the source, type, conversation association, and additional metadata
// about the content being used in conversations.
type LoadContext struct {
	// Type indicates the content source type (file, URL, or text)
	Type ContentType `db:"type" json:"type"`

	// URL contains the source URL if loading from a remote resource
	URL string `db:"url" json:"url"`

	// FilePath contains the local file path if loading from a file
	FilePath string `db:"file_path" json:"filePath"`

	// Content contains the actual loaded content as a string
	Content string `db:"content" json:"content"`

	// Name is a human-readable identifier for the content
	Name string `db:"name" json:"name"`

	// ConversationID associates the content with a specific conversation
	ConversationID string `db:"conversation_id" json:"conversationId"`

	// Metadata contains additional key-value pairs about the content
	Metadata map[string]string `db:"metadata" json:"metadata"`
}

// Conversation represents a chat conversation with metadata and history.
// It tracks the conversation ID, title, last update time, and optional model info.
type Conversation struct {
	// ID is the unique identifier for the conversation
	ID string `db:"id" json:"id"`

	// Title is a human-readable name for the conversation
	Title string `db:"title" json:"title"`

	// UpdatedAt tracks the last modification time of the conversation
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`

	// Model optionally specifies the AI model used in the conversation
	Model *string `db:"model" json:"model"`
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
