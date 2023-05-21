package c8yfetcher

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
)

type ConfigurationFetcher struct {
	*ManagedObjectFetcher
}

func NewConfigurationFetcher(factory *cmdutil.Factory) *ConfigurationFetcher {
	return &ConfigurationFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			CumulocityFetcher: &CumulocityFetcher{
				factory: factory,
			},
			Query: func(s string) string {
				return fmt.Sprintf("(type eq 'c8y_ConfigurationDump') and name eq '%s'", s)
			},
		},
	}
}
