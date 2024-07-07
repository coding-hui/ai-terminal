package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	execIcon          = "🚀 > "
	execPlaceholder   = "Execute something..."
	configIcon        = "🔒 > "
	configPlaceholder = "Enter your OpenAI key..."
	chatIcon          = "💬 > "
	chatPlaceholder   = "Ask me something..."
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

	if mode == ConfigPromptMode {
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
	case ConfigPromptMode:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(configColor))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(chatColor))
	}
}

func getPromptIcon(mode PromptMode) string {
	style := getPromptStyle(mode)

	switch mode {
	case ExecPromptMode:
		return style.Render(execIcon)
	case ConfigPromptMode:
		return style.Render(configIcon)
	default:
		return style.Render(chatIcon)
	}
}

func getPromptPlaceholder(mode PromptMode) string {
	switch mode {
	case ExecPromptMode:
		return execPlaceholder
	case ConfigPromptMode:
		return configPlaceholder
	default:
		return chatPlaceholder
	}
}
