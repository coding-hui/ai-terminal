package configure

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/coding-hui/ai-terminal/internal/errbook"
	"github.com/coding-hui/ai-terminal/internal/options"
	"github.com/coding-hui/ai-terminal/internal/ui/console"
	"github.com/coding-hui/ai-terminal/internal/util/genericclioptions"
)

type reset struct {
	genericclioptions.IOStreams
}

func newCmdResetConfig(ioStreams genericclioptions.IOStreams, cfg *options.Config) *cobra.Command {
	r := &reset{
		IOStreams: ioStreams,
	}
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to default values",
		Example: `  # Reset settings with backup
  ai cfg reset
  
  # Reset and open new config in editor
  ai cfg reset && ai cfg`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.resetSettings(cfg)
		},
	}

	return cmd
}

func (r *reset) resetSettings(cfg *options.Config) error {
	_, err := os.Stat(cfg.SettingsPath)
	if err != nil {
		return errbook.Wrap("Couldn't read config file.", err)
	}
	inputFile, err := os.Open(cfg.SettingsPath)
	if err != nil {
		return errbook.Wrap("Couldn't open config file.", err)
	}
	defer inputFile.Close() //nolint:errcheck
	outputFile, err := os.Create(cfg.SettingsPath + ".bak")
	if err != nil {
		return errbook.Wrap("Couldn't backup config file.", err)
	}
	defer outputFile.Close() //nolint:errcheck
	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return errbook.Wrap("Couldn't write config file.", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(cfg.SettingsPath)
	if err != nil {
		return errbook.Wrap("Couldn't remove config file.", err)
	}
	err = options.WriteConfigFile(cfg.SettingsPath)
	if err != nil {
		return errbook.Wrap("Couldn't write new config file.", err)
	}
	if !cfg.Quiet {
		_, _ = fmt.Fprintln(r.Out, "\nSettings restored to defaults!")
		_, _ = fmt.Fprintf(r.Out,
			"\n  %s %s\n\n",
			console.StderrStyles().Comment.Render("Your old settings have been saved to:"),
			console.StderrStyles().Link.Render(cfg.SettingsPath+".bak"),
		)
	}
	return nil
}
