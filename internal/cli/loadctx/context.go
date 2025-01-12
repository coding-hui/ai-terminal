package loadctx

import (
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

// NewCmdContext returns a cobra command for managing context
func NewCmdContext(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "context",
		Aliases: []string{"context", "ctx"},
		Short:   "Manage context for AI interactions",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(
		newCmdLoad(ioStreams, cfg),
		newCmdList(ioStreams, cfg),
		newCmdClean(ioStreams, cfg),
	)

	return cmd
}
