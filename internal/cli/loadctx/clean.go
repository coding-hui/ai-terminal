// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package loadctx

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

// clean represents a command to clean up loaded contexts in the AI terminal.
// It provides functionality to remove all context data for the current conversation.
type clean struct {
	genericclioptions.IOStreams
	cfg        *options.Config // Configuration for the AI terminal
	convoStore convo.Store     // Storage interface for conversation data
}

// newClean creates and initializes a new clean command instance.
// Parameters:
//   - ioStreams: Streams for input/output operations
//   - cfg: Configuration settings for the AI terminal
//
// Returns:
//   - *clean: A new clean command instance
func newClean(ioStreams genericclioptions.IOStreams, cfg *options.Config) *clean {
	return &clean{
		IOStreams: ioStreams,
		cfg:       cfg,
	}
}

// newCmdClean creates a new cobra command for the clean operation.
// It sets up the command structure and binds it to the clean functionality.
func newCmdClean(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	o := newClean(ioStreams, cfg)

	cmd := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"drop"},
		Short:   "Delete all loaded contexts",
		Long:    "Clean command removes all loaded contexts for the current conversation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	return cmd
}

// Run executes the clean command logic.
// It performs the following steps:
// 1. Initializes the conversation store
// 2. Retrieves the current conversation ID
// 3. Cleans all contexts associated with the current conversation
// 4. Displays the operation result to the user
//
// Returns:
//   - error: Any error that occurred during execution
func (o *clean) Run() error {
	ctx := context.Background()

	// Initialize conversation store
	var err error
	o.convoStore, err = convo.GetConversationStore(o.cfg)
	if err != nil {
		return errbook.Wrap("failed to initialize conversation store", err)
	}

	// Get current conversation details
	conversation, err := convo.GetCurrentConversationID(ctx, o.cfg, o.convoStore)
	if err != nil {
		return errbook.Wrap("failed to get current conversation", err)
	}

	// Clean all contexts for current conversation
	count, err := o.convoStore.CleanContexts(ctx, conversation.ReadID)
	if err != nil {
		return errbook.Wrap("failed to clean contexts", err)
	}

	// Display operation result
	if count > 0 {
		console.Render("Successfully deleted %d loaded contexts", count)
	} else {
		console.Render("No contexts to delete")
	}

	return nil
}
