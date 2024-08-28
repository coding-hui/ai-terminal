package coders

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// WaitFormUserConfirm is a model that displays a confirmation message to the user and waits for their response. It handles user input for confirmation (yes) or cancellation (no). This model uses a viewport to display the message and a channel to receive the user's choice.
type WaitFormUserConfirm struct {
	message string         // The message to display to the user.
	choice  chan bool      // Channel to receive the user's confirmation choice (true for yes, false for no).
	vp      viewport.Model // Viewport model to display the message.
}

func NewConfirmModel(message string) *WaitFormUserConfirm {
	formatMsg := components.renderer.RenderContent(fmt.Sprintf("\n\n %s \n", message))
	vp := viewport.New(defaultWidth, components.height-components.prompt.Height())
	vp.SetContent(formatMsg)
	return &WaitFormUserConfirm{
		message: message,
		vp:      vp,
		choice:  make(chan bool),
	}
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
		default:
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			return m, cmd
		}
	default:
		return m, nil
	}
}

func (m *WaitFormUserConfirm) View() string {
	return m.vp.View()
}

func (m *WaitFormUserConfirm) SetWidth(width int) {
	m.vp.Width = width
}

func (m *WaitFormUserConfirm) SetHeight(height int) {
	m.vp.Height = height
}

func (m *WaitFormUserConfirm) SetContent(content string) {
	m.vp.SetContent(content)
}

func (m *WaitFormUserConfirm) GotoBottom() {
	m.vp.GotoBottom()
}
