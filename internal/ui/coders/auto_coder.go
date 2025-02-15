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
}

func NewAutoCoder(opts ...AutoCoderOption) *AutoCoder {
	return applyAutoCoderOptions(opts...)
}

// saveContext persists a load context to the store
func (a *AutoCoder) saveContext(ctx context.Context, lc *convo.LoadContext) error {
	lc.ConversationID = a.cfg.CacheWriteToID
	return a.store.SaveContext(ctx, lc)
}

// deleteContext removes a load context from the store
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

	// Load contexts for this conversation
	contexts, err := a.store.ListContextsByteConvoID(context.Background(), details.WriteID)
	if err != nil {
		return errbook.Wrap("Failed to load conversation contexts", err)
	}

	// Convert to pointers and add to loaded contexts
	for _, ctx := range contexts {
		a.loadedContexts = append(a.loadedContexts, &ctx)
	}

	return nil
}

func (a *AutoCoder) Run() error {
	codingCmd := strings.TrimSpace(a.prompt) != ""
	if !codingCmd {
		a.printWelcome()
	}

	// Load any existing contexts from previous session
	if err := a.loadExistingContexts(); err != nil {
		return err
	}

	cmdExecutor := NewCommandExecutor(a)

	if codingCmd {
		cmdExecutor.Executor(fmt.Sprintf("/coding %s", a.prompt))
		return nil
	}

	cmdCompleter := NewCommandCompleter(a.repo)
	p := console.NewPrompt(
		a.cfg.AutoCoder.PromptPrefix,
		true,
		cmdCompleter.Complete,
		cmdExecutor.Executor,
	)

	// start prompt repl loop
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
