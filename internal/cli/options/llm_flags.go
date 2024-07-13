package options

import (
	"github.com/AlekSi/pointer"
	flag "github.com/spf13/pflag"
)

type LLMFlags struct {
	Token       *string
	Model       *string
	ApiBase     *string
	Temperature *float64
	TopP        *float64
	MaxTokens   *int
	Proxy       *string

	// If set to true, will use persistent client config and
	// propagate the config to the places that need it, rather than
	// loading the config multiple times
	usePersistentConfig bool
}

// AddFlags binds client configuration flags to a given flagset.
func (f *LLMFlags) AddFlags(flags *flag.FlagSet) {
	if f.Token != nil {
		flags.StringVar(f.Token, FlagAiToken, *f.Token, "Api token to use for CLI requests")
	}
	if f.Model != nil {
		flags.StringVar(f.Model, FlagAiModel, *f.Model, "The encoding of the model to be called.")
	}
	if f.ApiBase != nil {
		flags.StringVar(f.ApiBase, FlagAiApiBase, *f.ApiBase, "Interface for the API.")
	}
	if f.Temperature != nil {
		flags.Float64Var(f.Temperature, FlagAiTemperature, *f.Temperature, "Sampling temperature to control the randomness of the output.")
	}
	if f.TopP != nil {
		flags.Float64Var(f.TopP, FlagAiTopP, *f.TopP, "Nucleus sampling method to control the probability mass of the output.")
	}
	if f.MaxTokens != nil {
		flags.IntVar(f.MaxTokens, FlagAiMaxTokens, *f.MaxTokens, "The maximum number of tokens the model can output.")
	}
}

// NewLLMFlags returns LLMFlags with default values set.
func NewLLMFlags(usePersistentConfig bool) *LLMFlags {
	return &LLMFlags{
		Token:               pointer.ToString(""),
		Model:               pointer.ToString(""),
		ApiBase:             pointer.ToString(""),
		Temperature:         pointer.ToFloat64(0.5),
		TopP:                pointer.ToFloat64(0.5),
		MaxTokens:           pointer.ToInt(1024),
		usePersistentConfig: usePersistentConfig,
	}
}
