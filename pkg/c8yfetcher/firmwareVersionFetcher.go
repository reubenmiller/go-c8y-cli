package c8yfetcher

import (
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type FirmwareVersionFetcher struct {
	*ManagedObjectFetcher
}

func NewFirmwareVersionFetcher(client *c8y.Client) *FirmwareVersionFetcher {
	return &FirmwareVersionFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			client: client,
			Query: func(s string) string {
				return fmt.Sprintf("(type eq 'c8y_FirmwareBinary') and not(has(c8y_Patch)) and c8y_Firmware.version eq '%s'", s)
			},
		},
	}
}
