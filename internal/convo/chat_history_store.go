package convo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

const cacheExt = ".gob"

var errInvalidID = errors.New("invalid id")

type SimpleChatHistoryStore struct {
	dir      string
	messages map[string][]llms.ChatMessage
	loaded   map[string]bool // tracks which conversations have been loaded
}

func NewSimpleChatHistoryStore(dir string) *SimpleChatHistoryStore {
	return &SimpleChatHistoryStore{
		dir:      dir,
		messages: make(map[string][]llms.ChatMessage),
		loaded:   make(map[string]bool),
	}
}

// AddAIMessage adds an AIMessage to the chat message convo.
func (h *SimpleChatHistoryStore) AddAIMessage(ctx context.Context, convoID, message string) error {
	return h.AddMessage(ctx, convoID, llms.AIChatMessage{Content: message})
}

// AddUserMessage adds a user to the chat message convo.
func (h *SimpleChatHistoryStore) AddUserMessage(ctx context.Context, convoID, message string) error {
	return h.AddMessage(ctx, convoID, llms.HumanChatMessage{Content: message})
}

func (h *SimpleChatHistoryStore) AddMessage(_ context.Context, convoID string, message llms.ChatMessage) error {
	if !h.loaded[convoID] {
		if err := h.load(convoID); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		h.loaded[convoID] = true
	}
	h.messages[convoID] = append(h.messages[convoID], message)
	return nil
}

func (h *SimpleChatHistoryStore) SetMessages(ctx context.Context, convoID string, messages []llms.ChatMessage) error {
	if err := h.InvalidateMessages(ctx, convoID); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	h.messages[convoID] = messages
	return h.PersistentMessages(ctx, convoID)
}

func (h *SimpleChatHistoryStore) Messages(_ context.Context, convoID string) ([]llms.ChatMessage, error) {
	if !h.loaded[convoID] {
		if err := h.load(convoID); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		h.loaded[convoID] = true
	}
	return h.messages[convoID], nil
}

func (h *SimpleChatHistoryStore) load(convoID string) error {
	if convoID == "" {
		return fmt.Errorf("read: %w", errInvalidID)
	}
	file, err := os.Open(filepath.Join(h.dir, convoID+cacheExt))
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	defer file.Close() //nolint:errcheck

	var rawMessages []llms.ChatMessageModel
	if err := decode(file, &rawMessages); err != nil {
		return fmt.Errorf("read: %w", err)
	}

	h.messages[convoID] = nil
	for _, v := range rawMessages {
		h.messages[convoID] = append(h.messages[convoID], v.ToChatMessage())
	}

	return nil
}

func (h *SimpleChatHistoryStore) PersistentMessages(_ context.Context, convoID string) error {
	if convoID == "" {
		return fmt.Errorf("write: %w", errInvalidID)
	}

	// Ensure directory exists
	if err := os.MkdirAll(h.dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	file, err := os.Create(filepath.Join(h.dir, convoID+cacheExt))
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	defer file.Close() //nolint:errcheck

	var rawMessages []llms.ChatMessageModel
	for _, v := range h.messages[convoID] {
		if v != nil {
			rawMessages = append(rawMessages, llms.ConvertChatMessageToModel(v))
		}
	}
	if err := encode(file, &rawMessages); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func (h *SimpleChatHistoryStore) InvalidateMessages(_ context.Context, convoID string) error {
	if convoID == "" {
		return fmt.Errorf("delete: %w", errInvalidID)
	}
	if err := os.Remove(filepath.Join(h.dir, convoID+cacheExt)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete: %w", err)
	}
	delete(h.messages, convoID)
	delete(h.loaded, convoID)
	return nil
}
