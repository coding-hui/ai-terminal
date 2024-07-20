package options

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"k8s.io/klog/v2"

	"github.com/coding-hui/common/util/homedir"
	"github.com/coding-hui/iam/pkg/log"

	"github.com/coding-hui/ai-terminal/internal/system"
)

const (
	// RecommendedHomeDir defines the default directory used to place all iam service configurations.
	RecommendedHomeDir = ".wecoding"

	// RecommendedEnvPrefix defines the ENV prefix used by all iam service.
	RecommendedEnvPrefix = "AI"
)

// Config is a structure used to configure a AI.
// Its members are sorted roughly in order of importance for composers.
type Config struct {
	DefaultPromptMode string           `yaml:"default-prompt-mode,omitempty" mapstructure:"default-prompt-mode,omitempty"`
	SessionId         string           `yaml:"session-id,omitempty" mapstructure:"session-id,omitempty"`
	Ai                AiConfig         `yaml:"inline" mapstructure:"inline" json:"inline"`
	DataStore         DataStoreOptions `yaml:"datastore" mapstructure:"datastore" json:"datastore"`

	System *system.Analysis
}

type AiConfig struct {
	SystemPrompt string       `yaml:"system-prompt,omitempty" mapstructure:"system-prompt,omitempty"`
	Token        string       `yaml:"token,omitempty" mapstructure:"token,omitempty"`
	Model        string       `yaml:"model,omitempty" mapstructure:"model,omitempty"`
	ApiBase      string       `yaml:"api-base,omitempty" mapstructure:"api-base,omitempty"`
	Temperature  float64      `yaml:"temperature,omitempty" mapstructure:"temperature,omitempty"`
	TopP         float64      `yaml:"top-p,omitempty" mapstructure:"top-p,omitempty"`
	MaxTokens    int          `yaml:"max-tokens,omitempty" mapstructure:"max-tokens,omitempty"`
	Proxy        string       `yaml:"proxy,omitempty" mapstructure:"proxy,omitempty"`
	OutputFormat OutputFormat `yaml:"output-format,omitempty" mapstructure:"output-format,omitempty"`
}

type OutputFormat string

const (
	RawOutputFormat      OutputFormat = "raw"
	MarkdownOutputFormat OutputFormat = "markdown"
)

// NewConfig returns a Config struct with the default values.
func NewConfig() (*Config, error) {
	return &Config{
		SessionId: viper.GetString(FlagChatSessionId),
		Ai: AiConfig{
			SystemPrompt: viper.GetString(FlagDefaultSystemPrompt),
			Token:        viper.GetString(FlagAiToken),
			Model:        viper.GetString(FlagAiModel),
			ApiBase:      viper.GetString(FlagAiApiBase),
			Temperature:  viper.GetFloat64(FlagAiTemperature),
			TopP:         viper.GetFloat64(FlagAiTopP),
			MaxTokens:    viper.GetInt(FlagAiMaxTokens),
			OutputFormat: OutputFormat(viper.GetString(FlagOutputFormat)),
			Proxy:        "",
		},
		DataStore: DataStoreOptions{
			Type:     viper.GetString(FlagDatastoreType),
			Url:      viper.GetString(FlagDatastoreUrl),
			Username: viper.GetString(FlagDatastoreUsername),
			Password: viper.GetString(FlagDatastorePassword),
		},
		System: system.Analyse(),
	}, nil
}

// LoadConfig reads in config file and ENV variables if set.
func LoadConfig(cfg string, defaultName string) {
	if cfg != "" {
		viper.SetConfigFile(cfg)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath(filepath.Join(homedir.HomeDir(), RecommendedHomeDir))
		viper.AddConfigPath("/etc/wecoding/.config")
		viper.SetConfigName(defaultName)
	}

	// Use config file from the flag.
	viper.SetConfigType("yaml")              // set the type of the configuration to yaml.
	viper.AutomaticEnv()                     // read in environment variables that match.
	viper.SetEnvPrefix(RecommendedEnvPrefix) // set ENVIRONMENT variables prefix to IAM.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Debugf("WARNING: viper failed to discover and load the configuration file: %s", err.Error())
	}
}

func WriteConfig(model, apiBase, apiToken string, write bool) (*Config, error) {
	viper.Set(FlagAiModel, model)
	viper.Set(FlagAiApiBase, apiBase)
	viper.Set(FlagAiToken, apiToken)
	viper.SetDefault(FlagAiTemperature, 0.2)
	viper.SetDefault(FlagAiTopP, 0.9)
	viper.SetDefault(FlagAiMaxTokens, 2048)
	viper.SetDefault(FlagDefaultSystemPrompt, "")

	viper.SetConfigType("yaml")

	if write {
		defConfFile := system.GetDefaultConfigFile()
		log.Debugf("Writing config file: %s", defConfFile)
		if err := os.MkdirAll(filepath.Dir(defConfFile), 0755); err != nil {
			klog.Warningf("WARNING: failed to create config dir: %s", err.Error())
		}
		err := viper.WriteConfigAs(defConfFile)
		if err != nil {
			var existErr *viper.ConfigFileAlreadyExistsError
			if !errors.As(err, &existErr) {
				return nil, err
			}
		}
	}

	return NewConfig()
}
