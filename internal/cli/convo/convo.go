package convo

import (
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

// NewCmdConversation returns a cobra command for manager convo.
func NewCmdConversation(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convo",
		Short: "Managing chat conversations.",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(newCmdLsConversation(ioStreams, cfg))
	cmd.AddCommand(newCmdRemoveConversation(ioStreams, cfg))
	cmd.AddCommand(newCmdShowConversation(ioStreams, cfg))

	return cmd
}
