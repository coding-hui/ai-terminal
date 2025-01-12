// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package loadctx

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

// clean is a struct to support clean command
type clean struct {
	genericclioptions.IOStreams
	cfg        *options.Config
	convoStore convo.Store
}

// newClean returns initialized clean
func newClean(ioStreams genericclioptions.IOStreams, cfg *options.Config) *clean {
	return &clean{
		IOStreams: ioStreams,
		cfg:       cfg,
	}
}

func newCmdClean(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	o := newClean(ioStreams, cfg)

	cmd := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"drop"},
		Short:   "Delete all loaded contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	return cmd
}

// Run executes the clean command
func (o *clean) Run() error {
	// Initialize conversation store
	var err error
	o.convoStore, err = convo.GetConversationStore(o.cfg)
	if err != nil {
		return err
	}

	// Clean all contexts for current conversation
	count, err := o.convoStore.CleanContexts(context.Background(), o.cfg.ConversationID)
	if err != nil {
		return err
	}

	if count > 0 {
		console.Render("Successfully deleted %d loaded contexts", count)
	} else {
		console.Render("No contexts to delete")
	}

	return nil
}
