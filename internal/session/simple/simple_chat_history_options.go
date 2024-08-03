package simple

import "github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

// ChatMessageHistoryOption is a function for creating new chat message history
// with other than the default values.
type ChatMessageHistoryOption func(m *ChatMessageHistory)

// WithPreviousMessages is an option for NewChatMessageHistory for adding
// previous messages to the history.
func WithPreviousMessages(sessionID string, previousMessages []llms.ChatMessage) ChatMessageHistoryOption {
	return func(m *ChatMessageHistory) {
		m.messages[sessionID] = append(m.messages[sessionID], previousMessages...)
	}
}

func applyChatOptions(options ...ChatMessageHistoryOption) *ChatMessageHistory {
	h := &ChatMessageHistory{
		messages: make(map[string][]llms.ChatMessage),
	}

	for _, option := range options {
		option(h)
	}

	return h
}
