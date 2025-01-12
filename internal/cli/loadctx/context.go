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
	}

	cmd.AddCommand(newCmdLoad(ioStreams, cfg))

	return cmd
}
