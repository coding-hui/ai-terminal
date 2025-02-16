package console

import (
	"fmt"
	"html"
	"os"
	"sync"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

var StdoutRenderer = sync.OnceValue(func() *lipgloss.Renderer {
	return lipgloss.DefaultRenderer()
})

var StdoutStyles = sync.OnceValue(func() Styles {
	return MakeStyles(StdoutRenderer())
})

var StderrRenderer = sync.OnceValue(func() *lipgloss.Renderer {
	return lipgloss.NewRenderer(os.Stderr, termenv.WithColorCache(true))
})

var StderrStyles = sync.OnceValue(func() Styles {
	return MakeStyles(StderrRenderer())
})

func Render(format string, args ...interface{}) {
	msg := StdoutStyles().AppName.Render(fmt.Sprintf(format, args...))
	fmt.Println(msg)
}

func RenderComment(format string, args ...interface{}) {
	msg := StdoutStyles().Comment.Render(fmt.Sprintf(format, args...))
	fmt.Println(msg)
}

// RenderStep renders commit process step messages with a prefix
func RenderStep(format string, args ...interface{}) {
	msg := StdoutStyles().CommitStep.Render(fmt.Sprintf("➤ "+format, args...))
	fmt.Println(msg)
}

// RenderSuccess renders successful commit messages
func RenderSuccess(format string, args ...interface{}) {
	msg := StdoutStyles().CommitSuccess.Render(fmt.Sprintf("✓ "+format, args...))
	fmt.Println(msg)
}

func RenderError(err error, reason string, args ...interface{}) {
	header := StderrStyles().ErrPadding.Render(StderrStyles().ErrorHeader.String(), err.Error())
	detail := StderrStyles().ErrPadding.Render(StderrStyles().ErrorDetails.Render(fmt.Sprintf(reason, args...)))
	_, _ = fmt.Printf("\n%s\n%s\n\n", header, detail)
}

func RenderAppName(appName string, suffix string, args ...interface{}) {
	appName = MakeGradientText(StdoutStyles().AppName, appName)
	fmt.Print(appName + " " + fmt.Sprintf(suffix, args...))
}

func RenderChatMessages(messages []llms.ChatMessage) error {
	content, err := llms.GetBufferString(messages, "", "human", "ai")
	if err != nil {
		return err
	}
	out := html.UnescapeString(content)
	fmt.Println(out)
	_ = clipboard.WriteAll(out)
	termenv.Copy(out)
	PrintConfirmation("COPIED", "The content copied to clipboard!")
	return nil
}
