package c8yfetcher

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type DeviceGroupFetcher struct {
	client *c8y.Client
}

func NewDeviceGroupFetcher(client *c8y.Client) *DeviceGroupFetcher {
	return &DeviceGroupFetcher{
		client: client,
	}
}

func (f *DeviceGroupFetcher) getByID(id string) ([]fetcherResultSet, error) {
	mo, resp, err := f.client.Inventory.GetManagedObject(
		WithDisabledDryRunContext(f.client),
		id,
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 1)
	results[0] = fetcherResultSet{
		ID:    mo.ID,
		Name:  mo.Name,
		Value: *resp.JSON,
	}
	return results, nil
}

func (f *DeviceGroupFetcher) getByName(name string) ([]fetcherResultSet, error) {
	mcol, _, err := f.client.Inventory.GetManagedObjects(
		WithDisabledDryRunContext(f.client),
		&c8y.ManagedObjectOptions{
			Query:             fmt.Sprintf("has(c8y_IsDeviceGroup) and name eq '%s'", name),
			PaginationOptions: *c8y.NewPaginationOptions(5),
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, len(mcol.ManagedObjects))

	for i, device := range mcol.ManagedObjects {
		results[i] = fetcherResultSet{
			ID:    device.ID,
			Name:  device.Name,
			Value: mcol.Items[i],
		}
	}

	return results, nil
}
