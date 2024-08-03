package options

import (
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/spf13/pflag"
)

// Defines flag for ai cli.
const (
	FlagAIConfig          = "config"
	FlagChatID            = "chat-id"
	FlagLogFlushFrequency = "log-flush-frequency"
)

// ConfigFlags composes the set of values necessary
// for obtaining a REST client config.
type ConfigFlags struct {
	Config    *string
	SessionId *string
	ds        *DataStoreFlags
	model     *ModelFlags

	// If set to true, will use persistent client config and
	// propagate the config to the places that need it, rather than
	// loading the config multiple times
	usePersistentConfig bool
}

// NewConfigFlags returns ConfigFlags with default values set.
func NewConfigFlags(usePersistentConfig bool) *ConfigFlags {
	return &ConfigFlags{
		ds:                  NewDatastoreFlags("memory"),
		model:               NewModelFlags(),
		SessionId:           pointer.ToString("temp_session"),
		usePersistentConfig: usePersistentConfig,
	}
}

// AddFlags binds client configuration flags to a given flagset.
func (f *ConfigFlags) AddFlags(flags *pflag.FlagSet) {
	if f.Config != nil {
		flags.StringVar(f.Config, FlagAIConfig, *f.Config, fmt.Sprintf("Path to the %s file to use for CLI requests", FlagAIConfig))
	}
	if f.SessionId != nil {
		flags.StringVar(f.SessionId, FlagChatID, *f.SessionId, "Current session id")
	}
	if f.ds != nil {
		f.ds.AddFlags(flags)
	}
	if f.model != nil {
		f.model.AddFlags(flags)
	}
}
