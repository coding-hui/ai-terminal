package llm

import (
	"context"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/cli/schema"
)

// ChatMessageHistory is a struct that stores chat messages.
type ChatMessageHistory struct {
	messages []llms.ChatMessage
}

// Statically assert that History implement the chat message history interface.
var _ schema.History = &ChatMessageHistory{}

// NewChatMessageHistory creates a new History using chat message options.
func NewChatMessageHistory(options ...ChatMessageHistoryOption) *ChatMessageHistory {
	return applyChatOptions(options...)
}

// Messages returns all messages stored.
func (h *ChatMessageHistory) Messages(_ context.Context) ([]llms.ChatMessage, error) {
	return h.messages, nil
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *ChatMessageHistory) AddAIMessage(_ context.Context, text string) error {
	h.messages = append(h.messages, llms.AIChatMessage{Content: text})
	return nil
}

// AddUserMessage adds a user to the chat message history.
func (h *ChatMessageHistory) AddUserMessage(_ context.Context, text string) error {
	h.messages = append(h.messages, llms.HumanChatMessage{Content: text})
	return nil
}

func (h *ChatMessageHistory) Clear(_ context.Context) error {
	h.messages = make([]llms.ChatMessage, 0)
	return nil
}

func (h *ChatMessageHistory) AddMessage(_ context.Context, message llms.ChatMessage) error {
	h.messages = append(h.messages, message)
	return nil
}

func (h *ChatMessageHistory) SetMessages(_ context.Context, messages []llms.ChatMessage) error {
	h.messages = messages
	return nil
}
