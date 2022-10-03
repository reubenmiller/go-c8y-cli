package c8yfetcher

import (
	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type ChildAdditionFetcher struct {
	client   *c8y.Client
	parentID string
	Query    QueryFilter
	*DefaultFetcher
}

func NewChildAdditionFetcher(client *c8y.Client, parentID string) *ChildAdditionFetcher {
	return &ChildAdditionFetcher{
		client:   client,
		parentID: parentID,
	}
}

func (f *ChildAdditionFetcher) getByID(id string) ([]fetcherResultSet, error) {
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

func (f *ChildAdditionFetcher) getByName(name string) ([]fetcherResultSet, error) {
	query := "name eq '" + name + "'"
	if f.Query != nil {
		query = f.Query(name)
	}
	mcol, _, err := f.client.Inventory.GetChildAdditions(
		WithDisabledDryRunContext(f.client),
		f.parentID,
		&c8y.ManagedObjectOptions{
			Query:             query,
			PaginationOptions: *c8y.NewPaginationOptions(5),
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, len(mcol.References))

	for i, item := range mcol.References {
		results[i] = fetcherResultSet{
			ID:    item.ManagedObject.ID,
			Name:  item.ManagedObject.Name,
			Value: mcol.References[i].ManagedObject.Item,
		}
	}

	return results, nil
}
