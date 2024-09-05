// Package coders provides interfaces and types for working with code editors.
package coders

import (
	"context"
	"fmt"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/prompts"
)

// PartialCodeBlock represents a partial code block with its file path and original and updated text.
type PartialCodeBlock struct {
	// Path is the file path of the code block.
	Path string
	// OriginalText is the original text of the code block.
	OriginalText string
	// UpdatedText is the updated text of the code block.
	UpdatedText string
}

func (p PartialCodeBlock) String() string {
	return fmt.Sprintf("%s\n```<<<<<< Original\n%s\n====== Updated\n%s```",
		p.Path, p.OriginalText, p.UpdatedText,
	)
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
	// GetModifiedFiles retrieves the list of files that have been modified.
	GetModifiedFiles(ctx context.Context) ([]string, error)
	// UpdateCodeFences updates the code fence for the given code.
	UpdateCodeFences(ctx context.Context, code string) (string, string)
	// ApplyEdits applies the given list of edits to the code.
	ApplyEdits(ctx context.Context, edits []PartialCodeBlock, needConfirm bool) error
	// Execute runs the code editor with the specified input messages.
	Execute(ctx context.Context, messages []llms.MessageContent) error
}
