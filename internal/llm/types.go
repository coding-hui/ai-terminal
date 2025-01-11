package llm

import "github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

type EngineMode int

const (
	ExecEngineMode EngineMode = iota
	ChatEngineMode
)

func (m EngineMode) String() string {
	if m == ExecEngineMode {
		return "exec"
	} else {
		return "chat"
	}
}

type SummaryContentOutput struct {
	Content string `json:"content"`
}

type EngineExecOutput struct {
	Command     string `json:"cmd"`
	Explanation string `json:"exp"`
	Executable  bool   `json:"exec"`
}

func (eo EngineExecOutput) GetCommand() string {
	return eo.Command
}

func (eo EngineExecOutput) GetExplanation() string {
	return eo.Explanation
}

func (eo EngineExecOutput) IsExecutable() bool {
	return eo.Executable
}

// CompletionInput is a tea.Msg that wraps the content read from stdin.
type CompletionInput struct {
	Messages []llms.ChatMessage
}

// StreamCompletionOutput a tea.Msg that wraps the content returned from llm.
type StreamCompletionOutput struct {
	content    string
	last       bool
	interrupt  bool
	executable bool
}

func (co StreamCompletionOutput) GetContent() string {
	return co.content
}

func (co StreamCompletionOutput) IsLast() bool {
	return co.last
}

func (co StreamCompletionOutput) IsInterrupt() bool {
	return co.interrupt
}

func (co StreamCompletionOutput) IsExecutable() bool {
	return co.executable
}

type EngineLoggingCallbackOutput struct {
	content string
}

func (ec EngineLoggingCallbackOutput) GetContent() string {
	return ec.content
}
