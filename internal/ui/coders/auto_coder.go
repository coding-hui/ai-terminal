package coders

import (
	"path/filepath"

	"github.com/coding-hui/common/version"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
)

type AutoCoder struct {
	codeBasePath string
	repo         *git.Command
	absFileNames map[string]struct{}
	engine       *llm.Engine

	versionInfo version.Info
	cfg         *options.Config
}

func NewAutoCoder(cfg *options.Config) (*AutoCoder, error) {
	repo := git.New()
	root, _ := repo.GitDir()

	engine, err := llm.NewLLMEngine(llm.ChatEngineMode, cfg)
	if err != nil {
		return nil, errbook.Wrap("Could not initialized llm engine", err)
	}

	return &AutoCoder{
		codeBasePath: filepath.Dir(root),
		repo:         repo,
		cfg:          cfg,
		engine:       engine,
		versionInfo:  version.Get(),
		absFileNames: map[string]struct{}{},
	}, nil
}

func (a *AutoCoder) Run() error {
	a.printWelcome()

	cmdExecutor := NewCommandExecutor(a)
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
	console.Render("==============================================")
	console.RenderAppName("AutoCoder", " %s\n", a.versionInfo.GitVersion)
	console.Render("==============================================")
	console.Render("Welcome to AutoCoder - Your AI Coding Assistant!")
	console.Render("")
	console.Render("Configuration:")
	console.Render("  • Model: %s", a.cfg.Model)
	console.Render("  • Format: %s", a.cfg.AutoCoder.EditFormat)
	console.Render("")
	console.Render("Recommended Workflow:")
	console.Render("  1. /add <file> - Add files to work on")
	console.Render("  2. /coding <request> - Request code changes")
	console.Render("  3. /ask <question> - Ask questions about code")
	console.Render("  4. /commit - Commit changes when ready")
	console.Render("")
	console.Render("Quick Tips:")
	console.Render("  • Type your /coding requests directly")
	console.Render("  • Use `/help` to see all commands")
	console.Render("  • Use `/exit` or `Ctrl-C` to quit")
	console.Render("")
	console.Render("Let's start coding! 🚀")
	console.Render("==============================================")
}
