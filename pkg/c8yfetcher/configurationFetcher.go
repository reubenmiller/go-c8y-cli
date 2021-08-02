package c8yfetcher

import (
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type ConfigurationFetcher struct {
	*ManagedObjectFetcher
}

func NewConfigurationFetcher(client *c8y.Client) *ConfigurationFetcher {
	return &ConfigurationFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			client: client,
			Query: func(s string) string {
				return fmt.Sprintf("(type eq 'c8y_ConfigurationDump') and name eq '%s'", s)
			},
		},
	}
}
