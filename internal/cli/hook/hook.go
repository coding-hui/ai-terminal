package hook

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/git"
)

func NewCmdHook() *cobra.Command {
	hookCmd := &cobra.Command{
		Use:   "hook",
		Short: "install/uninstall git prepare-commit-msg hook",
	}

	hookCmd.AddCommand(
		&cobra.Command{
			Use:   "install",
			Short: "install git prepare-commit-msg hook",
			RunE: func(cmd *cobra.Command, args []string) error {
				g := git.New()

				if err := g.InstallHook(); err != nil {
					return err
				}
				color.Green("Install git hook: prepare-commit-msg successfully")
				color.Green("You can see the hook file: .git/hooks/prepare-commit-msg")

				return nil
			},
		},
		&cobra.Command{
			Use:   "uninstall",
			Short: "uninstall git prepare-commit-msg hook",
			RunE: func(cmd *cobra.Command, args []string) error {
				g := git.New()

				if err := g.UninstallHook(); err != nil {
					return err
				}
				color.Green("Remove git hook: prepare-commit-msg successfully")

				return nil
			},
		},
	)

	return hookCmd
}
