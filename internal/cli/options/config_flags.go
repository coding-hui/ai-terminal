package options

import (
	"fmt"
	"sync"

	"github.com/AlekSi/pointer"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/coding-hui/wecoding-sdk-go/rest"
	"github.com/coding-hui/wecoding-sdk-go/tools/clientcmd"
)

// Defines flag for ai cli.
const (
	FlagAIConfig            = "config"
	FlagDefaultSystemPrompt = "system-prompt"
	FlagAiModel             = "model"
	FlagAiToken             = "token"
	FlagAiApiBase           = "api-base"
	FlagAiTemperature       = "temperature"
	FlagAiTopP              = "top-p"
	FlagAiMaxTokens         = "max-tokens"

	FlagLogFlushFrequency = "log-flush-frequency"
)

// RESTClientGetter is an interface that the ConfigFlags describe to provide an easier way to mock for commands
// and eliminate the direct coupling to a struct type.  Users may wish to duplicate this type in their own packages
// as per the golang type overlapping.
type RESTClientGetter interface {
	// ToRESTConfig returns restconfig
	ToRESTConfig() (*rest.Config, error)
	// ToRawAIConfigLoader return aiconfig loader as-is
	ToRawAIConfigLoader() clientcmd.ClientConfig
}

var _ RESTClientGetter = &ConfigFlags{}

// ConfigFlags composes the set of values necessary
// for obtaining a REST client config.
type ConfigFlags struct {
	Config      *string
	Token       *string
	Model       *string
	ApiBase     *string
	Temperature *float64
	TopP        *float64
	MaxTokens   *int
	Proxy       *string

	clientConfig clientcmd.ClientConfig
	lock         sync.Mutex
	// If set to true, will use persistent client config and
	// propagate the config to the places that need it, rather than
	// loading the config multiple times
	usePersistentConfig bool
}

// ToRESTConfig implements RESTClientGetter.
// Returns a REST client configuration based on a provided path
// to a .aiconfig file, loading rules, and config flag overrides.
// Expects the AddFlags method to have been called.
func (f *ConfigFlags) ToRESTConfig() (*rest.Config, error) {
	return f.ToRawAIConfigLoader().ClientConfig()
}

// ToRawAIConfigLoader binds config flag values to config overrides
// Returns an interactive clientConfig if the password flag is enabled,
// or a non-interactive clientConfig otherwise.
func (f *ConfigFlags) ToRawAIConfigLoader() clientcmd.ClientConfig {
	if f.usePersistentConfig {
		return f.toRawAIPersistentConfigLoader()
	}

	return f.toRawAIConfigLoader()
}

func (f *ConfigFlags) toRawAIConfigLoader() clientcmd.ClientConfig {
	config := clientcmd.NewConfig()
	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	return clientcmd.NewClientConfigFromConfig(config)
}

// toRawAIPersistentConfigLoader binds config flag values to config overrides
// Returns a persistent clientConfig for propagation.
func (f *ConfigFlags) toRawAIPersistentConfigLoader() clientcmd.ClientConfig {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.clientConfig == nil {
		f.clientConfig = f.toRawAIConfigLoader()
	}

	return f.clientConfig
}

// AddFlags binds client configuration flags to a given flagset.
func (f *ConfigFlags) AddFlags(flags *pflag.FlagSet) {
	if f.Config != nil {
		flags.StringVar(f.Config, FlagAIConfig, *f.Config, fmt.Sprintf("Path to the %s file to use for CLI requests", FlagAIConfig))
	}
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

// NewConfigFlags returns ConfigFlags with default values set.
func NewConfigFlags(usePersistentConfig bool) *ConfigFlags {
	return &ConfigFlags{
		Config:              pointer.ToString(""),
		Token:               pointer.ToString(""),
		Model:               pointer.ToString(""),
		ApiBase:             pointer.ToString(""),
		Temperature:         pointer.ToFloat64(0.5),
		TopP:                pointer.ToFloat64(0.5),
		MaxTokens:           pointer.ToInt(1024),
		usePersistentConfig: usePersistentConfig,
	}
}
