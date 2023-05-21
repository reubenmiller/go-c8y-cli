package c8yfetcher

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
)

type FirmwareFetcher struct {
	*ManagedObjectFetcher
}

func NewFirmwareFetcher(factory *cmdutil.Factory) *FirmwareFetcher {
	return &FirmwareFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			CumulocityFetcher: &CumulocityFetcher{
				factory: factory,
			},
			Query: func(s string) string {
				return fmt.Sprintf("(type eq 'c8y_Firmware') and name eq '%s'", s)
			},
		},
	}
}
