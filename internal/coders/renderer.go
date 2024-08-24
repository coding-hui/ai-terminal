package coders

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultWidth  = 120
	defaultHeight = 120

	execColor    = "#ffa657"
	configColor  = "#ffffff"
	chatColor    = "#66b3ff"
	helpColor    = "#aaaaaa"
	errorColor   = "#cc3333"
	warningColor = "#ffcc00"
	successColor = "#46b946"
)

type Renderer struct {
	contentRenderer *glamour.TermRenderer
	successRenderer lipgloss.Style
	warningRenderer lipgloss.Style
	errorRenderer   lipgloss.Style
	helpRenderer    lipgloss.Style
}

func NewRenderer(options ...glamour.TermRendererOption) *Renderer {
	contentRenderer, err := glamour.NewTermRenderer(options...)
	if err != nil {
		return nil
	}

	successRenderer := lipgloss.NewStyle().Foreground(lipgloss.Color(successColor))
	warningRenderer := lipgloss.NewStyle().Foreground(lipgloss.Color(warningColor))
	errorRenderer := lipgloss.NewStyle().Foreground(lipgloss.Color(errorColor))
	helpRenderer := lipgloss.NewStyle().Foreground(lipgloss.Color(helpColor)).Italic(true)

	return &Renderer{
		contentRenderer: contentRenderer,
		successRenderer: successRenderer,
		warningRenderer: warningRenderer,
		errorRenderer:   errorRenderer,
		helpRenderer:    helpRenderer,
	}
}

func (r *Renderer) RenderContent(in string) string {
	contents := lipgloss.NewStyle().
		Width(defaultWidth). // pad to width.
		MaxWidth(defaultWidth).
		Render(in)
	out, _ := r.contentRenderer.Render(contents)
	return out
}

func (r *Renderer) RenderSuccess(in string) string {
	return r.successRenderer.Render(in)
}

func (r *Renderer) RenderWarning(in string) string {
	return r.warningRenderer.Render(in)
}

func (r *Renderer) RenderError(in string) string {
	return r.errorRenderer.Render(in)
}

func (r *Renderer) RenderHelp(in string) string {
	return r.helpRenderer.Render(in)
}

func (r *Renderer) RenderConfigMessage(username string) string {
	welcome := "Welcome! **" + username + "** 👋  \n\n"
	welcome += "I cannot find a configuration file, please enter an `LLM Model Name` you want to use"

	return welcome
}

func (r *Renderer) RenderApiTokenConfigMessage() string {
	return "Please enter an `API Token` for the model to access"
}

func (r *Renderer) RenderApiBaseConfigMessage() string {
	return "Please enter the inference `Api Base` for your model"
}

func (r *Renderer) RenderHelpMessage() string {
	help := "**Help**\n"
	help += "- `↑`/`↓`   : navigate in history\n"
	help += "- `/add`    : Add files to the chat so GPT can edit them or review them in detail\n"
	help += "- `/list`   : List files added to the chat\n"
	help += "- `/remove` : Remove files from the chat\n"
	help += "- `/drop`   : Drop all files from the chat\n"
	help += "- `/coding` : Execute coding tasks on added files\n"
	help += "- `/ask`    : Ask a question about the added files\n"
	help += "- `ctrl+h`  : show help\n"
	help += "- `ctrl+s`  : edit settings\n"
	help += "- `ctrl+r`  : clear terminal and reset discussion history\n"
	help += "- `ctrl+l`  : clear terminal but keep discussion history\n"
	help += "- `ctrl+c`  : exit or interrupt command execution\n"

	return help
}
