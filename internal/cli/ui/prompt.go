package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	execIcon                 = "🚀 > "
	configIcon               = "🔒 > "
	botIcon                  = "🤖 > "
	apiBaseIcon              = "🛠️ > "
	chatIcon                 = "💬 > "
	execPlaceholder          = "Execute something..."
	chatPlaceholder          = "Ask me something..."
	modelConfigPlaceholder   = "Enter your LLM Model Name..."
	apiBaseConfigPlaceholder = "Enter your LLM Api Base..."
	tokenConfigPlaceholder   = "Enter your LLM API Token..."
)

type Prompt struct {
	mode  PromptMode
	input textinput.Model
}

func NewPrompt(mode PromptMode) *Prompt {
	input := textinput.New()
	input.Placeholder = getPromptPlaceholder(mode)
	input.TextStyle = getPromptStyle(mode)
	input.Prompt = getPromptIcon(mode)

	if mode == TokenConfigPromptMode {
		input.EchoMode = textinput.EchoPassword
	}

	input.Focus()

	return &Prompt{
		mode:  mode,
		input: input,
	}
}

func (p *Prompt) GetMode() PromptMode {
	return p.mode
}

func (p *Prompt) SetMode(mode PromptMode) *Prompt {
	p.mode = mode

	p.input.TextStyle = getPromptStyle(mode)
	p.input.Prompt = getPromptIcon(mode)
	p.input.Placeholder = getPromptPlaceholder(mode)

	return p
}

func (p *Prompt) SetValue(value string) *Prompt {
	p.input.SetValue(value)

	return p
}

func (p *Prompt) GetValue() string {
	return p.input.Value()
}

func (p *Prompt) Blur() *Prompt {
	p.input.Blur()

	return p
}

func (p *Prompt) Focus() *Prompt {
	p.input.Focus()

	return p
}

func (p *Prompt) Update(msg tea.Msg) (*Prompt, tea.Cmd) {
	var updateCmd tea.Cmd
	p.input, updateCmd = p.input.Update(msg)

	return p, updateCmd
}

func (p *Prompt) View() string {
	return p.input.View()
}

func (p *Prompt) AsString() string {
	style := getPromptStyle(p.mode)

	return fmt.Sprintf("%s%s", style.Render(getPromptIcon(p.mode)), style.Render(p.input.Value()))
}

func getPromptStyle(mode PromptMode) lipgloss.Style {
	switch mode {
	case ExecPromptMode:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(execColor))
	case ChatPromptMode:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(chatColor))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(configColor))
	}
}

func getPromptIcon(mode PromptMode) string {
	style := getPromptStyle(mode)

	switch mode {
	case ExecPromptMode:
		return style.Render(execIcon)
	case ChatPromptMode:
		return style.Render(chatIcon)
	case ModelConfigPromptMode:
		return style.Render(botIcon)
	case ApiBaseConfigPromptMode:
		return style.Render(apiBaseIcon)
	default:
		return style.Render(configIcon)
	}
}

func getPromptPlaceholder(mode PromptMode) string {
	switch mode {
	case ExecPromptMode:
		return execPlaceholder
	case TokenConfigPromptMode:
		return tokenConfigPlaceholder
	case ModelConfigPromptMode:
		return modelConfigPlaceholder
	case ApiBaseConfigPromptMode:
		return apiBaseConfigPlaceholder
	default:
		return chatPlaceholder
	}
}
