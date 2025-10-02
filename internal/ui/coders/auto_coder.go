package coders

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coding-hui/ai-terminal/internal/ai"
	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/chat"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/common/version"
)

//go:embed banner.txt
var banner string

// Define constants for chat history functionality
const (
	chatHistoryFilename = ".ai.chat.history.md"
)

type AutoCoder struct {
	codeBasePath, prompt string
	repo                 *git.Command
	loadedContexts       []*convo.LoadContext
	engine               *ai.Engine
	store                convo.Store

	versionInfo version.Info
	cfg         *options.Config
	promptMode  PromptMode
}

func NewAutoCoder(opts ...AutoCoderOption) *AutoCoder {
	return applyAutoCoderOptions(opts...)
}

// saveContext persists the conversation context to the store for future reference
func (a *AutoCoder) saveContext(ctx context.Context, lc *convo.LoadContext) error {
	lc.ConversationID = a.cfg.CacheWriteToID
	return a.store.SaveContext(ctx, lc)
}

// deleteContext removes a specific conversation context from the store by its ID
func (a *AutoCoder) deleteContext(ctx context.Context, id uint64) error {
	return a.store.DeleteContexts(ctx, id)
}

// writeChatHistory writes commands and responses to the chat history file
func (a *AutoCoder) writeChatHistory(command, response string) error {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return errbook.Wrap("Failed to get current working directory", err)
	}

	historyFilePath := filepath.Join(wd, chatHistoryFilename)

	// Build history content
	var historyContent strings.Builder

	// Add file header if this is the first write
	fileInfo, err := os.Stat(historyFilePath)
	if err != nil || fileInfo.Size() == 0 {
		historyContent.WriteString(fmt.Sprintf("# AI chat started at %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
		// Add command invocation details if available
		if len(os.Args) > 0 {
			historyContent.WriteString(fmt.Sprintf("> %s\n", strings.Join(os.Args, " ")))
		}
		// Add model info
		historyContent.WriteString(fmt.Sprintf("> Model: %s\n", a.cfg.Model))
		// Add git repo info if available
		// Check if we're in a git repository
		if _, err := os.Stat(".git"); err == nil {
			historyContent.WriteString("> Git repo: .git\n")
		}
		historyContent.WriteString("\n")
	}

	// Write command and response
	if command != "" {
		// Format commands with '>' prefix or '####' for slash commands
		if strings.HasPrefix(command, "/") {
			historyContent.WriteString(fmt.Sprintf("#### %s\n", command))
		} else {
			// Add prompt prefix before the command
			promptPrefix := a.getPromptPrefix()
			historyContent.WriteString(fmt.Sprintf("> %s %s\n", promptPrefix, command))
		}
		// Add a newline after command
		historyContent.WriteString("\n")
	}

	if response != "" {
		// Format responses that indicate file operations
		if strings.Contains(response, "Added") && strings.Contains(response, "to the chat") {
			historyContent.WriteString(fmt.Sprintf("> %s\n", response))
		} else {
			// For other responses, just write them directly
			historyContent.WriteString(response)
			if !strings.HasSuffix(response, "\n") {
				historyContent.WriteString("\n")
			}
		}
		// Add a newline after response
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

func (a *AutoCoder) loadExistingContexts() error {
	// Get current conversation details
	details, err := convo.GetCurrentConversationID(context.Background(), a.cfg, a.store)
	if err != nil {
		return errbook.Wrap("Failed to get current conversation", err)
	}

	a.cfg.CacheWriteToID = details.WriteID
	a.cfg.CacheWriteToTitle = details.Title
	a.cfg.CacheReadFromID = details.ReadID
	a.cfg.Model = details.Model

	// Load all conversation contexts associated with the current session
	contexts, err := a.store.ListContextsByteConvoID(context.Background(), details.WriteID)
	if err != nil {
		return errbook.Wrap("Failed to load conversation contexts", err)
	}

	// Convert loaded contexts to pointers and store them in the AutoCoder instance
	for _, ctx := range contexts {
		a.loadedContexts = append(a.loadedContexts, &ctx)
	}

	return nil
}

func (a *AutoCoder) Run() error {
	// Create history writer
	historyWriter := chat.NewHistoryWriter()

	initial := strings.TrimSpace(a.prompt)
	if initial == "" {
		a.printWelcome(historyWriter)
	}

	// Load any existing contexts from previous session
	if err := a.loadExistingContexts(); err != nil {
		return err
	}

	cmdExecutor := NewCommandExecutor(a, historyWriter)
	if initial != "" {
		// If the provided prompt is a slash/exec command, run it directly.
		// Otherwise, map by promptMode:
		if !strings.HasPrefix(initial, "/") && !strings.HasPrefix(initial, "!") {
			switch a.promptMode {
			case ChatPromptMode:
				initial = fmt.Sprintf("/ask %s", initial)
			case ExecPromptMode:
				initial = fmt.Sprintf("/exec %s", initial)
			default:
				initial = fmt.Sprintf("/coding %s", initial)
			}
		}

		// Record initial command to history
		if err := a.writeChatHistory(initial, ""); err != nil {
			historyWriter.RenderError(err, "Failed to write initial command to history")
		}

		cmdExecutor.Executor(initial)
	}

	cmdCompleter := NewCommandCompleter(a.repo)

	p := console.NewPrompt(
		true,
		cmdCompleter.Complete,
		func(input string) {
			// Record command to history before execution
			if err := a.writeChatHistory(input, ""); err != nil {
				historyWriter.RenderError(err, "Failed to write command to history")
			}
			cmdExecutor.Executor(input)
		},
		func() string {
			return a.getPromptPrefix() + " > "
		},
	)

	// Start the interactive REPL (Read-Eval-Print Loop) for command processing
	p.Run()

	return nil
}

func (a *AutoCoder) getPromptPrefix() (promptPrefix string) {
	promptPrefix = a.promptMode.String()
	switch a.promptMode {
	case ChatPromptMode:
		if a.cfg.AutoCoder.PromptPrefixChat != "" {
			promptPrefix = a.cfg.AutoCoder.PromptPrefixChat
		}
	case ExecPromptMode:
		if a.cfg.AutoCoder.PromptPrefixExec != "" {
			promptPrefix = a.cfg.AutoCoder.PromptPrefixExec
		}
	default:
		if a.cfg.AutoCoder.PromptPrefixCoding != "" {
			promptPrefix = a.cfg.AutoCoder.PromptPrefixCoding
		}
	}
	return promptPrefix
}

func (a *AutoCoder) printWelcome(historyWriter *chat.HistoryWriter) {
	fmt.Println(banner)
	historyWriter.RenderComment("")
	historyWriter.RenderComment("Welcome to AutoCoder - Your AI Coding Assistant! (%s) [Model: %s]\n", a.versionInfo.GitVersion, a.cfg.CurrentModel.Name)

	// Get current conversation info from config
	if a.cfg.CacheWriteToID != "" {
		historyWriter.RenderComment("Current Session:")
		historyWriter.RenderComment("  • ID: %s", a.cfg.CacheWriteToID)
		if a.cfg.CacheWriteToTitle != "" {
			historyWriter.RenderComment("  • Title: %s", a.cfg.CacheWriteToTitle)
		}
		historyWriter.RenderComment("")
	}

	historyWriter.Render("Let's start coding! 🚀")
	historyWriter.RenderComment("")
}

func (a *AutoCoder) determineBeatCodeFences(rawCode string) (string, string) {
	if len(a.cfg.AutoCoder.GetDefaultFences()) == 2 {
		f := a.cfg.AutoCoder.GetDefaultFences()
		return f[0], f[1]
	}
	return chooseBestFence(rawCode)
}
