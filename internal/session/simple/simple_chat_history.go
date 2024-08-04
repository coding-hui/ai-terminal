package simple

import (
	"context"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"golang.org/x/exp/maps"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/session"
)

func init() {
	session.RegisterHistoryStore(&simpleChatMessageHistoryFactory{})
}

type simpleChatMessageHistoryFactory struct{}

func (h *simpleChatMessageHistoryFactory) Type() string {
	return "memory"
}

func (h *simpleChatMessageHistoryFactory) Create(_ options.Config, _ string) (session.History, error) {
	return &ChatMessageHistory{
		messages: make(map[string][]llms.ChatMessage),
	}, nil
}

// ChatMessageHistory is a struct that stores chat messages.
type ChatMessageHistory struct {
	messages map[string][]llms.ChatMessage
}

// Statically assert that History implement the chat message history interface.
var _ session.History = &ChatMessageHistory{}

// NewChatMessageHistory creates a new History using chat message options.
func NewChatMessageHistory(options ...ChatMessageHistoryOption) *ChatMessageHistory {
	return applyChatOptions(options...)
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *ChatMessageHistory) AddAIMessage(ctx context.Context, sessionID, message string) error {
	return h.AddMessage(ctx, sessionID, llms.AIChatMessage{Content: message})
}

// AddUserMessage adds a user to the chat message history.
func (h *ChatMessageHistory) AddUserMessage(ctx context.Context, sessionID, message string) error {
	return h.AddMessage(ctx, sessionID, llms.HumanChatMessage{Content: message})
}

func (h *ChatMessageHistory) AddMessage(_ context.Context, sessionID string, message llms.ChatMessage) error {
	h.messages[sessionID] = append(h.messages[sessionID], message)
	return nil
}

func (h *ChatMessageHistory) SetMessages(_ context.Context, sessionID string, messages []llms.ChatMessage) error {
	h.messages[sessionID] = messages
	return nil
}

func (h *ChatMessageHistory) Messages(_ context.Context, sessionID string) ([]llms.ChatMessage, error) {
	return h.messages[sessionID], nil
}

func (h *ChatMessageHistory) Sessions(_ context.Context) ([]string, error) {
	return maps.Keys(h.messages), nil
}

func (h *ChatMessageHistory) Clear(_ context.Context, sessionID string) error {
	h.messages[sessionID] = make([]llms.ChatMessage, 0)
	return nil
}

func (h *ChatMessageHistory) Exists(_ context.Context, sessionID string) (bool, error) {
	return len(h.messages[sessionID]) > 0, nil
}
