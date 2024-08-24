package coders

type PromptMode int

const (
	ExecPromptMode PromptMode = iota
	ChatPromptMode
	DefaultPromptMode
)

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
