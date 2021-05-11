package c8yfetcher

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type AgentFetcher struct {
	client *c8y.Client
}

func NewAgentFetcher(client *c8y.Client) *AgentFetcher {
	return &AgentFetcher{
		client: client,
	}
}

func (f *AgentFetcher) getByID(id string) ([]fetcherResultSet, error) {
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

func (f *AgentFetcher) getByName(name string) ([]fetcherResultSet, error) {
	mcol, _, err := f.client.Inventory.GetManagedObjects(
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
		results[i] = fetcherResultSet{
			ID:    device.ID,
			Name:  device.Name,
			Value: mcol.Items[i],
		}
	}

	return results, nil
}
