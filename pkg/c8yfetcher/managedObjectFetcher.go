package c8yfetcher

import (
	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type QueryFilter func(string) string

type ManagedObjectFetcher struct {
	client *c8y.Client
	Query  QueryFilter
	*DefaultFetcher
}

func NewManagedObjectFetcher(client *c8y.Client) *ManagedObjectFetcher {
	return &ManagedObjectFetcher{
		client: client,
	}
}

func (f *ManagedObjectFetcher) getByID(id string) ([]fetcherResultSet, error) {
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
		Value: resp.JSON(),
	}
	return results, nil
}

func (f *ManagedObjectFetcher) getByName(name string) ([]fetcherResultSet, error) {
	query := "name eq '" + name + "'"
	if f.Query != nil {
		query = f.Query(name)
	}
	mcol, _, err := f.client.Inventory.GetManagedObjects(
		WithDisabledDryRunContext(f.client),
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
