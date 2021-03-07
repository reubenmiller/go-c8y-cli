package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type agentFetcher struct {
	client *c8y.Client
}

func newAgentFetcher(client *c8y.Client) *agentFetcher {
	return &agentFetcher{
		client: client,
	}
}

func (f *agentFetcher) getByID(id string) ([]fetcherResultSet, error) {
	mo, resp, err := client.Inventory.GetManagedObject(
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

func (f *agentFetcher) getByName(name string) ([]fetcherResultSet, error) {
	mcol, _, err := client.Inventory.GetManagedObjects(
		WithDisabledDryRunContext(f.client),
		&c8y.ManagedObjectOptions{
			// fmt.Sprintf("$filter=%s+$orderby=name", filter)
			Query:             fmt.Sprintf("$filter=has(com_cumulocity_model_Agent) and name eq '%s' $orderby=name", name),
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
