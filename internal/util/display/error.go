package display

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("9"))

// Error handles and displays an error message along with optional additional messages, if any.
func Error(err error, msgs ...string) {
	if err == nil {
		return
	}

	errMsg := err.Error()
	if errMsg == "" {
		return
	}

	if len(msgs) > 0 {
		ErrorMsg(msgs...)
	}
	ErrorMsg(err.Error())
}

// ErrorMsg displays one or more error messages using a predefined style.
func ErrorMsg(msgs ...string) {
	for _, msg := range msgs {
		fmt.Println(style.Render(msg))
	}
}

// FatalErr handles and displays an error message along with optional additional messages, if any, then exits the program with a status code of 1.
func FatalErr(err error, msgs ...string) {
	Error(err, msgs...)
	os.Exit(1)
}
