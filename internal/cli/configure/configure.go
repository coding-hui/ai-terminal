package configure

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/runner"
	"github.com/coding-hui/ai-terminal/internal/system"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

// NewCmdConfigure implements the configure command.
func NewCmdConfigure(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configure",
		Aliases: []string{"conf", "cfg", "config", "settings"},
		Short:   "Configure AI settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			editor := system.Analyse().GetEditor()
			editorCmd := runner.PrepareEditSettingsCommand(editor, cfg.SettingsPath)
			editorCmd.Stdin, editorCmd.Stdout, editorCmd.Stderr = ioStreams.In, ioStreams.Out, ioStreams.ErrOut

			err := editorCmd.Start()
			if err != nil {
				return errbook.Wrap("Could not edit your settings file.", err)
			}

			err = editorCmd.Wait()
			if err != nil {
				return errbook.Wrap("Could not wait for your settings file to be saved.", err)
			}

			if !cfg.Quiet {
				fmt.Fprintln(os.Stderr, "Wrote config file to:", cfg.SettingsPath)
			}

			return nil
		},
	}

	cmd.AddCommand(newCmdResetConfig(ioStreams, cfg))
	cmd.AddCommand(newCmdEchoConfig(ioStreams, cfg))

	return cmd
}
