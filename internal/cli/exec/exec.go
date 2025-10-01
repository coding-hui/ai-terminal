package exec

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

	"github.com/coding-hui/ai-terminal/internal/ai"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/runner"
	"github.com/coding-hui/ai-terminal/internal/ui/chat"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type Options struct {
	genericclioptions.IOStreams
	cfg  *options.Config
	auto bool
}

func NewOptions(ioStreams genericclioptions.IOStreams, cfg *options.Config) *Options {
	return &Options{IOStreams: ioStreams, cfg: cfg}
}

// NewCmdExec returns a cobra command that executes a shell command.
func NewCmdExec(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	o := NewOptions(ioStreams, cfg)

	cmd := &cobra.Command{
		Use:   "exec [instruction]",
		Short: "Use AI to interpret your instruction and execute the shell command",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no command provided to execute")
			}

			input := strings.TrimSpace(strings.Join(args, " "))

			// Always infer the actual shell command using AI
			inferred, err := o.inferCommandWithAI(input)
			if err != nil {
				return err
			}
			if inferred == "" {
				return errors.New("AI did not return a runnable command")
			}

			if !o.auto {
				if !console.WaitForUserConfirm(console.Yes, "\nRun inferred command? %s", console.StderrStyles().InlineCode.Render(inferred)) {
					return nil
				}
			}

			input = inferred

			shellCmd := runner.PrepareInteractiveCommand(input)
			shellCmd.Stdin = o.In
			shellCmd.Stdout = o.Out
			shellCmd.Stderr = o.ErrOut
			return shellCmd.Run()
		},
	}

	cmd.Flags().BoolVarP(&o.auto, "yes", "y", false, "Run without confirmation (auto-execute the inferred command)")

	return cmd
}

// inferCommandWithAI sends the user's natural language to the AI engine and
// expects a single-line shell command in response.
func (o *Options) inferCommandWithAI(natural string) (string, error) {
	engine, err := ai.New(
		ai.WithConfig(o.cfg),
		ai.WithMode(ai.ExecEngineMode),
	)
	if err != nil {
		return "", err
	}

	system := llms.SystemChatMessage{Content: strings.Join([]string{
		"You are a helpful terminal assistant.",
		"Convert the user's request into a single-line POSIX shell command.",
		"Return ONLY the command without explanations, quotes, code fences, or newlines.",
		"If the request is unclear or unsafe (like destructive operations), respond with [noexec] and a short reason.",
	}, " ")}

	human := llms.HumanChatMessage{Content: natural}

	// Use chat UI to render a cool terminal experience while inferring the command
	ch := chat.NewChat(o.cfg,
		chat.WithEngine(engine),
		chat.WithMessages([]llms.ChatMessage{system, human}),
	)
	if err := ch.Run(); err != nil {
		return "", err
	}

	// Extract AI output from chat
	content := strings.TrimSpace(ch.GetOutput())
	content = strings.Trim(content, "`")
	if content == "" || strings.HasPrefix(strings.ToLower(content), "[noexec]") {
		return "", nil
	}
	if idx := strings.IndexByte(content, '\n'); idx >= 0 {
		content = strings.TrimSpace(content[:idx])
	}

	return content, nil
}
