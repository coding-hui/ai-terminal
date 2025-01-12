package configure

import (
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type echo struct {
	genericclioptions.IOStreams
	Template string
}

func newCmdEchoConfig(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	e := &echo{
		IOStreams: ioStreams,
		Template:  "",
	}
	cmd := &cobra.Command{
		Use:          "echo",
		Short:        "Display current configuration settings",
		Example: `  # Show full configuration
  ai cfg echo
  
  # Show specific setting using template
  ai cfg echo -t "{{.Model}}"
  
  # Show API endpoints
  ai cfg echo -t "{{range .APIs}}{{.Name}}: {{.BaseURL}}\n{{end}}"
  
  # Check cache location
  ai cfg echo -t "Cache: {{.DataStore.CachePath}}"`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return e.echoSettings(cfg)
		},
	}

	cmd.Flags().StringVarP(&e.Template, "template", "t", e.Template, "Template string to format the settings output.")

	return cmd
}

func (e *echo) echoSettings(cfg *options.Config) error {
	_, err := os.Stat(cfg.SettingsPath)
	if err != nil {
		return errbook.Wrap("Couldn't read config file.", err)
	}
	inputFile, err := os.Open(cfg.SettingsPath)
	if err != nil {
		return errbook.Wrap("Couldn't open config file.", err)
	}
	defer inputFile.Close() //nolint:errcheck

	if e.Template != "" {
		tmpl, err := template.New("settings").Parse(e.Template)
		if err != nil {
			return errbook.Wrap("Couldn't pares template.", err)
		}
		err = tmpl.Execute(e.Out, cfg)
		if err != nil {
			return errbook.Wrap("Couldn't render template.", err)
		}
		return nil
	}

	_, _ = fmt.Fprintln(e.Out, "Current settings:")
	_, err = io.Copy(e.Out, inputFile)
	if err != nil {
		return errbook.Wrap("Couldn't echo config file.", err)
	}

	return nil
}
