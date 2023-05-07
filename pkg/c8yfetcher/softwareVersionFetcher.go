package c8yfetcher

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type SoftwareVersionFetcher struct {
	*ManagedObjectFetcher
}

func NewSoftwareVersionFetcher(factory *cmdutil.Factory, software string) *SoftwareVersionFetcher {
	return &SoftwareVersionFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			CumulocityFetcher: &CumulocityFetcher{
				factory: factory,
			},
			Query: func(s string) string {
				// Check
				client, err := factory.Client()
				if err != nil {
					return ""
				}

				if !IsID(software) {
					// Lookup software by name
					moSoftware, _, err := client.Software.GetSoftwareByName(WithDisabledDryRunContext(client), software, &c8y.PaginationOptions{
						PageSize: 5,
					})
					if err == nil && moSoftware != nil && len(moSoftware.ManagedObjects) > 0 {
						software = moSoftware.ManagedObjects[0].ID
					}
				}

				if IsID(software) {
					return fmt.Sprintf("(type eq 'c8y_SoftwareBinary') and not(has(c8y_Patch)) and c8y_Software.version eq '%s' and (bygroupid(%s))", s, software)
				}
				return fmt.Sprintf("(type eq 'c8y_SoftwareBinary') and not(has(c8y_Patch)) and c8y_Software.version eq '%s'", s)
			},
		},
	}
}
