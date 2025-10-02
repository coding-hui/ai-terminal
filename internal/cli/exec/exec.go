package exec

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/ai"
	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/git"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/coders"
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
			input := strings.TrimSpace(strings.Join(args, " "))
			if o.auto {
				input += " --yes"
			}

			// Delegate to coder's /exec subcommand via AutoCoder
			repo := git.New()
			root, err := repo.GitDir()
			if err != nil {
				return err
			}

			engine, err := ai.New(ai.WithConfig(o.cfg))
			if err != nil {
				return err
			}

			store, err := convo.GetConversationStore(o.cfg)
			if err != nil {
				return err
			}

			autoCoder := coders.NewAutoCoder(
				coders.WithConfig(o.cfg),
				coders.WithEngine(engine),
				coders.WithRepo(repo),
				coders.WithCodeBasePath(filepath.Dir(root)),
				coders.WithStore(store),
				coders.WithPrompt(strings.TrimSpace(input)),
				coders.WithPromptMode(coders.ExecPromptMode),
			)

			return autoCoder.Run()
		},
	}

	cmd.Flags().BoolVarP(&o.auto, "yes", "y", false, "Run without confirmation (auto-execute the inferred command)")
	cmd.Flags().BoolVarP(&o.cfg.Interactive, "interactive", "i", false, "Interactive dialogue mode.")

	return cmd
}
