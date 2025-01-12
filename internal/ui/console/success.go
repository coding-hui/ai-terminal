package console

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var successStyle = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("2"))

// Success prints a bold green success message with top padding to indicate successful operations.
// The message is styled with a green foreground color and bold text for emphasis.
// Example: console.Success("Operation completed successfully")
func Success(text string) {
	fmt.Println(successStyle.Render(text))
}

// Successf prints a formatted bold green success message with top padding.
// Similar to Success but supports formatting with fmt.Sprintf syntax.
// Example: console.Successf("Successfully processed %d items", count)
func Successf(format string, args ...interface{}) {
	fmt.Println(successStyle.Render(fmt.Sprintf(format, args...)))
}
