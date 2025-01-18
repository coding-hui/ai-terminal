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
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/flag"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type rm struct {
	genericclioptions.IOStreams
	cfg             *options.Config
	DeleteOlderThan time.Duration
}

func newCmdRemoveConversation(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	o := &rm{IOStreams: ioStreams, cfg: cfg}
	cmd := &cobra.Command{
		Use:          "rm",
		Short:        "Remove chat conversation.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run(args)
		},
	}

	cmd.Flags().Var(flag.NewDurationFlag(o.DeleteOlderThan, &o.DeleteOlderThan), "delete-older-than", console.StdoutStyles().FlagDesc.Render(options.Help["delete-older-than"]))

	return cmd
}

func (r *rm) Run(args []string) error {
	if r.DeleteOlderThan <= 0 && len(args) <= 0 {
		return errbook.New("Please provide at least one conversation ID or --delete-older-than flag")
	}

	store, err := convo.GetConversationStore(r.cfg)
	if err != nil {
		return err
	}

	if r.DeleteOlderThan > 0 {
		return r.deleteConversationOlderThan(store)
	}

	return r.deleteConversation(store, args[0])
}

func (r *rm) deleteConversation(store convo.Store, conversationID string) error {
	conversation, err := store.GetConversation(context.Background(), conversationID)
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

	if !r.cfg.Quiet {
		fmt.Fprintln(os.Stderr, "Conversation deleted:", conversation.ID[:convo.Sha1minLen])
	}

	return nil
}

func (r *rm) deleteConversationOlderThan(store convo.Store) error {
	conversations, err := store.ListConversationsOlderThan(context.Background(), r.DeleteOlderThan)
	if err != nil {
		return errbook.Wrap("Couldn't list conversations.", err)
	}

	if len(conversations) == 0 {
		if !r.cfg.Quiet {
			fmt.Fprintln(os.Stderr, "No conversations found.")
			return nil
		}
		return nil
	}

	if !r.cfg.Quiet {
		printList(conversations)
		confirm := console.WaitForUserConfirm(console.No, "Delete conversations older than %s?", r.DeleteOlderThan)
		if !confirm {
			return errbook.NewUserErrorf("Aborted by user.")
		}
	}

	for _, c := range conversations {
		if err := r.deleteConversation(store, c.ID); err != nil {
			return err
		}
	}

	return nil
}
