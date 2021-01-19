package cmd

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
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

func (f *agentFetcher) getByName(name string) ([]fetcherResultSet, error) {
	mcol, _, err := client.Inventory.GetManagedObjects(
		context.Background(),
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

// getFormattedAgentSlice returns the agent id and name
// returns raw strings, looked up values, and errors
func getFormattedAgentSlice(cmd *cobra.Command, args []string, name string) ([]string, []string, error) {
	f := newAgentFetcher(client)

	if !cmd.Flags().Changed(name) {
		// TODO: Read from os.PIPE
		pipedInput, err := getPipe()
		if err != nil {
			Logger.Debug("No pipeline input detected")
		} else {
			fmt.Printf("PIPED Input: %s\n", pipedInput)
			return nil, nil, nil
		}
	}

	values, err := cmd.Flags().GetStringSlice(name)
	if err != nil {
		Logger.Warning("Flag is missing", err)
	}

	values = ParseValues(append(values, args...))

	formattedValues, err := lookupEntity(f, values, false)

	if err != nil {
		Logger.Warningf("Failed to fetch entities. %s", err)
		return values, nil, err
	}

	results := []string{}

	invalidLookups := []string{}
	for _, item := range formattedValues {
		if item.ID != "" {
			if item.Name != "" {
				results = append(results, fmt.Sprintf("%s|%s", item.ID, item.Name))
			} else {
				results = append(results, item.ID)
			}
		} else {
			if item.Name != "" {
				invalidLookups = append(invalidLookups, item.Name)
			}
		}
	}

	var errors error

	if len(invalidLookups) > 0 {
		errors = fmt.Errorf("no results %v", invalidLookups)
	}

	return values, results, errors
}
