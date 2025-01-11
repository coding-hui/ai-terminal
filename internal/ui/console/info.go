package console

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// infoStyle defines the style for informational messages.
var infoStyle = lipgloss.NewStyle().
	Bold(false).
	PaddingTop(1).
	PaddingBottom(1)

// Info prints an informational message with the defined style.
func Info(text string) {
	fmt.Println(infoStyle.Render(text))
}

// Infof formats and prints an informational message with the defined style.
func Infof(format string, a ...interface{}) {
	Info(fmt.Sprintf(format, a...))
}
