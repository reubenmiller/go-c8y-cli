package c8yfetcher

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type DeviceProfileFetcher struct {
	*CumulocityFetcher
}

func NewDeviceProfileFetcher(factory *cmdutil.Factory) *DeviceProfileFetcher {
	return &DeviceProfileFetcher{
		CumulocityFetcher: &CumulocityFetcher{
			factory: factory,
		},
	}
}

func (f *DeviceProfileFetcher) getByID(id string) ([]fetcherResultSet, error) {
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

func (f *DeviceProfileFetcher) getByName(name string) ([]fetcherResultSet, error) {
	mcol, _, err := f.Client().Inventory.GetManagedObjects(
		WithDisabledDryRunContext(f.Client()),
		&c8y.ManagedObjectOptions{
			Query:             fmt.Sprintf("(type eq 'c8y_Profile') and name eq '%s'", name),
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
