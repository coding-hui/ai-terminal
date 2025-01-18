package convo

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	timeago "github.com/caarlos0/timea.go"
	"github.com/charmbracelet/huh"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
	"github.com/coding-hui/ai-terminal/internal/util/term"
)

type ls struct{}

func newCmdLsConversation(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	o := &ls{}
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "Show chat conversations.",
		Example: `# Managing conversations:
          ai convo ls`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return o.Run(ioStreams, cfg)
		},
	}

	return cmd
}

// Run executes convo command.
func (o *ls) Run(ioStreams genericclioptions.IOStreams, cfg *options.Config) error {
	store, err := convo.GetConversationStore(cfg)
	if err != nil {
		return err
	}

	conversations, err := store.ListConversations(context.Background())
	if err != nil {
		return err
	}

	if len(conversations) == 0 {
		fmt.Fprintln(ioStreams.ErrOut, "No conversations found.")
		return nil
	}

	if term.IsInputTTY() && term.IsOutputTTY() {
		selectFromList(conversations)
		return nil
	}

	printList(conversations)

	return nil
}

func makeOptions(conversations []convo.Conversation) []huh.Option[string] {
	opts := make([]huh.Option[string], 0, len(conversations))
	for _, c := range conversations {
		timea := console.StdoutStyles().Timeago.Render(timeago.Of(c.UpdatedAt))
		left := console.StdoutStyles().SHA1.Render(c.ID[:convo.Sha1short])
		right := console.StdoutStyles().ConversationList.Render(c.Title, timea)
		if c.Model != nil {
			right += console.StdoutStyles().Comment.Render(*c.Model)
		}
		opts = append(opts, huh.NewOption(left+" "+right, c.ID))
	}
	return opts
}

func selectFromList(conversations []convo.Conversation) {
	var selected string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Conversations").
				Value(&selected).
				Options(makeOptions(conversations)...),
		),
	).Run(); err != nil {
		if !errors.Is(err, huh.ErrUserAborted) {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		return
	}

	_ = clipboard.WriteAll(selected)
	termenv.Copy(selected)
	console.PrintConfirmation("COPIED", selected)
	// suggest actions to use this conversation ID
	fmt.Println(console.StdoutStyles().Comment.Render(
		"You can use this conversation ID with the following commands:",
	))

	type suggestion struct {
		cmd   string
		usage string
	}

	suggestions := []suggestion{
		{
			cmd:   "show-convo",
			usage: "ai convo show",
		},
		{
			cmd:   "continue",
			usage: "ai ask --continue",
		},
		{
			cmd:   "rm-convo",
			usage: "ai convo rm",
		},
	}
	for _, flag := range suggestions {
		fmt.Printf(
			"  %-44s %s\n",
			console.StdoutStyles().Flag.Render(flag.usage),
			console.StdoutStyles().FlagDesc.Render(options.Help[flag.cmd]),
		)
	}
}

func printList(conversations []convo.Conversation) {
	for _, conversation := range conversations {
		_, _ = fmt.Fprintf(
			os.Stdout,
			"%s\t%s\t%s\n",
			console.StdoutStyles().SHA1.Render(conversation.ID[:convo.Sha1short]),
			conversation.Title,
			console.StdoutStyles().Timeago.Render(timeago.Of(conversation.UpdatedAt)),
		)
	}
}
