package c8yfetcher

import (
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type DeviceServiceFetcher struct {
	*ManagedObjectFetcher
}

func NewDeviceServiceFetcher(client *c8y.Client, device string) *DeviceServiceFetcher {
	return &DeviceServiceFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			client: client,
			Query: func(s string) string {
				if !IsID(device) {
					// Lookup software by name
					moDevice, _, err := client.Inventory.GetDevicesByName(WithDisabledDryRunContext(client), device, &c8y.PaginationOptions{
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
