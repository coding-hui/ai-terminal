package history

import (
	"github.com/spf13/cobra"

	"github.com/coding-hui/iam/pkg/cli/genericclioptions"

	"github.com/coding-hui/ai-terminal/internal/util/templates"
)

var historyExample = templates.Examples(`
		# Managing session history:
          ai history ls
`)

// NewCmdHistory returns a cobra command for manager history.
func NewCmdHistory(ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "history",
		Short:   "Managing chat session history.",
		Long:    "Managing chat session history.",
		Example: historyExample,
		PostRunE: func(c *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(newCmdLsHistory(ioStreams))
	cmd.AddCommand(newCmdRemoveHistory(ioStreams))

	return cmd
}
