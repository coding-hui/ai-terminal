package convo

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type show struct {
	genericclioptions.IOStreams
	DeleteOlderThan time.Duration
}

func newCmdShowConversation(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	o := &show{IOStreams: ioStreams}
	cmd := &cobra.Command{
		Use:          "show",
		Short:        "Show chat conversation.",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run(args, cfg)
		},
	}

	return cmd
}

func (s *show) Run(args []string, cfg *options.Config) error {
	store, err := convo.GetConversationStore(cfg)
	if err != nil {
		return err
	}

	chatID := args[0]
	conversation, err := store.GetConversation(context.Background(), chatID)
	if err != nil {
		return errbook.Wrap("Couldn't find conversation to delete.", err)
	}

	err = store.DeleteConversation(context.Background(), conversation.ID)
	if err != nil {
		return errbook.Wrap("Couldn't delete conversation.", err)
	}

	err = store.InvalidateMessages(context.Background(), conversation.ID)
	if err != nil {
		return errbook.Wrap("Couldn't invalidate conversation.", err)
	}

	if !cfg.Quiet {
		fmt.Fprintln(os.Stderr, "Conversation deleted:", conversation.ID[:convo.Sha1minLen])
	}

	return nil
}
