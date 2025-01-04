package history

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/session"
	"github.com/coding-hui/ai-terminal/internal/ui/display"
	"github.com/coding-hui/ai-terminal/internal/util"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type rm struct {
	genericclioptions.IOStreams
}

func newCmdRemoveHistory(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := &rm{IOStreams: ioStreams}
	cmd := &cobra.Command{
		Use:          "rm",
		Short:        "Remove chat session history.",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run(args))
		},
	}

	return cmd
}

func (r *rm) Validate() error {
	return nil
}

func (r *rm) Run(args []string) error {
	chatID := args[0]

	chatHistory, err := session.GetHistoryStore(*options.NewConfig(), llm.ChatEngineMode.String())
	if err != nil {
		return err
	}

	exists, err := chatHistory.Exists(context.Background(), chatID)
	if err != nil {
		display.Fatalf("Failed to check existence of chat session history %s: %v", chatID, err)
	}
	if !exists {
		display.Error("Chat session history does not exist.")
		return nil
	}

	err = chatHistory.Clear(context.Background(), chatID)
	if err != nil {
		display.Fatalf("Failed to remove chat history %s: %v", chatID, err)
	}

	display.Successf("Removed chat history %s", chatID)

	return nil
}
