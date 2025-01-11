package console

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var successStyle = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("2"))

// Success prints the given text with a success style (green color and bold).
// It adds top padding and automatically appends a newline.
// Example:
//   console.Success("Operation completed successfully")
func Success(text string) {
	fmt.Println(successStyle.Render(text))
}

// Successf prints formatted text with a success style (green color and bold).
// It adds top padding and automatically appends a newline.
// The format string follows fmt.Printf conventions.
// Example:
//   console.Successf("Processed %d items successfully", count)
func Successf(format string, args ...interface{}) {
	fmt.Println(successStyle.Render(fmt.Sprintf(format, args...)))
}
