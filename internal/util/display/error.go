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

func ErrorMsg(msgs ...string) {
	for _, msg := range msgs {
		fmt.Println(style.Render(msg))
	}
}

func FatalErr(err error, msgs ...string) {
	Error(err, msgs...)
	os.Exit(1)
}
