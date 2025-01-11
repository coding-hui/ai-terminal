package coder

import (
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/coders"
)

type Options struct {
	cfg *options.Config
}

func NewCmdCoder(cfg *options.Config) *cobra.Command {
	ops := &Options{cfg: cfg}
	cmd := &cobra.Command{
		Use:   "coder",
		Short: "Automatically generate code based on prompts.",
		RunE:  ops.run,
	}

	return cmd
}

func (o *Options) run(cmd *cobra.Command, args []string) error {
	autoCoder, err := coders.NewAutoCoder(o.cfg)
	if err != nil {
		return err
	}

	return autoCoder.Run()
}
