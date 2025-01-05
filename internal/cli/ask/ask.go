// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ask

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/runner"
	"github.com/coding-hui/ai-terminal/internal/system"
	"github.com/coding-hui/ai-terminal/internal/ui"
	"github.com/coding-hui/ai-terminal/internal/ui/chat"
	"github.com/coding-hui/ai-terminal/internal/ui/display"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
	"github.com/coding-hui/ai-terminal/internal/util/templates"
	"github.com/coding-hui/ai-terminal/internal/util/term"
)

const (
	promptInstructions = `👉  Write your prompt below, then save and exit to send it to AI.`
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
			return o.Run(args)
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
func (o *Options) Run(_ []string) error {
	runMode := ui.CliMode
	if o.cfg.Interactive {
		runMode = ui.ReplMode
	}
	input, err := ui.NewInput(runMode, ui.ChatPromptMode, o.pipe, o.prompts)
	if err != nil {
		return err
	}

	chatModel, err := chat.NewChat(input, display.StderrRenderer(), o.cfg)
	if err != nil {
		return errbook.Wrap("Couldn't create Bubble Tea chat model.", err)
	}

	if _, err := tea.NewProgram(chatModel).Run(); err != nil {
		return errbook.Wrap("Couldn't start Bubble Tea program.", err)
	}

	if chatModel.Error != nil {
		return *chatModel.Error
	}

	if term.IsOutputTTY() {
		switch {
		case chatModel.GlamOutput != "":
			fmt.Print(chatModel.GlamOutput)
		case chatModel.Output != "":
			fmt.Print(chatModel.Output)
		}
	}

	return nil
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
	if len(o.prompts) == 0 && len(o.pipe) == 0 && !o.cfg.Interactive {
		prompt, err := o.getEditorPrompt()
		if err != nil {
			return err
		}
		o.prompts = append(o.prompts, prompt)
	}

	return nil
}

func (o *Options) getEditorPrompt() (string, error) {
	tempFile, err := os.CreateTemp(os.TempDir(), "ai_prompt_*.txt")
	if err != nil {
		return "", errbook.Wrap("Failed to create temporary file.", err)
	}

	filename := tempFile.Name()
	o.tempPromptFile = filename
	err = os.WriteFile(filename, []byte(promptInstructions), 0644)
	if err != nil {
		return "", errbook.Wrap("Failed to write instructions to temporary file.", err)
	}

	editor := system.Analyse().GetEditor()
	editorCmd := runner.PrepareEditSettingsCommand(editor, filename)
	editorCmd.Stdin, editorCmd.Stdout, editorCmd.Stderr = o.In, o.Out, o.ErrOut
	err = editorCmd.Start()
	if err != nil {
		return "", errbook.Wrap("Error opening editor.", err)
	}
	_ = editorCmd.Wait()

	bytes, err := os.ReadFile(filename)
	if err != nil {
		return "", errbook.Wrap("Error reading temporary file.", err)
	}

	prompt := string(bytes)

	prompt = strings.TrimPrefix(prompt, strings.TrimSpace(promptInstructions))
	prompt = strings.TrimSpace(prompt)

	return prompt, nil
}
