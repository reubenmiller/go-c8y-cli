package cmd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type hostedApplicationFetcher struct {
	client *c8y.Client
}

func newHostedApplicationFetcher(client *c8y.Client) *hostedApplicationFetcher {
	return &hostedApplicationFetcher{
		client: client,
	}
}

func (f *hostedApplicationFetcher) getByID(id string) ([]fetcherResultSet, error) {
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
func (f *hostedApplicationFetcher) getByName(name string) ([]fetcherResultSet, error) {
	col, _, err := client.Application.GetApplications(
		context.Background(),
		&c8y.ApplicationOptions{
			PaginationOptions: *c8y.NewPaginationOptions(2000),

			Type: c8y.ApplicationTypeHosted,
		},
	)

	pattern, err := regexp.Compile(regexp.QuoteMeta(name))

	if err != nil {
		return nil, errors.Wrap(err, "invalid regex")
	}

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 0)

	for i, app := range col.Applications {
		if app.Type == "HOSTED" && pattern.MatchString(app.Name) {
			results = append(results, fetcherResultSet{
				ID:    app.ID,
				Name:  app.Name,
				Value: col.Items[i],
			})
		}

	}

	return results, nil
}

// getHostedApplicationSlice returns the hosted application id and name
// returns raw strings, lookuped values, and errors
func getHostedApplicationSlice(cmd *cobra.Command, args []string, name string) ([]string, []string, error) {

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

	values := make([]string, 0)

	if value, err := cmd.Flags().GetString(name); err != nil {
		Logger.Error("Flag is missing", err)
	} else {
		values = append(values, value)
	}

	if len(values) == 0 {
		return nil, nil, fmt.Errorf("Failed to find matching applications")
	}

	refs, err := findHostedApplications(values, true)

	if err != nil {
		return nil, nil, err
	}

	results, _ := getFetchedResultsAsString(refs)

	return values, results, nil
}

// findHostedApplications returns hosted applications given either an id or search text
// @values: An array of ids, or names (with wildcards)
// @lookupID: Lookup the data if an id is given. If a non-id text is given, the result will always be looked up.
func findHostedApplications(values []string, lookupID bool) ([]entityReference, error) {
	f := newHostedApplicationFetcher(client)

	formattedValues, err := lookupEntity(f, values, lookupID)

	if err != nil {
		Logger.Errorf("Failed to fetch entities. %s", err)
		return nil, err
	}

	results := []entityReference{}

	invalidLookups := []string{}
	for _, item := range formattedValues {
		if item.ID != "" {
			results = append(results, item)
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

	return results, errors
}
