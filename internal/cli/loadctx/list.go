package loadctx

import (
	"context"
	"fmt"
	"os"

	timeago "github.com/caarlos0/timea.go"
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type list struct {
	genericclioptions.IOStreams
	cfg        *options.Config
	convoStore convo.Store
}

func newList(ioStreams genericclioptions.IOStreams, cfg *options.Config) *list {
	return &list{
		IOStreams: ioStreams,
		cfg:       cfg,
	}
}

func newCmdList(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	o := newList(ioStreams, cfg)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all loaded contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	return cmd
}

func (o *list) Run() error {
	// Initialize conversation store
	var err error
	o.convoStore, err = convo.GetConversationStore(o.cfg)
	if err != nil {
		return errbook.Wrap("failed to initialize conversation store", err)
	}

	conversation, err := convo.GetCurrentConversationID(context.Background(), o.cfg, o.convoStore)
	if err != nil {
		return errbook.Wrap("failed to get current conversation", err)
	}

	// Get contexts for current conversation
	ctxs, err := o.convoStore.ListContextsByteConvoID(context.Background(), conversation.ReadID)
	if err != nil {
		return errbook.Wrap("failed to list contexts", err)
	}

	if len(ctxs) == 0 {
		console.Render("No contexts loaded")
		return nil
	}

	// Display contexts in a table
	printList(ctxs)

	return nil
}

func printList(ctxs []convo.LoadContext) {
	for _, ctx := range ctxs {
		_, _ = fmt.Fprintf(
			os.Stdout,
			"%s\t%s\t%s\t%s\n",
			console.StdoutStyles().SHA1.Render(ctx.ConversationID[:convo.Sha1short]),
			ctx.Name,
			ctx.Type,
			console.StdoutStyles().Timeago.Render(timeago.Of(ctx.UpdatedAt)),
		)
	}
}
