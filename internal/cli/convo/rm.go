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
	DeleteAll       bool
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

	cmd.Flags().Var(flag.NewDurationFlag(o.DeleteOlderThan, &o.DeleteOlderThan), "delete-older-than", console.StdoutStyles().FlagDesc.Render(options.Help["rm-convo-older-than"]))
	cmd.Flags().BoolVar(&o.DeleteAll, "all", false, console.StdoutStyles().FlagDesc.Render(options.Help["rm-all-convo"]))

	return cmd
}

func (r *rm) Run(args []string) error {
	if r.DeleteOlderThan <= 0 && len(args) <= 0 && !r.DeleteAll {
		return errbook.New("Please provide at least one conversation ID or --delete-older-than flag")
	}

	store, err := convo.GetConversationStore(r.cfg)
	if err != nil {
		return err
	}

	if r.DeleteOlderThan > 0 {
		return r.deleteConversationOlderThan(store, false)
	}

	if r.DeleteAll {
		return r.deleteConversationOlderThan(store, true)
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

	_, err = store.CleanContexts(context.Background(), conversation.ID)
	if err != nil {
		return errbook.Wrap("Couldn't clean conversation load contexts.", err)
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

func (r *rm) deleteConversationOlderThan(store convo.Store, deleteAll bool) error {
	var err error
	var conversations []convo.Conversation

	if deleteAll {
		conversations, err = store.ListConversations(context.Background())
		if err != nil {
			return errbook.Wrap("Couldn't list conversations.", err)
		}
	} else {
		conversations, err = store.ListConversationsOlderThan(context.Background(), r.DeleteOlderThan)
		if err != nil {
			return errbook.Wrap("Couldn't list conversations.", err)
		}
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
		confirmTitle := "Delete all conversations?"
		if !deleteAll {
			confirmTitle = fmt.Sprintf("Delete conversations older than %s?", r.DeleteOlderThan)
		}
		confirm := console.WaitForUserConfirm(console.Yes, confirmTitle)
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
