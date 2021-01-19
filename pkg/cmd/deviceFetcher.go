package cmd

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type deviceFetcher struct {
	client *c8y.Client
}

func newDeviceFetcher(client *c8y.Client) *deviceFetcher {
	return &deviceFetcher{
		client: client,
	}
}

func (f *deviceFetcher) getByID(id string) ([]fetcherResultSet, error) {
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

func (f *deviceFetcher) getByName(name string) ([]fetcherResultSet, error) {
	mcol, _, err := client.Inventory.GetDevicesByName(
		context.Background(),
		name,
		c8y.NewPaginationOptions(5),
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

// getFormattedDeviceSlice returns the device id and name
// returns raw strings, lookuped values, and errors
func getFormattedDeviceSlice(cmd *cobra.Command, args []string, name string) ([]string, []string, error) {
	f := newDeviceFetcher(client)

	if !cmd.Flags().Changed(name) {
		// TODO: Read from os.PIPE
		pipedInput, err := getPipe()
		if err != nil {
			Logger.Debug("No pipeline input detected")
		} else {
			Logger.Debugf("PIPED Input: %s\n", pipedInput)
			return nil, nil, nil
		}
	}

	values, err := cmd.Flags().GetStringSlice(name)
	if err != nil {
		Logger.Error("Flag is missing", err)
	}

	values = ParseValues(append(values, args...))

	formattedValues, err := lookupEntity(f, values, false)

	if err != nil {
		Logger.Errorf("Failed to fetch entities. %s", err)
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
