package c8yfetcher

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type DeviceServiceFetcher struct {
	*ManagedObjectFetcher
}

func NewDeviceServiceFetcher(factory *cmdutil.Factory, device string) *DeviceServiceFetcher {
	return &DeviceServiceFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			CumulocityFetcher: &CumulocityFetcher{
				factory: factory,
			},
			Query: func(s string) string {

				client, err := factory.Client()
				if err != nil {
					return ""
				}

				if !IsID(device) {
					// Lookup software by name
					moDevice, _, err := client.Inventory.GetDevicesByName(c8y.WithDisabledDryRunContext(context.Background()), device, &c8y.PaginationOptions{
						PageSize: 5,
					})
					if err == nil && moDevice != nil && len(moDevice.ManagedObjects) > 0 {
						device = moDevice.ManagedObjects[0].ID
					}
				}

				if IsID(device) {
					return fmt.Sprintf("(type eq 'c8y_Service') and name eq '%s' and (bygroupid(%s))", s, device)
				}
				return fmt.Sprintf("(type eq 'c8y_Service') and name eq '%s'", s)
			},
		},
	}
}
