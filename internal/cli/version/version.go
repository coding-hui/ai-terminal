// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package version print the client and server version information.
package version

import (
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/coding-hui/common/version"
	"github.com/coding-hui/iam/pkg/cli/genericclioptions"

	"github.com/coding-hui/ai-terminal/internal/util"
	"github.com/coding-hui/ai-terminal/internal/util/templates"
)

var versionExample = templates.Examples(`
		# Print the client and server versions for the current context
		ai version

		# Print the version in JSON format
		ai version --output json

		# Print the version in YAML format
		ai version --output yaml

		# Print the version using a custom Go template
		ai version --template '{{.GitVersion}}'`)

// Options is a struct to support version command.
type Options struct {
	ClientOnly bool
	Short      bool
	Output     string
	Template   string

	genericclioptions.IOStreams
}

// NewOptions returns initialized Options.
func NewOptions(ioStreams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: ioStreams,
		Template:  "",
	}
}

// NewCmdVersion returns a cobra command for fetching versions.
func NewCmdVersion(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the cli version information",
		Long: `Print the cli version information for the current context.

The version information includes the following fields:
- GitVersion: The semantic version of the build.
- GitCommit: The git commit hash of the build.
- GitTreeState: The state of the git tree, either 'clean' or 'dirty'.
- BuildDate: The date of the build.
- GoVersion: The version of Go used to compile the binary.
- Compiler: The compiler used to compile the binary.
- Platform: The platform (OS/Architecture) for which the binary was built.`,
		Example: versionExample,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
		},
	}

	cmd.Flags().BoolVar(&o.Short, "short", o.Short, "If true, print just the version number.")
	cmd.Flags().StringVarP(&o.Output, "output", "", o.Output, "One of 'yaml' or 'json'.")
	cmd.Flags().StringVarP(&o.Template, "template", "t", o.Template, "Template string to format the version output.")

	return cmd
}

// Validate validates the provided options.
func (o *Options) Validate() error {
	if o.Output != "" && o.Output != "yaml" && o.Output != "json" {
		return errors.New("Invalid output format. Please use 'yaml' or 'json'.")
	}

	return nil
}

// Run executes version command.
func (o *Options) Run() error {
	var serverErr error
	versionInfo := version.Get()

	if o.Template != "" {
		tmpl, err := template.New("version").Parse(o.Template)
		if err != nil {
			return fmt.Errorf("Failed to parse template: %v", err)
		}
		err = tmpl.Execute(o.Out, versionInfo)
		if err != nil {
			return err
		}
	} else {
		switch o.Output {
		case "":
			if o.Short {
				fmt.Fprintf(o.Out, "%s\n", versionInfo.GitVersion)
			} else {
				fmt.Fprintf(o.Out, "Version: %s\n", versionInfo.GitVersion)
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
			return fmt.Errorf("Invalid output format: %q. Please use 'yaml' or 'json'.", o.Output)
		}
	}

	return serverErr
}
