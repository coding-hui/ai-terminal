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
// The message is styled with a green foreground color (terminal color 2) and bold text.
// Example usage: console.Success("Operation completed successfully")
func Success(text string) {
	fmt.Println(successStyle.Render(text))
}

// Successf prints a formatted bold green success message with top padding.
// Uses fmt.Sprintf syntax for formatting and applies the same styling as Success.
// Example usage: console.Successf("Successfully processed %d items", count)
func Successf(format string, args ...interface{}) {
	fmt.Println(successStyle.Render(fmt.Sprintf(format, args...)))
}
