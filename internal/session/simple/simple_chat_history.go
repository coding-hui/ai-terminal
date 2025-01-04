package simple

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/coding-hui/common/util/fileutil"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/session"
)

func init() {
	session.RegisterHistoryStore(&simpleChatMessageHistoryFactory{})
}

type simpleChatMessageHistoryFactory struct{}

func (h *simpleChatMessageHistoryFactory) Type() string {
	return "file"
}

func (h *simpleChatMessageHistoryFactory) Create(cfg options.Config, chatEngineMode string) (session.History, error) {
	return NewChatMessageHistory(
		WithStorePath(cfg.DataStore.Path),
		WithChatEngineMode(chatEngineMode),
	), nil
}

// ChatMessageHistory is a struct that stores chat messages.
type ChatMessageHistory struct {
	storePath      string
	chatEngineMode string
	messages       map[string][]llms.ChatMessage
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

func (h *ChatMessageHistory) AddMessage(ctx context.Context, sessionID string, message llms.ChatMessage) error {
	if err := h.load(ctx, sessionID); err != nil {
		return err
	}
	h.messages[sessionID] = append(h.messages[sessionID], message)
	return h.persistent(ctx, sessionID)
}

func (h *ChatMessageHistory) SetMessages(ctx context.Context, sessionID string, messages []llms.ChatMessage) error {
	h.messages[sessionID] = messages
	if err := h.invalidate(ctx, sessionID); err != nil {
		return err
	}
	return h.persistent(ctx, sessionID)
}

func (h *ChatMessageHistory) Messages(ctx context.Context, sessionID string) ([]llms.ChatMessage, error) {
	if err := h.load(ctx, sessionID); err != nil {
		return nil, err
	}
	return h.messages[sessionID], nil
}

func (h *ChatMessageHistory) Sessions(_ context.Context) ([]string, error) {
	var sessions []string
	err := filepath.Walk(h.storePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			name := strings.Split(info.Name(), "_")[0]
			if len(name) > 0 && !slices.Contains(sessions, name) {
				sessions = append(sessions, name)
			}
		}
		return nil
	})
	return sessions, err
}

func (h *ChatMessageHistory) Clear(ctx context.Context, sessionID string) error {
	h.messages[sessionID] = make([]llms.ChatMessage, 0)
	return h.invalidate(ctx, sessionID)
}

func (h *ChatMessageHistory) Exists(ctx context.Context, sessionID string) (bool, error) {
	if err := h.load(ctx, sessionID); err != nil {
		return false, err
	}
	return len(h.messages[sessionID]) > 0, nil
}

func (h *ChatMessageHistory) invalidate(_ context.Context, sessionID string) error {
	filePath := filepath.Join(h.storePath, sessionID+"_"+h.chatEngineMode)
	if err := os.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return nil
}

func (h *ChatMessageHistory) load(_ context.Context, sessionID string) error {
	if h.isTempSession(sessionID) {
		return nil
	}
	bytes, err := os.ReadFile(filepath.Join(h.storePath, sessionID+"_"+h.chatEngineMode))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if len(bytes) == 0 {
		return nil
	}
	var rawMessages []llms.ChatMessageModel
	err = json.Unmarshal(bytes, &rawMessages)
	if err != nil {
		return err
	}
	messages := []llms.ChatMessage{}
	for _, msg := range rawMessages {
		messages = append(messages, msg.ToChatMessage())
	}
	h.messages[sessionID] = messages
	return nil
}

func (h *ChatMessageHistory) persistent(_ context.Context, sessionID string) error {
	if len(h.messages[sessionID]) == 0 || h.isTempSession(sessionID) {
		return nil
	}
	rawMessages := []llms.ChatMessageModel{}
	for _, msg := range h.messages[sessionID] {
		rawMessages = append(rawMessages, llms.ConvertChatMessageToModel(msg))
	}
	bytes, err := json.Marshal(rawMessages)
	if err != nil {
		return err
	}
	return fileutil.WriteFile(filepath.Join(h.storePath, sessionID+"_"+h.chatEngineMode), bytes)
}

func (h *ChatMessageHistory) isTempSession(sessionID string) bool {
	return sessionID == "temp_session"
}
