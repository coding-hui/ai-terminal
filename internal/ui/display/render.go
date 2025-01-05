package display

import (
	"os"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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
