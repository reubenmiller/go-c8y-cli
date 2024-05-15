package c8yfetcher

import (
	"context"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type DeviceIdentity struct {
	*CumulocityFetcher
}

func NewDeviceIdentity(factory *cmdutil.Factory) *DeviceIdentity {
	return &DeviceIdentity{
		CumulocityFetcher: &CumulocityFetcher{
			factory: factory,
		},
	}
}

func (f *DeviceIdentity) getByID(id string) ([]fetcherResultSet, error) {
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

func (f *DeviceIdentity) getByName(name string) ([]fetcherResultSet, error) {
	mcol, _, err := f.Client().Inventory.GetDevicesByName(
		c8y.WithDisabledDryRunContext(context.Background()),
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
