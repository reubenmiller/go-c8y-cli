package cmd

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type deviceGroupFetcher struct {
	client *c8y.Client
}

func newDeviceGroupFetcher(client *c8y.Client) *deviceGroupFetcher {
	return &deviceGroupFetcher{
		client: client,
	}
}

func (f *deviceGroupFetcher) getByID(id string) ([]fetcherResultSet, error) {
	mo, resp, err := client.Inventory.GetManagedObject(
		context.Background(),
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

func (f *deviceGroupFetcher) getByName(name string) ([]fetcherResultSet, error) {
	mcol, _, err := client.Inventory.GetManagedObjects(
		context.Background(),
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
		results = append(results, fetcherResultSet{
			ID:    device.ID,
			Name:  device.Name,
			Value: mcol.Items[i],
		})
	}

	return results, nil
}
