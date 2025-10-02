package chat

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coding-hui/ai-terminal/internal/errbook"
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
	defer file.Close()

	if _, err := file.WriteString(historyContent.String()); err != nil {
		return errbook.Wrap("Failed to write chat history", err)
	}

	return nil
}

// Render writes to console and chat history
func (h *HistoryWriter) Render(format string, args ...interface{}) {
	content := fmt.Sprintf(format, args...)
	console.Render(content)
	h.WriteToHistory(content)
}

// RenderError writes error to console and chat history
func (h *HistoryWriter) RenderError(err error, reason string, args ...interface{}) {
	console.RenderError(err, reason, args...)
	errorMsg := fmt.Sprintf("Error: %s - %s", err.Error(), fmt.Sprintf(reason, args...))
	// Format error with special markdown
	formattedError := fmt.Sprintf("**Error:** %s\n", errorMsg)
	h.WriteToHistory(formattedError)
}

// RenderComment writes comment to console and chat history
func (h *HistoryWriter) RenderComment(format string, args ...interface{}) {
	console.RenderComment(format, args...)
	content := fmt.Sprintf(format, args...)
	h.WriteToHistory(content)
}

// RenderStep writes step information to console and chat history
func (h *HistoryWriter) RenderStep(format string, args ...interface{}) {
	// Assuming console has a RenderStep function, if not, use Render
	console.Render(format, args...)
	content := fmt.Sprintf(format, args...)
	h.WriteToHistory(content)
}
