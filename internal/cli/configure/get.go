package configure

import (
	"fmt"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type get struct {
	genericclioptions.IOStreams
	Template string
}

func newCmdGetConfig(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	g := &get{
		IOStreams: ioStreams,
		Template:  "",
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a specific config value using Go template syntax",
		Example: `  # Get a specific config value
  ait config get -t '{{.SomeConfigKey}}'
  
  # Get nested config value
  ait config get -t '{{.SomeSection.SomeKey}}'`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if g.Template == "" {
				return errbook.New("template flag is required")
			}
			return g.getConfigValue(cfg)
		},
	}

	cmd.Flags().StringVarP(&g.Template, "template", "t", g.Template, "Go template string to extract config value (required)")

	return cmd
}

func (g *get) getConfigValue(cfg *options.Config) error {
	tmpl, err := template.New("config").Parse(g.Template)
	if err != nil {
		return errbook.Wrap("couldn't parse template", err)
	}

	err = tmpl.Execute(g.Out, cfg)
	if err != nil {
		return errbook.Wrap("couldn't render template", err)
	}

	// Add newline for clean output
	fmt.Fprintln(g.Out)
	return nil
}
