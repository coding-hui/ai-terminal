package options

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"github.com/coding-hui/common/util/homedir"
	"github.com/coding-hui/iam/pkg/log"

	"github.com/coding-hui/ai-terminal/internal/system"
)

const (
	// RecommendedHomeDir defines the default directory used to place all iam service configurations.
	RecommendedHomeDir = ".config"

	// RecommendedEnvPrefix defines the ENV prefix used by all iam service.
	RecommendedEnvPrefix = "AI"

	// DefaultAIConfigName default ai config name
	DefaultAIConfigName = "ai"
)

// Config is a structure used to configure a AI.
// Its members are sorted roughly in order of importance for composers.
type Config struct {
	DefaultPromptMode string
	Preferences       string
	SystemPrompt      string
	System            *system.Analysis
}

// NewConfig returns a Config struct with the default values.
func NewConfig() *Config {
	return &Config{
		DefaultPromptMode: viper.GetString(FlagDefaultPromptMode),
		Preferences:       viper.GetString(FlagPreferences),
		SystemPrompt:      viper.GetString(FlagDefaultSystemPrompt),
		System:            system.Analyse(),
	}
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
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			if err := viper.WriteConfigAs(filepath.Join(homedir.HomeDir(), RecommendedHomeDir, DefaultAIConfigName)); err != nil {
				log.Warnf("WARNING: viper failed to write default config: %s", err.Error())
			}
			return
		}
		log.Warnf("WARNING: viper failed to discover and load the configuration file: %s", err.Error())
	}
}
