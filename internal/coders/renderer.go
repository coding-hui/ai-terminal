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

func (r *Renderer) RenderWelcomeMessage(username string) string {
	welcome := "\n\nWelcome! **" + username + "** 👋  \n\n"

	return welcome
}

func (r *Renderer) RenderHelpMessage() string {
	help := "**Help**\n"
	help += "- `↑`/`↓`   : Navigate through history\n"
	help += "- `/add`    : Add files to the chat for editing or detailed review by GPT\n"
	help += "- `/list`   : List files currently added to the chat\n"
	help += "- `/remove` : Remove specific files from the chat\n"
	help += "- `/drop`   : Remove all files from the chat\n"
	help += "- `/coding` : Execute coding tasks on files added to the chat\n"
	help += "- `/ask`    : Ask a question about the files added to the chat\n"
	help += "- `ctrl+h`  : Show this help message\n"
	help += "- `ctrl+r`  : Clear the terminal and reset the discussion history\n"
	help += "- `ctrl+l`  : Clear the terminal while preserving the discussion history\n"
	help += "- `ctrl+c`  : Exit the application or interrupt command execution\n"

	return help
}
