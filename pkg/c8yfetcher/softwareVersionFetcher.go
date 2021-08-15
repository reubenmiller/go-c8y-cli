package c8yfetcher

import (
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type SoftwareVersionFetcher struct {
	*ManagedObjectFetcher
}

func NewSoftwareVersionFetcher(client *c8y.Client, software string) *SoftwareVersionFetcher {
	return &SoftwareVersionFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			client: client,
			Query: func(s string) string {
				// Check
				if IsID(software) {
					return fmt.Sprintf("(type eq 'c8y_SoftwareBinary') and not(has(c8y_Patch)) and c8y_Software.version eq '%s' and (bygroupid(%s))", s, software)
				}
				if software != "" {

				}
				return fmt.Sprintf("(type eq 'c8y_SoftwareBinary') and not(has(c8y_Patch)) and c8y_Software.version eq '%s'", s)
			},
		},
	}
}