// Copyright (c) 2023-2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package cli create a root cobra command and add subcommands to it.
package cli

import (
	"flag"
	"io"
	"os"
	"slices"
	"time"

	cliflag "github.com/coding-hui/common/cli/flag"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/coding-hui/ai-terminal/internal/cli/ask"
	"github.com/coding-hui/ai-terminal/internal/cli/coder"
	"github.com/coding-hui/ai-terminal/internal/cli/commit"
	"github.com/coding-hui/ai-terminal/internal/cli/completion"
	"github.com/coding-hui/ai-terminal/internal/cli/configure"
	"github.com/coding-hui/ai-terminal/internal/cli/convo"
	"github.com/coding-hui/ai-terminal/internal/cli/hook"
	"github.com/coding-hui/ai-terminal/internal/cli/loadctx"
	"github.com/coding-hui/ai-terminal/internal/cli/manpage"
	"github.com/coding-hui/ai-terminal/internal/cli/review"
	"github.com/coding-hui/ai-terminal/internal/cli/version"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/util/debug"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
	"github.com/coding-hui/ai-terminal/internal/util/templates"

	_ "github.com/coding-hui/ai-terminal/internal/convo/sqlite3"
)

var logFlushFreq = pflag.Duration(options.FlagLogFlushFrequency, 5*time.Second, "Maximum number of seconds between log flushes")

// NewDefaultAICommand creates the `ai` command with default arguments.
func NewDefaultAICommand() *cobra.Command {
	return NewAICommand(os.Stdin, os.Stdout, os.Stderr)
}

// NewAICommand returns new initialized instance of 'ai' root command.
func NewAICommand(in io.Reader, out, errOut io.Writer) *cobra.Command {
	cfg, err := options.EnsureConfig()
	if err != nil {
		errbook.HandleError(errbook.Wrap("Could not load your configuration file.", err))
		// if user is editing the settings, only print out the error, but do
		// not exit.
		if !slices.Contains(os.Args, "--settings") {
			os.Exit(1)
		}
	}

	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "ai",
		Short: "AI driven development in your terminal.",
		Long: templates.LongDesc(`
      AI driven development in your terminal.

      Find more information at:
            https://github.com/coding-hui/ai-terminal`),
		SilenceUsage:  true,
		SilenceErrors: true,
		Run:           runHelp,
		// Hook before and after Run initialize and write profiles to disk,
		// respectively.
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return initProfiling()
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			return postRunHook(&cfg)
		},
	}

	flags := cmds.PersistentFlags()
	flags.SetNormalizeFunc(cliflag.WarnWordSepNormalizeFunc) // Warn for "_" flags

	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	addProfilingFlags(flags)

	options.AddBasicFlags(flags, &cfg)

	cmds.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// From this point and forward we get warnings on flags that contain "_" separators
	cmds.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: errOut}

	groups := templates.CommandGroups{
		templates.CommandGroup{
			Message: "AI Commands:",
			Commands: []*cobra.Command{
				coder.NewCmdCoder(&cfg),
				ask.NewCmdASK(ioStreams, &cfg),
				convo.NewCmdConversation(ioStreams, &cfg),
				commit.NewCmdCommit(ioStreams, &cfg),
				review.NewCmdCommit(ioStreams),
				loadctx.NewCmdContext(ioStreams, &cfg),
			},
		},
		templates.CommandGroup{
			Message: "Settings Commands:",
			Commands: []*cobra.Command{
				configure.NewCmdConfigure(ioStreams, &cfg),
				completion.NewCmdCompletion(),
				manpage.NewCmdManPage(cmds),
				hook.NewCmdHook(),
			},
		},
	}
	groups.Add(cmds)

	filters := []string{"options"}
	templates.ActsAsRootCommand(cmds, filters, groups...)

	cmds.AddCommand(version.NewCmdVersion(ioStreams))
	cmds.AddCommand(options.NewCmdOptions(ioStreams.Out))

	defer func() {
		debug.Teardown()
	}()

	return cmds
}

func runHelp(cmd *cobra.Command, _ []string) {
	_ = cmd.Help()
}

func postRunHook(cfg *options.Config) error {
	if err := flushProfiling(); err != nil {
		return err
	}
	return nil
}
