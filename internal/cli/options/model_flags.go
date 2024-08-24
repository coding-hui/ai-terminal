package options

import (
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	FlagDefaultSystemPrompt = "system-prompt"
	FlagAiModel             = "model"
	FlagAiToken             = "token"
	FlagAiApiBase           = "api-base"
	FlagAiTemperature       = "temperature"
	FlagAiTopP              = "top-p"
	FlagAiMaxTokens         = "max-tokens"
	FlagOutputFormat        = "output-format"
	FlagMultiContentEnabled = "multi-content-enabled"
)

type ModelFlags struct {
	Token               *string
	Model               *string
	ApiBase             *string
	Temperature         *float64
	TopP                *float64
	MaxTokens           *int
	Proxy               *string
	OutputFormat        *string
	MultiContentEnabled *bool
}

// NewModelFlags returns ModelFlags with default values set.
func NewModelFlags() *ModelFlags {
	return &ModelFlags{
		Token:        pointer.ToString(""),
		Model:        pointer.ToString(""),
		ApiBase:      pointer.ToString(""),
		Temperature:  pointer.ToFloat64(0.5),
		TopP:         pointer.ToFloat64(0.5),
		MaxTokens:    pointer.ToInt(1024),
		OutputFormat: pointer.ToString(string(MarkdownOutputFormat)),
	}
}

// AddFlags binds client configuration flags to a given flagset.
func (m *ModelFlags) AddFlags(flags *pflag.FlagSet) {
	if m.Token != nil {
		flags.StringVar(m.Token, FlagAiToken, *m.Token, "Api token to use for CLI requests")
		_ = viper.BindPFlag(FlagAiToken, flags.Lookup(FlagAiToken))
	}
	if m.Model != nil {
		flags.StringVar(m.Model, FlagAiModel, *m.Model, "The encoding of the model to be called.")
		_ = viper.BindPFlag(FlagAiModel, flags.Lookup(FlagAiModel))
	}
	if m.ApiBase != nil {
		flags.StringVar(m.ApiBase, FlagAiApiBase, *m.ApiBase, "Interface for the API.")
		_ = viper.BindPFlag(FlagAiApiBase, flags.Lookup(FlagAiApiBase))
	}
	if m.Temperature != nil {
		flags.Float64Var(m.Temperature, FlagAiTemperature, *m.Temperature, "Sampling temperature to control the randomness of the output.")
		_ = viper.BindPFlag(FlagAiTemperature, flags.Lookup(FlagAiTemperature))
	}
	if m.TopP != nil {
		flags.Float64Var(m.TopP, FlagAiTopP, *m.TopP, "Nucleus sampling method to control the probability mass of the output.")
		_ = viper.BindPFlag(FlagAiTopP, flags.Lookup(FlagAiTopP))
	}
	if m.MaxTokens != nil {
		flags.IntVar(m.MaxTokens, FlagAiMaxTokens, *m.MaxTokens, "The maximum number of tokens the model can output.")
		_ = viper.BindPFlag(FlagAiMaxTokens, flags.Lookup(FlagAiMaxTokens))
	}
	if m.OutputFormat != nil {
		flags.StringVarP(m.OutputFormat, FlagOutputFormat, "o", *m.OutputFormat, "Output format. One of: (markdown, raw).")
		_ = viper.BindPFlag(FlagOutputFormat, flags.Lookup(FlagOutputFormat))
	}
	if m.MultiContentEnabled != nil {
		flags.BoolVar(m.MultiContentEnabled, FlagMultiContentEnabled, *m.MultiContentEnabled, "LLM multi content is enabled")
		_ = viper.BindPFlag(FlagMultiContentEnabled, flags.Lookup(FlagMultiContentEnabled))
	}
}

func (m *ModelFlags) Validate() error {
	if m.OutputFormat != nil {
		output := *m.OutputFormat
		if output != string(MarkdownOutputFormat) && output != string(RawOutputFormat) {
			return fmt.Errorf("invalid output format: %s", output)
		}
	}
	return nil
}
