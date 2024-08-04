package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg error

type TextareaModel struct {
	Textarea textarea.Model
	Err      error
}

func InitialTextareaPrompt(value string) TextareaModel {
	ti := textarea.New()
	ti.InsertString(value)
	ti.SetWidth(80)
	ti.SetHeight(len(strings.Split(value, "\n")))
	ti.Focus()

	return TextareaModel{
		Textarea: ti,
		Err:      nil,
	}
}

func (m TextareaModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m TextareaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type { //nolint:exhaustive
		case tea.KeyEsc:
			if m.Textarea.Focused() {
				m.Textarea.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			if !m.Textarea.Focused() {
				cmd = m.Textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.Err = msg
		return m, nil
	}

	m.Textarea, cmd = m.Textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m TextareaModel) View() string {
	return fmt.Sprintf(
		"Please confirm the following commit message.\n\n%s\n\n%s",
		m.Textarea.View(),
		"(ctrl+c to continue.)",
	) + "\n\n"
}
