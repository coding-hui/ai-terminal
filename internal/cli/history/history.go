package history

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"

	"github.com/coding-hui/iam/pkg/cli/genericclioptions"

	"github.com/coding-hui/ai-terminal/internal/cli/options"
	"github.com/coding-hui/ai-terminal/internal/cli/util"
	"github.com/coding-hui/ai-terminal/internal/util/templates"
)

var historyExample = templates.Examples(`
		# Managing session history:
          ai history ls
`)

// Options is a struct to support history command.
type Options struct {
	model     *options.ModelOptions
	datastore *options.DataStoreOptions

	genericclioptions.IOStreams
}

// NewOptions returns initialized Options.
func NewOptions(ioStreams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: ioStreams,
	}
}

// NewCmdHistory returns a cobra command for manager history.
func NewCmdHistory(ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:     "history",
		Short:   "Managing chat session history.",
		Long:    "Managing chat session history.",
		Example: historyExample,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run(args))
		},
		PostRunE: func(c *cobra.Command, args []string) error {
			return nil
		},
	}

	o.model = options.NewModelFlags(false)
	o.model.AddFlags(cmd.Flags())

	o.datastore = options.NewDatastoreOptions(false)
	o.datastore.AddFlags(cmd.Flags())

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = viper.BindPFlag(flag.Name, flag)
	})

	cmd.AddCommand(newCmdLsHistory(o.model, o.datastore))

	return cmd
}

// Validate validates the provided options.
func (o *Options) Validate() error {
	if err := o.model.Validate(); err != nil {
		return err
	}
	if err := o.datastore.Validate(); err != nil {
		return err
	}
	return nil
}

// Run executes history command.
func (o *Options) Run(args []string) error {
	klog.InfoS("history start", "args", args)
	return nil
}
