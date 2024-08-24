package coders

import (
	"context"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/prompts"
)

// PartialCodeBlock represents a partial code block with its file path and original and updated text.
type PartialCodeBlock struct {
	Path, OriginalText, UpdatedText string
}

// Coder defines the interface for a code editor.
type Coder interface {
	// Name returns the name of the code editor.
	Name() string

	// Prompt returns the prompt template used by the code editor.
	Prompt() prompts.ChatPromptTemplate

	// FormatMessages formats the messages with the provided values and returns the formatted messages.
	FormatMessages(values map[string]any) ([]llms.MessageContent, error)

	// GetEdits retrieves the list of edits made to the code.
	GetEdits(ctx context.Context) ([]PartialCodeBlock, error)

	// ApplyEdits applies the given list of edits to the code.
	ApplyEdits(ctx context.Context, edits []PartialCodeBlock) error

	// Execute runs the code editor with the specified input messages.
	Execute(ctx context.Context, messages []llms.MessageContent) error
}
