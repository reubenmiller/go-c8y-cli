package c8yfetcher

import (
	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type DeviceFetcher struct {
	client *c8y.Client
}

func NewDeviceFetcher(client *c8y.Client) *DeviceFetcher {
	return &DeviceFetcher{
		client: client,
	}
}

func (f *DeviceFetcher) getByID(id string) ([]fetcherResultSet, error) {
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

func (f *DeviceFetcher) getByName(name string) ([]fetcherResultSet, error) {
	mcol, _, err := f.client.Inventory.GetDevicesByName(
		WithDisabledDryRunContext(f.client),
		name,
		c8y.NewPaginationOptions(5),
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
