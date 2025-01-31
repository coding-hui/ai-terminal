package coder

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/ai"
	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/coders"
)

type Options struct {
	cfg    *options.Config
	prompt string
}

func NewCmdCoder(cfg *options.Config) *cobra.Command {
	ops := &Options{cfg: cfg}
	cmd := &cobra.Command{
		Use:   "coder",
		Short: "Automatically generate code based on prompts.",
		RunE:  ops.run,
	}

	cmd.Flags().StringVarP(&ops.prompt, "prompt", "p", "", "Prompt to generate code.")

	return cmd
}

func (o *Options) run(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		o.prompt = strings.Join(args, " ") + "\n" + o.prompt
	}

	repo := git.New()
	root, err := repo.GitDir()
	if err != nil {
		return errbook.Wrap("Could not get git root", err)
	}

	engine, err := ai.New(ai.WithConfig(o.cfg))
	if err != nil {
		return errbook.Wrap("Could not initialized ai engine", err)
	}

	store, err := convo.GetConversationStore(o.cfg)
	if err != nil {
		return errbook.Wrap("Could not initialize conversation store", err)
	}

	autoCoder := coders.NewAutoCoder(
		coders.WithConfig(o.cfg),
		coders.WithEngine(engine),
		coders.WithRepo(repo),
		coders.WithCodeBasePath(filepath.Dir(root)),
		coders.WithStore(store),
		coders.WithPrompt(o.prompt),
	)

	return autoCoder.Run()
}
