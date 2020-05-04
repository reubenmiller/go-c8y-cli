package cmd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type microserviceFetcher struct {
	client *c8y.Client
}

func newMicroserviceFetcher(client *c8y.Client) *microserviceFetcher {
	return &microserviceFetcher{
		client: client,
	}
}

func (f *microserviceFetcher) getByID(id string) ([]fetcherResultSet, error) {
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
func (f *microserviceFetcher) getByName(name string) ([]fetcherResultSet, error) {
	col, _, err := client.Application.GetApplications(
		context.Background(),
		&c8y.ApplicationOptions{
			PaginationOptions: *c8y.NewPaginationOptions(2000),

			Type: c8y.ApplicationTypeMicroservice,
		},
	)

	pattern, err := regexp.Compile(name)

	if err != nil {
		return nil, errors.Wrap(err, "invalid regex")
	}

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 0)

	for i, app := range col.Applications {
		if app.Type == "MICROSERVICE" && pattern.MatchString(app.Name) {
			results = append(results, fetcherResultSet{
				ID:    app.ID,
				Name:  app.Name,
				Value: col.Items[i],
			})
		}

	}

	return results, nil
}

// getMicroserviceSlice returns the microservice (application) id and name
// returns raw strings, lookuped values, and errors
func getMicroserviceSlice(cmd *cobra.Command, args []string, name string) ([]string, []string, error) {

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

	refs, err := findMicroservices(values, true)

	if err != nil {
		return nil, nil, err
	}

	results, _ := getFetchedResultsAsString(refs)

	return results, results, nil
}

// findMicroservices returns microservices given either an id or search text
// @values: An array of ids, or names (with wildcards)
// @lookupID: Lookup the data if an id is given. If a non-id text is given, the result will always be looked up.
func findMicroservices(values []string, lookupID bool) ([]entityReference, error) {
	f := newMicroserviceFetcher(client)

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
