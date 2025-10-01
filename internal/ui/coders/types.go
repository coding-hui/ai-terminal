package coders

const (
	FlagVerbose = "verbose"
	FlagYes     = "yes"
)

// PromptMode represents the mode of the prompt.
type PromptMode int

// Constants representing different prompt modes.
const (
	ExecPromptMode    PromptMode = iota // ExecPromptMode represents the execution mode.
	ChatPromptMode                      // ChatPromptMode represents the chat mode.
	DefaultPromptMode                   // DefaultPromptMode represents the default mode.
)

// String returns the string representation of the PromptMode.
func (m PromptMode) String() string {
	switch m {
	case ExecPromptMode:
		return "exec"
	case ChatPromptMode:
		return "chat"
	default:
		return "default"
	}
}

// GetPromptModeFromString returns the PromptMode corresponding to the given string.
func GetPromptModeFromString(s string) PromptMode {
	switch s {
	case "exec":
		return ExecPromptMode
	case "chat":
		return ChatPromptMode
	default:
		return DefaultPromptMode
	}
}
