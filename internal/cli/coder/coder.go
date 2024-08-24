package coder

import (
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/coders"
)

type Options struct {
}

func NewCmdCoder() *cobra.Command {
	ops := &Options{}
	cmd := &cobra.Command{
		Use:   "coder",
		Short: "Automatically generate code based on prompts.",
		RunE:  ops.run,
	}

	return cmd
}

func (o *Options) run(cmd *cobra.Command, args []string) error {
	return coders.StartAutCoder()
}
