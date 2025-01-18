package convo

import (
	"context"
	"time"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

// ContentType defines the type of content being loaded into the convo.
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
// It tracks the source, type, convo association, and additional metadata
// about the content being used in conversations.
type LoadContext struct {
	// ID is the auto-increment primary key
	ID uint64 `db:"id" json:"id"`

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

	// ConversationID associates the content with a specific convo
	ConversationID string `db:"conversation_id" json:"conversationId"`

	// UpdatedAt tracks the last modification time of the convo
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

// Conversation represents a chat convo with metadata and history.
// It tracks the convo ID, title, last update time, and optional model info.
type Conversation struct {
	// ID is the unique identifier for the convo
	ID string `db:"id" json:"id"`

	// Title is a human-readable name for the convo
	Title string `db:"title" json:"title"`

	// UpdatedAt tracks the last modification time of the convo
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`

	// Model optionally specifies the AI model used in the convo
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
	// PersistentMessages saves messages to persistent storage
	PersistentMessages(ctx context.Context, convoID string) error
	// InvalidateMessages removes messages from persistent storage
	InvalidateMessages(ctx context.Context, convoID string) error
}

// Store is the interface for chat history convo store.
type Store interface {
	ChatMessageHistory
	LoadContextStore

	// LatestConversation returns the last message in the chat history.
	LatestConversation(ctx context.Context) (*Conversation, error)
	// GetConversation retrieves a convo from the store
	GetConversation(ctx context.Context, convoID string) (*Conversation, error)
	// ListConversations retrieves all convo id from the store
	ListConversations(ctx context.Context) ([]Conversation, error)
	// ListConversationsOlderThan retrieves all convo id from the store that are older than the given time.
	ListConversationsOlderThan(ctx context.Context, t time.Duration) ([]Conversation, error)
	// SaveConversation saves a convo to the store
	SaveConversation(ctx context.Context, id, title, model string) error
	// DeleteConversation removes a convo from the store
	DeleteConversation(ctx context.Context, convoID string) error
	// ClearConversations removes all convo from the store.
	ClearConversations(ctx context.Context) error
	// ConversationExists checks if the given chat convo exists.
	ConversationExists(ctx context.Context, sessionID string) (bool, error)
}

// LoadContextStore manages loaded content contexts
type LoadContextStore interface {
	// SaveContext saves a load context
	SaveContext(ctx context.Context, lc *LoadContext) error
	// GetContext retrieves a load context by ID
	GetContext(ctx context.Context, id uint64) (*LoadContext, error)
	// ListContextsByteConvoID retrieves all load contexts for a convo
	ListContextsByteConvoID(ctx context.Context, conversationID string) ([]LoadContext, error)
	// DeleteContexts removes a load context
	DeleteContexts(ctx context.Context, id uint64) error
	// CleanContexts removes all load contexts for a convo and returns the count deleted
	CleanContexts(ctx context.Context, conversationID string) (int64, error)
}
