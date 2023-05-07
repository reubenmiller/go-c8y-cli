package c8yfetcher

import (
	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type QueryFilter func(string) string

type ManagedObjectFetcher struct {
	Query QueryFilter
	*CumulocityFetcher
}

func NewManagedObjectFetcher(factory *cmdutil.Factory) *ManagedObjectFetcher {
	return &ManagedObjectFetcher{
		CumulocityFetcher: &CumulocityFetcher{
			factory: factory,
		},
	}
}

func (f *ManagedObjectFetcher) getByID(id string) ([]fetcherResultSet, error) {
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

func (f *ManagedObjectFetcher) getByName(name string) ([]fetcherResultSet, error) {
	query := "name eq '" + name + "'"
	if f.Query != nil {
		query = f.Query(name)
	}
	mcol, _, err := f.Client().Inventory.GetManagedObjects(
		WithDisabledDryRunContext(f.Client()),
		&c8y.ManagedObjectOptions{
			Query:             query,
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
