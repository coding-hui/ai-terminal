// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ask

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/ai"
	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/coders"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
	"github.com/coding-hui/ai-terminal/internal/util/templates"
	"github.com/coding-hui/ai-terminal/internal/util/term"
)

var askExample = templates.Examples(`
		# You can ask any question, enforcing 💬 ask prompt mode:
		ai ask generate me a go application example using fiber

		# You can also pipe input that will be taken into account in your request:
		cat some_script.go | ai ask generate unit tests

		# Write new sections for a readme": 
		cat README.md | ai ask "write a new section to this README documenting a pdf sharing feature"
`)

// Options is a struct to support ask command.
type Options struct {
	genericclioptions.IOStreams
	pipe           string
	prompts        []string
	tempPromptFile string
	cfg            *options.Config
}

// NewOptions returns initialized Options.
func NewOptions(ioStreams genericclioptions.IOStreams, cfg *options.Config) *Options {
	return &Options{
		IOStreams: ioStreams,
		cfg:       cfg,
	}
}

// NewCmdASK returns a cobra command for ask any question.
func NewCmdASK(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	o := NewOptions(ioStreams, cfg)
	cmd := &cobra.Command{
		Use:     "ask",
		Short:   "CLI mode is made to be integrated in your command lines workflow.",
		Example: askExample,
		PreRunE: func(c *cobra.Command, args []string) error {
			err := o.preparePrompts(args)
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
		PostRunE: func(c *cobra.Command, args []string) error {
			if o.tempPromptFile != "" {
				err := os.Remove(o.tempPromptFile)
				if err != nil {
					return errbook.Wrap("Failed to remove temporary file: "+o.tempPromptFile, err)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&o.cfg.Interactive, "interactive", "i", o.cfg.Interactive, "Interactive dialogue model.")
	cmd.Flags().StringVarP(&o.cfg.PromptFile, "file", "f", o.cfg.PromptFile, "File containing prompt.")

	return cmd
}

// Validate validates the provided options.
func (o *Options) Validate() error {
	return nil
}

// Run executes ask command.
func (o *Options) Run() error {
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

	content := o.pipe + "\n\n" + strings.Join(o.prompts, "\n\n")
	autoCoder := coders.NewAutoCoder(
		coders.WithConfig(o.cfg),
		coders.WithEngine(engine),
		coders.WithRepo(repo),
		coders.WithCodeBasePath(filepath.Dir(root)),
		coders.WithStore(store),
		coders.WithPrompt(strings.TrimSpace(content)),
		coders.WithPromptMode(coders.ChatPromptMode),
	)

	return autoCoder.Run()
}

func (o *Options) preparePrompts(args []string) error {
	if len(args) > 0 {
		o.prompts = append(o.prompts, strings.Join(args, " "))
	}

	if o.cfg.PromptFile != "" {
		bytes, err := os.ReadFile(o.cfg.PromptFile)
		if err != nil {
			return errbook.Wrap("Couldn't reading prompt file.", err)
		}
		o.prompts = append(o.prompts, string(bytes))
	}

	o.pipe = term.ReadPipeInput()

	return nil
}
