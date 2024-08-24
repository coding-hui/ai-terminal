package coders

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type WaitFormUserConfirm struct {
	message string
	choice  chan bool
}

func NewConfirmModel(message string) *WaitFormUserConfirm {
	return &WaitFormUserConfirm{
		message: message,
		choice:  make(chan bool),
	}
}

func (m *WaitFormUserConfirm) Init() tea.Cmd {
	return nil
}

func (m *WaitFormUserConfirm) Update(msg tea.Msg) (*WaitFormUserConfirm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.choice <- true
			return m, textinput.Blink
		case "n", "N":
			m.choice <- false
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m *WaitFormUserConfirm) View() string {
	return fmt.Sprintf("\n\n %s \n", m.message)
}
