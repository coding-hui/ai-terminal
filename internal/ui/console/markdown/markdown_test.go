package markdown

import (
	"testing"

	"github.com/fatih/color"
)

func TestMarkdownRendering(t *testing.T) {
	doc := "I'll help improve the comments in the code to be more descriptive and consistent in English. Here are the *SEARCH/REPLACE* blocks:\n\n```go\ninternal/ai/ai.go\n<<<<<<< SEARCH\n// Engine represents the core AI engine that handles completions and streaming\n=======\n// Engine represents the core AI engine that handles completions and streaming.\n// It manages conversation state, model interactions, and output channels.\n>>>>>>> REPLACE\n```\n\n```go\ninternal/ai/ai.go\n<<<<<<< SEARCH\n// New creates a new Engine instance with the provided options\n=======\n// New creates a new Engine instance with the provided options.\n// Returns an initialized Engine and any error that occurred during setup.\n>>>>>>> REPLACE\n```\n\n```go\ninternal/ai/ai.go\n<<<<<<< SEARCH\n// CreateCompletion generates a single completion response for the given messages\n=======\n// CreateCompletion generates a single completion response for the given messages.\n// Handles context setup, model interaction, and response processing.\n// Returns the completion output or any error that occurred.\n>>>>>>> REPLACE\n```\n\n```go\ninternal/ai/ai.go\n<<<<<<< SEARCH\n// CreateStreamCompletion generates a streaming completion response\n=======\n// CreateStreamCompletion generates a streaming completion response.\n// Handles real-time output through channels and manages conversation state.\n// Returns the final stream output or any error that occurred.\n>>>>>>> REPLACE\n```\n\n```go\ninternal/ai/ai.go\n<<<<<<< SEARCH\n// callOptions builds the LLM call options based on engine configuration\n=======\n// callOptions builds the LLM call options based on engine configuration.\n// Includes model settings, streaming callbacks, and generation parameters.\n// Returns a slice of configured llms.CallOption values.\n>>>>>>> REPLACE\n```\n\n```go\ninternal/ai/ai.go\n<<<<<<< SEARCH\n// setupChatContext prepares the chat context by loading conversation history\n=======\n// setupChatContext prepares the chat context by loading conversation history.\n// Handles cache retrieval and message preparation.\n// Returns an error if context setup fails.\n>>>>>>> REPLACE\n```\n\n```go\ninternal/ai/ai.go\n<<<<<<< SEARCH\n// appendAssistantMessage adds the assistant's response to the conversation store\n=======\n// appendAssistantMessage adds the assistant's response to the conversation store.\n// Handles message persistence and error logging.\n>>>>>>> REPLACE\n```\n\n```go\ninternal/ai/ai.go\n<<<<<<< SEARCH\n// convert transforms a ChatMessage to MessageContent format\n=======\n// convert transforms a ChatMessage to MessageContent format.\n// Handles message type conversion for LLM compatibility.\n// Returns the converted MessageContent structure.\n>>>>>>> REPLACE\n```"

	// Create renderer with options
	renderer := NewTerminalRenderer(
		WithTheme("default"),
		WithCodeTheme("monokai"),
		WithBackground(color.BgBlack),
		WithHeaderStyle(color.New(color.FgHiCyan, color.Bold)),
		WithListStyle(color.New(color.FgHiGreen)),
		WithTextEmphasis(color.New(color.Bold)),
		WithRenderMode("full"),
	)
	defer renderer.Finalize()

	// Render document
	renderDocument(doc, renderer)
}
