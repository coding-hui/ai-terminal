package coders

type PromptMode int

const (
	ExecPromptMode PromptMode = iota
	TokenConfigPromptMode
	ModelConfigPromptMode
	ApiBaseConfigPromptMode
	ChatPromptMode
	DefaultPromptMode
)

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

type RunMode int

const (
	CliMode RunMode = iota
	ReplMode
)

func (m RunMode) String() string {
	if m == CliMode {
		return "cli"
	} else {
		return "repl"
	}
}
