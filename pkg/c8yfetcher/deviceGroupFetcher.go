package c8yfetcher

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type DeviceGroupFetcher struct {
	*CumulocityFetcher
}

func NewDeviceGroupFetcher(factory *cmdutil.Factory) *DeviceGroupFetcher {
	return &DeviceGroupFetcher{
		CumulocityFetcher: &CumulocityFetcher{
			factory: factory,
		},
	}
}

func (f *DeviceGroupFetcher) getByID(id string) ([]fetcherResultSet, error) {
	mo, resp, err := f.Client().Inventory.GetManagedObject(
		c8y.WithDisabledDryRunContext(context.Background()),
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
		Value: resp.JSON(),
	}
	return results, nil
}

func (f *DeviceGroupFetcher) getByName(name string) ([]fetcherResultSet, error) {
	mcol, _, err := f.Client().Inventory.GetManagedObjects(
		c8y.WithDisabledDryRunContext(context.Background()),
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
