package ui

// PromptMode represents different modes of user interaction
type PromptMode int

const (
	// ExecPromptMode is for executing commands
	ExecPromptMode PromptMode = iota

	// TokenConfigPromptMode is for configuring API tokens
	TokenConfigPromptMode

	// ModelConfigPromptMode is for configuring models
	ModelConfigPromptMode

	// ApiBaseConfigPromptMode is for configuring API base URLs
	ApiBaseConfigPromptMode

	// ChatPromptMode is for chat interactions
	ChatPromptMode

	// DefaultPromptMode is the fallback mode
	DefaultPromptMode
)

// String returns the string representation of a PromptMode
func (m PromptMode) String() string {
	switch m {
	case ExecPromptMode:
		return "exec"
	case TokenConfigPromptMode:
		return "tokenConfig"
	case ModelConfigPromptMode:
		return "modelConfig"
	case ApiBaseConfigPromptMode:
		return "apiBaseConfig"
	case ChatPromptMode:
		return "ask"
	default:
		return "default"
	}
}

// GetPromptModeFromString converts a string to its corresponding PromptMode
func GetPromptModeFromString(s string) PromptMode {
	switch s {
	case "exec":
		return ExecPromptMode
	case "tokenConfig":
		return TokenConfigPromptMode
	case "modelConfig":
		return ModelConfigPromptMode
	case "apiBaseConfig":
		return ApiBaseConfigPromptMode
	case "ask":
		return ChatPromptMode
	default:
		return DefaultPromptMode
	}
}

// RunMode represents different runtime modes
type RunMode int

const (
	// CliMode is for command-line interface operation
	CliMode RunMode = iota

	// ReplMode is for read-eval-print-loop operation
	ReplMode
)

// String returns the string representation of a RunMode
func (m RunMode) String() string {
	if m == CliMode {
		return "cli"
	} else {
		return "repl"
	}
}

type Command struct {
	Name string
	Desc string
}
