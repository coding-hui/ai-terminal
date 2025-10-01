package coders

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/coding-hui/common/version"

	"github.com/coding-hui/ai-terminal/internal/ai"
	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
)

//go:embed banner.txt
var banner string

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
	initial := strings.TrimSpace(a.prompt)
	if initial == "" {
		a.printWelcome()
	}

	// Load any existing contexts from previous session
	if err := a.loadExistingContexts(); err != nil {
		return err
	}

	cmdExecutor := NewCommandExecutor(a)
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
		cmdExecutor.Executor(initial)
	}

	cmdCompleter := NewCommandCompleter(a.repo)
	// Choose prompt prefix based on active mode with fallbacks
	promptPrefix := a.cfg.AutoCoder.PromptPrefix
	switch a.promptMode {
	case ChatPromptMode:
		promptPrefix = "chat"
		if a.cfg.AutoCoder.PromptPrefixChat != "" {
			promptPrefix = a.cfg.AutoCoder.PromptPrefixChat
		}
	case ExecPromptMode:
		promptPrefix = "exec"
		if a.cfg.AutoCoder.PromptPrefixExec != "" {
			promptPrefix = a.cfg.AutoCoder.PromptPrefixExec
		}
	default:
		if a.cfg.AutoCoder.PromptPrefixCoding != "" {
			promptPrefix = a.cfg.AutoCoder.PromptPrefixCoding
		}
	}

	p := console.NewPrompt(
		promptPrefix,
		true,
		cmdCompleter.Complete,
		cmdExecutor.Executor,
	)

	// Start the interactive REPL (Read-Eval-Print Loop) for command processing
	p.Run()

	return nil
}

func (a *AutoCoder) printWelcome() {
	fmt.Println(banner)
	console.RenderComment("")
	console.RenderComment("Welcome to AutoCoder - Your AI Coding Assistant! (%s) [Model: %s]\n", a.versionInfo.GitVersion, a.cfg.CurrentModel.Name)

	// Get current conversation info from config
	if a.cfg.CacheWriteToID != "" {
		console.RenderComment("Current Session:")
		console.RenderComment("  • ID: %s", a.cfg.CacheWriteToID)
		if a.cfg.CacheWriteToTitle != "" {
			console.RenderComment("  • Title: %s", a.cfg.CacheWriteToTitle)
		}
		console.RenderComment("")
	}

	console.Render("Let's start coding! 🚀")
	console.RenderComment("")
}

func (a *AutoCoder) determineBeatCodeFences(rawCode string) (string, string) {
	if len(a.cfg.AutoCoder.GetDefaultFences()) == 2 {
		f := a.cfg.AutoCoder.GetDefaultFences()
		return f[0], f[1]
	}
	return chooseBestFence(rawCode)
}
