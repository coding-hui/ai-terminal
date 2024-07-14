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
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"

	"github.com/coding-hui/iam/pkg/cli/genericclioptions"

	"github.com/coding-hui/ai-terminal/internal/cli/llm"
	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/cli/ui"
	"github.com/coding-hui/ai-terminal/internal/cli/util"
	"github.com/coding-hui/ai-terminal/internal/display"
	"github.com/coding-hui/ai-terminal/internal/run"
	"github.com/coding-hui/ai-terminal/internal/system"
	"github.com/coding-hui/ai-terminal/internal/util/templates"
)

const (
	promptInstructions = `👉  Write your prompt below, then save and exit to send it to AI.`
)

var askExample = templates.Examples(`
		# You can ask any question, enforcing 💬 ask prompt mode:
		  	ai ask generate me a go application example using fiber
		  You can also pipe input that will be taken into account in your request:
			cat some_script.go | ai ask generate unit tests
`)

// Options is a struct to support ask command.
type Options struct {
	interactive, printRaw bool
	prompts               []string
	promptFile            string
	pipe                  string
	genericclioptions.IOStreams

	tempPromptFile string
}

// NewOptions returns initialized Options.
func NewOptions(ioStreams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: ioStreams,
	}
}

// NewCmdASK returns a cobra command for ask any question.
func NewCmdASK(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:     "ask",
		Short:   "CLI mode is made to be integrated in your command lines workflow.",
		Long:    "CLI mode is made to be integrated in your command lines workflow.",
		Example: askExample,
		PreRunE: func(c *cobra.Command, args []string) error {
			err := o.preparePrompts(args)
			if err != nil {
				return err
			}
			if len(o.prompts) == 0 && o.pipe == "" {
				o.interactive = true
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run(args))
		},
		PostRunE: func(c *cobra.Command, args []string) error {
			if o.tempPromptFile != "" {
				err := os.Remove(o.tempPromptFile)
				if err != nil {
					display.FatalErr(err, "Error removing temporary file")
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&o.interactive, "interactive", "i", o.interactive, "Interactive dialogue model.")
	cmd.Flags().StringVarP(&o.promptFile, "file", "f", o.promptFile, "File containing prompt.")
	cmd.Flags().BoolVar(&o.printRaw, "raw", o.printRaw, "Return model raw return, no Stream UI.")

	options.NewLLMFlags(false).AddFlags(cmd.Flags())

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = viper.BindPFlag(flag.Name, flag)
	})

	return cmd
}

// Validate validates the provided options.
func (o *Options) Validate() error {
	return nil
}

// Run executes version command.
func (o *Options) Run(args []string) error {
	runMode := ui.CliMode
	if o.interactive {
		runMode = ui.ReplMode
	}

	input, err := ui.NewInput(runMode, ui.ChatPromptMode, o.pipe, o.prompts)
	if err != nil {
		return err
	}

	klog.V(2).InfoS("start ask cli mode.", "args", args, "runMode", runMode, "pipe", input.GetPipe())

	if o.printRaw {
		cfg, err := options.NewConfig()
		if err != nil {
			display.FatalErr(err, "Failed to load ask cmd config")
		}
		engine, err := llm.NewDefaultEngine(llm.ChatEngineMode, cfg)
		if err != nil {
			display.FatalErr(err, "Failed to initialize engine")
		}
		out, err := engine.ExecCompletion(strings.Join(o.prompts, "\n") + "\n" + o.pipe)
		if err != nil {
			display.FatalErr(err, "Error executing completion")
		}
		fmt.Println(out.Explanation)
		return nil
	}

	if _, err := tea.NewProgram(ui.NewUi(input)).Run(); err != nil {
		return err
	}

	return nil
}

func (o *Options) preparePrompts(args []string) error {
	if len(args) > 0 {
		o.prompts = append(o.prompts, strings.Join(args, " "))
	}

	if o.promptFile != "" {
		bytes, err := os.ReadFile(o.promptFile)
		if err != nil {
			display.FatalErr(err, "Error reading prompt file")
		}
		o.prompts = append(o.prompts, string(bytes))
	}

	o.pipe = util.ReadPipeInput()
	if len(o.prompts) == 0 && o.pipe == "" && !o.interactive {
		o.prompts = append(o.prompts, o.getEditorPrompt())
	}

	return nil
}

func (o *Options) getEditorPrompt() string {
	tempFile, err := os.CreateTemp(os.TempDir(), "ai_prompt_*.txt")
	if err != nil {
		display.FatalErr(err, "Failed to create temporary file")
	}

	filename := tempFile.Name()
	o.tempPromptFile = filename
	err = os.WriteFile(filename, []byte(promptInstructions), 0644)
	if err != nil {
		display.FatalErr(err, "Failed to write instructions to temporary file")
	}

	editor := system.Analyse().GetEditor()
	editorCmd := run.PrepareEditSettingsCommand(fmt.Sprintf("%s %s", editor, filename))
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	err = editorCmd.Start()
	if err != nil {
		display.FatalErr(err, "Error opening editor")
	}
	_ = editorCmd.Wait()

	bytes, err := os.ReadFile(filename)
	if err != nil {
		display.FatalErr(err, "Error reading temporary file")
	}

	prompt := string(bytes)

	prompt = strings.TrimPrefix(prompt, strings.TrimSpace(promptInstructions))
	prompt = strings.TrimSpace(prompt)

	return prompt
}
