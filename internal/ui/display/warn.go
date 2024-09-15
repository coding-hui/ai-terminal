package display

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var warnStyle = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("3"))

// Warn prints the given text with a warning style.
func Warn(text string) {
	fmt.Println(warnStyle.Render(text))
}

// Warnf prints the formatted string with a warning style.
func Warnf(format string, args ...interface{}) {
	fmt.Println(warnStyle.Render(fmt.Sprintf(format, args...)))
}
