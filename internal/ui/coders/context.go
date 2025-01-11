package coders

import (
	"context"
	"path/filepath"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/options"
)

type CoderContext struct {
	context.Context

	codeBasePath string
	repo         *git.Command
	absFileNames map[string]struct{}
	engine       *llm.Engine
}

func NewCoderContext(cfg *options.Config) (*CoderContext, error) {
	repo := git.New()
	root, _ := repo.GitDir()
	engine, err := llm.NewLLMEngine(llm.ChatEngineMode, cfg)
	if err != nil {
		return nil, errbook.Wrap("Could not initialized llm engine", err)
	}
	return &CoderContext{
		codeBasePath: filepath.Dir(root),
		repo:         repo,
		engine:       engine,
		absFileNames: map[string]struct{}{},
	}, nil
}
