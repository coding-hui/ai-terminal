package chat

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/ui"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
)

// HistoryWriter wraps console output functions to also write to chat history
type HistoryWriter struct {
	workingDir string
}

// NewHistoryWriter creates a new HistoryWriter instance
func NewHistoryWriter() *HistoryWriter {
	wd, _ := os.Getwd()
	return &HistoryWriter{
		workingDir: wd,
	}
}

// WriteToHistory writes content to the chat history file
func (h *HistoryWriter) WriteToHistory(content string) error {
	historyFilePath := filepath.Join(h.workingDir, chatHistoryFilename)

	// Build history content
	var historyContent strings.Builder

	// Add file header if this is the first write
	fileInfo, err := os.Stat(historyFilePath)
	if err != nil || fileInfo.Size() == 0 {
		historyContent.WriteString(fmt.Sprintf("# AI chat started at %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
		if len(os.Args) > 0 {
			historyContent.WriteString(fmt.Sprintf("> %s\n", strings.Join(os.Args, " ")))
		}
		// Note: Model info would need to be passed in or retrieved from context
		historyContent.WriteString("> Model: [model]\n")
		if _, err := os.Stat(".git"); err == nil {
			historyContent.WriteString("> Git repo: .git\n")
		}
		historyContent.WriteString("\n")
	}

	// Write the content
	if content != "" {
		historyContent.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			historyContent.WriteString("\n")
		}
		historyContent.WriteString("\n")
	}

	// Write to file
	file, err := os.OpenFile(historyFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return errbook.Wrap("Failed to open chat history file", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err := file.WriteString(historyContent.String()); err != nil {
		return errbook.Wrap("Failed to write chat history", err)
	}

	return nil
}

// Render writes to console and chat history
func (h *HistoryWriter) Render(format string, args ...interface{}) {
	console.Render(format, args...)
	_ = h.WriteToHistory(fmt.Sprintf(format, args...))
}

// RenderHelps writes to console and chat history
func (h *HistoryWriter) RenderHelps(commands []ui.Command) {
	var content strings.Builder
	for _, cmd := range commands {
		formatted := fmt.Sprintf("  %-44s", cmd.Name)
		console.Render(
			"%s%s",
			console.StdoutStyles().Flag.Render(formatted),
			console.StdoutStyles().FlagDesc.Render(cmd.Desc),
		)
		content.WriteString(fmt.Sprintf(">%s%s  \n", formatted, cmd.Desc))
	}
	_ = h.WriteToHistory(content.String())
}

// RenderError writes error to console and chat history
func (h *HistoryWriter) RenderError(err error, reason string, args ...interface{}) {
	console.RenderError(err, reason, args...)
	header := fmt.Sprintf("> ERROR: %s", err)
	detail := fmt.Sprintf("> "+reason, args...)
	errorMsg := fmt.Sprintf("%s  \n%s\n", header, detail)
	_ = h.WriteToHistory(errorMsg)
}

// RenderComment writes comment to console and chat history
func (h *HistoryWriter) RenderComment(format string, args ...interface{}) {
	console.RenderComment(format, args...)
	content := fmt.Sprintf(format, args...)
	_ = h.WriteToHistory("> " + content)
}

// RenderStep writes step information to console and chat history
func (h *HistoryWriter) RenderStep(format string, args ...interface{}) {
	// Assuming console has a RenderStep function, if not, use Render
	console.Render(format, args...)
	content := fmt.Sprintf(format, args...)
	_ = h.WriteToHistory(content)
}
