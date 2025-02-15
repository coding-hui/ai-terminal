package ai

import "github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

type EngineMode int

const (
	ChatEngineMode EngineMode = iota
	ExecEngineMode
)

func (m EngineMode) String() string {
	if m == ExecEngineMode {
		return "exec"
	} else {
		return "chat"
	}
}

type CompletionOutput struct {
	Command     string `json:"cmd"`
	Explanation string `json:"exp"`
	Executable  bool   `json:"exec"`

	Usage llms.Usage `json:"usage"`
}

func (c CompletionOutput) GetCommand() string {
	return c.Command
}

func (c CompletionOutput) GetExplanation() string {
	return c.Explanation
}

func (c CompletionOutput) IsExecutable() bool {
	return c.Executable
}

// CompletionInput is a tea.Msg that wraps the content read from stdin.
type CompletionInput struct {
	Messages []llms.ChatMessage
}

// StreamCompletionOutput a tea.Msg that wraps the content returned from ai.
type StreamCompletionOutput struct {
	Content    string
	Last       bool
	Interrupt  bool
	Executable bool

	Usage llms.Usage `json:"usage"`
}

func (c StreamCompletionOutput) GetContent() string {
	return c.Content
}

func (c StreamCompletionOutput) IsLast() bool {
	return c.Last
}

func (c StreamCompletionOutput) IsInterrupt() bool {
	return c.Interrupt
}

func (c StreamCompletionOutput) IsExecutable() bool {
	return c.Executable
}

func (c StreamCompletionOutput) GetUsage() llms.Usage {
	return c.Usage
}
