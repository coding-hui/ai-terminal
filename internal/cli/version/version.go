// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package version print the client and server version information.
package version

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/cli/util"
	"github.com/coding-hui/ai-terminal/internal/util/templates"

	"github.com/coding-hui/common/version"
	"github.com/coding-hui/iam/pkg/cli/genericclioptions"
	restclient "github.com/coding-hui/wecoding-sdk-go/rest"
)

var versionExample = templates.Examples(`
		# Print the client and server versions for the current context
		ai version`)

// Options is a struct to support version command.
type Options struct {
	ClientOnly bool
	Short      bool
	Output     string

	client *restclient.RESTClient
	genericclioptions.IOStreams
}

// NewOptions returns initialized Options.
func NewOptions(ioStreams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: ioStreams,
	}
}

// NewCmdVersion returns a cobra command for fetching versions.
func NewCmdVersion(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print the cli version information",
		Long:    "Print the cli version information for the current context",
		Example: versionExample,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
		},
	}

	cmd.Flags().BoolVar(&o.Short, "short", o.Short, "If true, print just the version number.")
	cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "One of 'yaml' or 'json'.")

	return cmd
}

// Validate validates the provided options.
func (o *Options) Validate() error {
	if o.Output != "" && o.Output != "yaml" && o.Output != "json" {
		return errors.New(`--output must be 'yaml' or 'json'`)
	}

	return nil
}

// Run executes version command.
func (o *Options) Run() error {
	var serverErr error
	versionInfo := version.Get()

	switch o.Output {
	case "":
		if o.Short {
			fmt.Fprintf(o.Out, "Version: %s\n", versionInfo.GitVersion)
		} else {
			fmt.Fprintf(o.Out, "Version: %s\n", fmt.Sprintf("%#v", versionInfo))
		}
	case "yaml":
		marshaled, err := yaml.Marshal(&versionInfo)
		if err != nil {
			return err
		}

		fmt.Fprintln(o.Out, string(marshaled))
	case "json":
		marshaled, err := json.MarshalIndent(&versionInfo, "", "  ")
		if err != nil {
			return err
		}

		fmt.Fprintln(o.Out, string(marshaled))
	default:
		// There is a bug in the program if we hit this case.
		// However, we follow a policy of never panicking.
		return fmt.Errorf("VersionOptions were not validated: --output=%q should have been rejected", o.Output)
	}

	return serverErr
}
