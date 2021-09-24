package c8yfetcher

import (
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type FirmwareFetcher struct {
	*ManagedObjectFetcher
}

func NewFirmwareFetcher(client *c8y.Client) *FirmwareFetcher {
	return &FirmwareFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			client: client,
			Query: func(s string) string {
				return fmt.Sprintf("(type eq 'c8y_Firmware') and name eq '%s'", s)
			},
		},
	}
}
