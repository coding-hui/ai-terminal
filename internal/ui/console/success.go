package console

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var successStyle = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("2"))

// Success prints styled success message to standard output.
// The message is rendered in bold green text with top padding and automatically
// includes a trailing newline.
//
// Use this for indicating successful operations to users.
//
// Example:
//
//	console.Success("Operation completed successfully")
func Success(text string) {
	fmt.Println(successStyle.Render(text))
}

// Successf prints a formatted success message to standard output.
// The message is rendered in bold green text with top padding and automatically
// includes a trailing newline. Formatting follows fmt.Printf conventions.
//
// Use this for indicating parameterized successful operations to users.
//
// Example:
//
//	console.Successf("Processed %d items successfully", count)
func Successf(format string, args ...interface{}) {
	fmt.Println(successStyle.Render(fmt.Sprintf(format, args...)))
}
