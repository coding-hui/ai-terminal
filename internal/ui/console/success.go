package console

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var successStyle = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("2"))

// Success prints a bold green success message with top padding.
// Example: console.Success("Operation completed")
func Success(text string) {
	fmt.Println(successStyle.Render(text))
}

// Successf prints a formatted bold green success message.
// Example: console.Successf("Processed %d items", count)
func Successf(format string, args ...interface{}) {
	fmt.Println(successStyle.Render(fmt.Sprintf(format, args...)))
}
