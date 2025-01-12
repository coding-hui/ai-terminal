// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package load

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/convo"
	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
	"github.com/coding-hui/ai-terminal/internal/util/rest"
	"github.com/coding-hui/ai-terminal/internal/util/term"
)

// Options is a struct to support load command
type Options struct {
	genericclioptions.IOStreams
	cfg        *options.Config
	convoStore convo.Store
}

// NewOptions returns initialized Options
func NewOptions(ioStreams genericclioptions.IOStreams, cfg *options.Config) *Options {
	return &Options{
		IOStreams: ioStreams,
		cfg:       cfg,
	}
}

// NewCmdLoad returns a cobra command for loading files
func NewCmdLoad(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	o := NewOptions(ioStreams, cfg)

	cmd := &cobra.Command{
		Use:   "load <file|url>",
		Short: "Preload files or remote documents for later use",
		Example: `  # Load a local file
  ai load ./example.txt

  # Load a remote document
  ai load https://example.com/doc.txt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errbook.New("Please provide at least one file or URL to load")
			}
			return o.Run(args)
		},
	}

	return cmd
}

// Run executes the load command
func (o *Options) Run(args []string) (err error) {
	// Initialize conversation store
	o.convoStore, err = convo.GetConversationStore(o.cfg)
	if err != nil {
		return errbook.Wrap("Failed to initialize conversation store", err)
	}

	for _, path := range args {
		if err = o.loadPath(path); err != nil {
			return err
		}
	}

	return nil
}

func (o *Options) loadPath(path string) error {
	// Handle remote URLs
	if rest.IsValidURL(path) {
		console.Render("Loading remote content [%s]", path)
		content, err := rest.FetchURLContent(path)
		if err != nil {
			return errbook.Wrap("Failed to load remote content", err)
		}
		return o.saveContent(path, content, convo.ContentTypeURL)
	}

	// Handle local files
	console.Render("Loading local file [%s]", path)
	if err := o.saveContent(path, "", convo.ContentTypeFile); err != nil {
		return err
	}

	return nil
}

func (o *Options) saveContent(sourcePath, content string, contentType convo.ContentType) error {
	// Create cache directory if it doesn't exist
	cacheDir := filepath.Join(o.cfg.DataStore.CachePath, "loaded")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return errbook.Wrap("Failed to create cache directory", err)
	}

	// Generate safe filename
	filename := term.SanitizeFilename(sourcePath)
	cachePath := filepath.Join(cacheDir, filename)

	if contentType != convo.ContentTypeFile {
		// Save content to cache
		if err := os.WriteFile(cachePath, []byte(content), 0644); err != nil {
			return errbook.Wrap("Failed to save content", err)
		}
	}

	err := o.convoStore.SaveContext(context.Background(), &convo.LoadContext{
		ConversationID: o.cfg.ConversationID,
		Name:           filename,
		Type:           contentType,
		URL:            sourcePath,
		FilePath:       cachePath,
		Content:        content,
	})
	if err != nil {
		return errbook.Wrap("Failed to save load content", err)
	}

	console.Render("Successfully loaded and cached [%s]", filename)

	return nil
}
