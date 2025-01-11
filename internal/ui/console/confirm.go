package console

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

func WaitForUserConfirm(format string, args ...interface{}) (confirm bool) {
	err := huh.Run(
		huh.NewConfirm().
			Title(fmt.Sprintf(format, args...)).
			Inline(true).
			Affirmative("Yes!").
			Negative("No.").
			Value(&confirm),
	)
	if err != nil {
		return false
	}
	return
}

func WaitForUserConfirmApplyChanges(format string, args ...interface{}) (confirm bool) {
	return true
}
