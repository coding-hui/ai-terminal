package options

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagDatastoreType     = "datastore.type"
	FlagDatastoreUrl      = "datastore.url"
	FlagDatastoreUsername = "datastore.username"
	FlagDatastorePassword = "datastore.password"
)

type DataStoreOptions struct {
	Type     string
	Url      string
	Username string
	Password string
}

// AddFlags binds client configuration flags to a given flagset.
func (d *DataStoreOptions) AddFlags(flags *flag.FlagSet) {
	flags.StringVar(&d.Type, FlagDatastoreType, d.Type, "Datastore provider type")
	flags.StringVar(&d.Url, FlagDatastoreUrl, d.Url, "Datastore connection url")
	flags.StringVar(&d.Username, FlagDatastoreUsername, d.Username, "Datastore username")
	flags.StringVar(&d.Password, FlagDatastorePassword, d.Password, "Datastore password")
}

// NewDatastoreOptions returns DataStoreOptions with default values set.
func NewDatastoreOptions(usePersistentConfig bool) *DataStoreOptions {
	return &DataStoreOptions{
		Type:     "mongo",
		Url:      "mongodb://localhost:27017/",
		Username: "",
		Password: "",
	}
}

func (d *DataStoreOptions) Validate() error {
	return nil
}
