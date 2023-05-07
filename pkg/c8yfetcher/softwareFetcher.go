package c8yfetcher

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
)

type SoftwareFetcher struct {
	*ManagedObjectFetcher
}

func NewSoftwareFetcher(factory *cmdutil.Factory) *SoftwareFetcher {
	return &SoftwareFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			CumulocityFetcher: &CumulocityFetcher{
				factory: factory,
			},
			Query: func(s string) string {
				return fmt.Sprintf("(type eq 'c8y_Software') and name eq '%s'", s)
			},
		},
	}
}
