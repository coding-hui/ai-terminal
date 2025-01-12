package history

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type rm struct {
	genericclioptions.IOStreams
}

func newCmdRemoveHistory(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := &rm{IOStreams: ioStreams}
	cmd := &cobra.Command{
		Use:          "rm",
		Short:        "Remove chat conversation history.",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run(args)
		},
	}

	return cmd
}

func (r *rm) Run(args []string) error {
	chatID := args[0]

	chatHistory, err := convo.GetConversationStore(*options.NewConfig())
	if err != nil {
		return err
	}

	exists, err := chatHistory.ConversationExists(context.Background(), chatID)
	if err != nil {
		return errbook.Wrap("Failed to check existence of chat conversation history: "+chatID, err)
	}
	if !exists {
		return errbook.Wrap("Chat conversation history does not exist: "+chatID, err)
	}

	err = chatHistory.ClearConversations(context.Background())
	if err != nil {
		return errbook.Wrap("Failed to remove chat conversation history: "+chatID, err)
	}

	console.Successf("Removed chat history %s", chatID)

	return nil
}
