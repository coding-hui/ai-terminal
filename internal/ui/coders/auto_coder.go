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
	console.RenderAppName("AutoCoder", "%s\n", a.versionInfo.GitVersion)
	console.Render("Model: %s with %s format", a.cfg.Model, a.cfg.AutoCoder.EditFormat)
	console.Render("Please use `exit` or `Ctrl-C` to exit this program.")
}
