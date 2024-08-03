package options

import (
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	FlagDatastoreType     = "datastore.type"
	FlagDatastoreUrl      = "datastore.url"
	FlagDatastoreUsername = "datastore.username"
	FlagDatastorePassword = "datastore.password"
)

type DataStoreFlags struct {
	Type     *string
	Url      *string
	Username *string
	Password *string
}

// NewDatastoreFlags returns DataStoreFlags with default values set.
func NewDatastoreFlags(dsType string) *DataStoreFlags {
	return &DataStoreFlags{
		Type: pointer.ToString(dsType),
		Url:  pointer.ToString(""),
	}
}

// AddFlags binds client configuration flags to a given cmd.
func (d *DataStoreFlags) AddFlags(flags *pflag.FlagSet) {
	if d.Type != nil {
		flags.StringVar(d.Type, FlagDatastoreType, *d.Type, "Datastore provider type")
		_ = viper.BindPFlag(FlagDatastoreType, flags.Lookup(FlagDatastoreType))
	}
	if d.Url != nil {
		flags.StringVar(d.Url, FlagDatastoreUrl, *d.Url, "Datastore connection url")
		_ = viper.BindPFlag(FlagDatastoreUrl, flags.Lookup(FlagDatastoreUrl))
	}
	if d.Username != nil {
		flags.StringVar(d.Username, FlagDatastoreUsername, *d.Username, "Datastore username")
		_ = viper.BindPFlag(FlagDatastoreUsername, flags.Lookup(FlagDatastoreUsername))
	}
	if d.Password != nil {
		flags.StringVar(d.Password, FlagDatastorePassword, *d.Password, "Datastore password")
		_ = viper.BindPFlag(FlagDatastorePassword, flags.Lookup(FlagDatastorePassword))
	}
}

func (d *DataStoreFlags) Validate() error {
	if d.Type != nil {
		dsType := *d.Type
		if dsType != "memory" && dsType != "mongo" {
			return fmt.Errorf("invalid datastore type: %s", dsType)
		}
	}
	return nil
}
