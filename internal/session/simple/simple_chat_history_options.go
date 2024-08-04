package simple

import (
	"path/filepath"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/system"
)

const (
	storeFile      = "chat_history_cache"
	chatEngineMode = "chat"
)

// ChatMessageHistoryOption is a function for creating new chat message history
// with other than the default values.
type ChatMessageHistoryOption func(m *ChatMessageHistory)

func applyChatOptions(options ...ChatMessageHistoryOption) *ChatMessageHistory {
	h := &ChatMessageHistory{
		messages: make(map[string][]llms.ChatMessage),
	}

	for _, option := range options {
		option(h)
	}

	if h.storePath == "" {
		h.storePath = filepath.Join(system.GetHomeDirectory(), ".cache", storeFile)
	}
	if h.chatEngineMode == "" {
		h.chatEngineMode = chatEngineMode
	}

	return h
}

// WithStorePath Path to save the history session file.
func WithStorePath(storePath string) ChatMessageHistoryOption {
	return func(p *ChatMessageHistory) {
		p.storePath = storePath
	}
}

// WithChatEngineMode is an arbitrary key that is used to store the messages of a single chat session,
// like exec,chat etc. Must be set.
func WithChatEngineMode(chatEngineMode string) ChatMessageHistoryOption {
	return func(p *ChatMessageHistory) {
		p.chatEngineMode = chatEngineMode
	}
}
