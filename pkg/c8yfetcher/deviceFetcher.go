package c8yfetcher

import (
	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type DeviceFetcher struct {
	*CumulocityFetcher
}

func NewDeviceFetcher(factory *cmdutil.Factory) *DeviceFetcher {
	return &DeviceFetcher{
		CumulocityFetcher: &CumulocityFetcher{
			factory: factory,
		},
	}
}

func (f *DeviceFetcher) getByID(id string) ([]fetcherResultSet, error) {
	mo, resp, err := f.Client().Inventory.GetManagedObject(
		WithDisabledDryRunContext(f.Client()),
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

func (f *DeviceFetcher) getByName(name string) ([]fetcherResultSet, error) {
	mcol, _, err := f.Client().Inventory.GetDevicesByName(
		WithDisabledDryRunContext(f.Client()),
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
