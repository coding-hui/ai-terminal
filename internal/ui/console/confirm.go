package console

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/confirmation"
)

var No = confirmation.No
var Yes = confirmation.Yes

func WaitForUserConfirm(defaultVal confirmation.Value, format string, args ...interface{}) bool {
	input := confirmation.New(fmt.Sprintf(format, args...), defaultVal)
	confirm, err := input.RunPrompt()
	if err != nil {
		return false
	}
	return confirm
}

// action messages

const defaultAction = "WROTE"

var outputHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#6C50FF")).Bold(true).Padding(0, 1).MarginRight(1)

func PrintConfirmation(action, content string) {
	if action == "" {
		action = defaultAction
	}
	outputHeader = outputHeader.SetString(strings.ToUpper(action))
	fmt.Println(lipgloss.JoinHorizontal(lipgloss.Center, outputHeader.String(), content))
}
