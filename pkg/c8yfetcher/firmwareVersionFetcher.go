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
			Query: func(s string) (string, error) {
				client, err := factory.Client()
				if err != nil {
					return "", err
				}

				// Use default value of a non-existent id so that the query does not return anything
				// Use a default (non-existent) managed object to ensure the inventory
				// query is still valid when the firmware is not found (for whatever reason)
				firmwareID := "0"
				if IsID(firmware) {
					firmwareID = firmware
				} else {
					// Lookup firmware by name
					res, _, err := client.Firmware.GetFirmwareByName(c8y.WithDisabledDryRunContext(context.Background()), firmware, c8y.NewPaginationOptions(5))
					if err != nil {
						return "", NewQueryBuildErr(err)
					}
					if len(res.ManagedObjects) > 0 {
						firmwareID = res.ManagedObjects[0].ID
					}
				}

				patchFilter := "has(c8y_Patch)"
				if !includePatch {
					patchFilter = "not(" + patchFilter + ")"
				}

				return fmt.Sprintf("(type eq 'c8y_FirmwareBinary') and %s and c8y_Firmware.version eq '%s' and (bygroupid(%s))", patchFilter, s, firmwareID), nil
			},
		},
	}
}
