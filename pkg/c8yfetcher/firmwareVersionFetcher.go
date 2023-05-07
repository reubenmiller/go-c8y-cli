package c8yfetcher

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type FirmwareVersionFetcher struct {
	*ManagedObjectFetcher
}

func NewFirmwareVersionFetcher(factory *cmdutil.Factory, firmware string, includePatch bool) *FirmwareVersionFetcher {
	return &FirmwareVersionFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			CumulocityFetcher: &CumulocityFetcher{
				factory: factory,
			},
			Query: func(s string) string {
				client, err := factory.Client()
				if err != nil {
					return ""
				}
				var firmwareID string
				if IsID(firmware) {
					firmwareID = firmware
				} else {
					// Lookup firmware by name
					res, _, err := client.Firmware.GetFirmwareByName(context.Background(), firmware, c8y.NewPaginationOptions(5))
					if err == nil && len(res.ManagedObjects) > 0 {
						firmwareID = res.ManagedObjects[0].ID
					}
				}

				patchFilter := "has(c8y_Patch)"
				if !includePatch {
					patchFilter = "not(" + patchFilter + ")"
				}

				return fmt.Sprintf("(type eq 'c8y_FirmwareBinary') and %s and c8y_Firmware.version eq '%s' and (bygroupid(%s))", patchFilter, s, firmwareID)
			},
		},
	}
}
