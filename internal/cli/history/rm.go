package history

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coding-hui/iam/pkg/cli/genericclioptions"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/llm"
	"github.com/coding-hui/ai-terminal/internal/session"
	"github.com/coding-hui/ai-terminal/internal/util"
	"github.com/coding-hui/ai-terminal/internal/util/display"
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
		display.FatalErr(err, fmt.Sprintf("Failed to check existence of chat session history %s.", chatID))
	}
	if !exists {
		display.ErrorMsg("Chat session history does not exist.")
		return nil
	}

	err = chatHistory.Clear(context.Background(), chatID)
	if err != nil {
		display.FatalErr(err, fmt.Sprintf("Failed to remove chat history %s", chatID))
	}

	display.Success(fmt.Sprintf("Removed chat history %s", chatID))

	return nil
}
