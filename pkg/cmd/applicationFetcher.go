package cmd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type applicationFetcher struct {
	client *c8y.Client
}

func newApplicationFetcher(client *c8y.Client) *applicationFetcher {
	return &applicationFetcher{
		client: client,
	}
}

func (f *applicationFetcher) getByID(id string) ([]fetcherResultSet, error) {
	app, resp, err := client.Application.GetApplication(
		context.Background(),
		id,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 1)
	results[0] = fetcherResultSet{
		ID:    app.ID,
		Name:  app.Name,
		Value: *resp.JSON,
	}
	return results, nil
}

// getByName returns applications matching a given using regular expression
func (f *applicationFetcher) getByName(name string) ([]fetcherResultSet, error) {
	col, _, err := client.Application.GetApplications(
		context.Background(),
		&c8y.ApplicationOptions{
			PaginationOptions: *c8y.NewPaginationOptions(2000),
		},
	)

	pattern, err := regexp.Compile(name)

	if err != nil {
		return nil, errors.Wrap(err, "invalid regex")
	}

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, len(col.Applications))

	for i, app := range col.Applications {
		if pattern.MatchString(app.Name) {
			results = append(results, fetcherResultSet{
				ID:    app.ID,
				Name:  app.Name,
				Value: col.Items[i],
			})
		}

	}

	return results, nil
}

// getApplicationSlice returns the application id and name
// returns raw strings, lookuped values, and errors
func getApplicationSlice(cmd *cobra.Command, args []string, name string) ([]string, []string, error) {
	f := newApplicationFetcher(client)

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

	values := make([]string, 1)

	if value, err := cmd.Flags().GetString(name); err != nil {
		Logger.Error("Flag is missing", err)
	} else {
		values[0] = value
	}

	// values = ParseValues(append(values, args...))

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
