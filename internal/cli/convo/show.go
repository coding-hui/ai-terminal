package convo

import (
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/cli/ask"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type show struct {
	genericclioptions.IOStreams
	last bool
}

func newCmdShowConversation(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	o := &show{IOStreams: ioStreams}
	cmd := &cobra.Command{
		Use:          "show",
		Short:        "Show chat conversation.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run(args, cfg)
		},
	}

	cmd.Flags().BoolVarP(&o.last, "last", "l", false, "show last chat conversation.")

	return cmd
}

func (s *show) Run(args []string, cfg *options.Config) error {
	cfg.ShowLast = s.last
	if len(args) > 0 {
		cfg.Show = args[0]
	}
	if err := ask.NewOptions(s.IOStreams, cfg).Run(); err != nil {
		return err
	}

	return nil
}
