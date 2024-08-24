package coders

import (
	"context"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/prompts"
)

type EditContent struct {
	Path, OriginalText, UpdatedText string
}

type CodeEditor interface {
	// Name returns the name of the code editor.
	Name() string

	// Prompt returns the prompt template used by the code editor.
	Prompt() prompts.ChatPromptTemplate

	// FormatMessages formats the messages with the provided values and returns the formatted messages.
	FormatMessages(values map[string]any) ([]llms.MessageContent, error)

	// GetEdits retrieves the list of edits made to the code.
	GetEdits(ctx context.Context) ([]EditContent, error)

	// ApplyEdits applies the given list of edits to the code.
	ApplyEdits(ctx context.Context, edits []EditContent) error

	// Execute executes the code editor with the given input messages.
	Execute(ctx context.Context, messages []llms.MessageContent) error
}
