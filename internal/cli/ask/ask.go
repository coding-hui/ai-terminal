// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ask

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/coding-hui/iam/pkg/cli/genericclioptions"

	"github.com/coding-hui/ai-terminal/internal/cli/ui"
	"github.com/coding-hui/ai-terminal/internal/cli/util"
	"github.com/coding-hui/ai-terminal/internal/util/templates"
)

var askExample = templates.Examples(`
		# You can ask any question, enforcing 💬 ask prompt mode:
		  	ai ask generate me a go application example using fiber
		  You can also pipe input that will be taken into account in your request:
			cat some_script.go | ai ask generate unit tests
`)

// Options is a struct to support ask command.
type Options struct {
	tty, stdin bool
	genericclioptions.IOStreams
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
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run(args))
		},
	}

	cmd.Flags().BoolVarP(&o.stdin, "stdin", "i", o.stdin, "Pass stdin to the ai terminal.")
	cmd.Flags().BoolVarP(&o.tty, "tty", "t", o.tty, "Stdin is a TTY.")

	return cmd
}

// Validate validates the provided options.
func (o *Options) Validate() error {
	return nil
}

// Run executes version command.
func (o *Options) Run(args []string) error {
	runMode := ui.CliMode
	if o.tty {
		runMode = ui.ReplMode
	}
	input, err := ui.NewInput(runMode, ui.ChatPromptMode, args)
	if err != nil {
		return err
	}

	klog.V(2).InfoS("start ask cli mode.", "args", args, "runMode", runMode, "pipe", input.GetPipe())

	if _, err := tea.NewProgram(ui.NewUi(input)).Run(); err != nil {
		return err
	}

	return nil
}
