package console

import (
	"fmt"

	"github.com/erikgeiser/promptkit/confirmation"
)

var No = confirmation.No
var Yes = confirmation.Yes

func WaitForUserConfirm(defaultVal confirmation.Value, format string, args ...interface{}) bool {
	input := confirmation.New(fmt.Sprintf(format, args...), defaultVal)
	confirm, err := input.RunPrompt()
	if err != nil {
		return false
	}
	return confirm
}
