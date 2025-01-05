package errbook

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/coding-hui/ai-terminal/internal/ui/display"
	"github.com/coding-hui/ai-terminal/internal/util/term"
)

func HandleError(err error) {
	// exhaust stdin
	if !term.IsInputTTY() {
		_, _ = io.ReadAll(os.Stdin)
	}

	format := "\n%s\n\n"

	var args []interface{}
	var aiTermErr AiError
	if errors.As(err, &aiTermErr) {
		args = []interface{}{
			display.StderrStyles().ErrPadding.Render(display.StderrStyles().ErrorHeader.String(), aiTermErr.reason),
		}

		// Skip the errbook details if the user simply canceled out of huh.
		if !errors.Is(aiTermErr.err, ErrUserAborted) {
			format += "%s\n\n"
			args = append(args, display.StderrStyles().ErrPadding.Render(display.StderrStyles().ErrorDetails.Render(err.Error())))
		}
	} else {
		args = []interface{}{
			display.StderrStyles().ErrPadding.Render(display.StderrStyles().ErrorDetails.Render(err.Error())),
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}
