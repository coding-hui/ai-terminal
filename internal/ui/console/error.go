package console

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("9"))

// Deprecated: uses console.StderrRenderer()
// Error handles and displays an error message. It formats the error message using the provided arguments.
func Error(args interface{}) {
	Errorf("%s", args)
}

// Deprecated: uses console.StderrRenderer()
// Errorf formats and displays an error message using the provided format string and arguments.
func Errorf(format string, args ...interface{}) {
	fmt.Println(style.Render(fmt.Sprintf(format, args...)))
}

// Deprecated: uses console.StderrRenderer()
// Fatal handles and displays a fatal error message. It then exits the program with a status code of 1.
func Fatal(args interface{}) {
	Error(args)
	os.Exit(1)
}

// Deprecated: uses console.StderrRenderer()
// Fatalf formats and displays a fatal error message using the provided format string and arguments. It then exits the program with a status code of 1.
func Fatalf(format string, args ...interface{}) {
	Errorf(format, args...)
	os.Exit(1)
}
