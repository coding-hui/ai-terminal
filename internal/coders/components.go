package coders

import "github.com/charmbracelet/glamour"

var (
	components = NewComponents()
)

type Components struct {
	prompt   *Prompt
	renderer *Renderer
	spinner  *Spinner
}

func NewComponents() *Components {
	return &Components{
		prompt:  NewPrompt(ChatPromptMode),
		spinner: NewSpinner(),
		renderer: NewRenderer(
			glamour.WithEmoji(),
			glamour.WithAutoStyle(),
			glamour.WithPreservedNewLines(),
			glamour.WithWordWrap(defaultWidth),
		),
	}
}
