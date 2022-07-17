package c8yfetcher

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type MicroserviceFetcher struct {
	client *c8y.Client
	*DefaultFetcher
}

func NewMicroserviceFetcher(client *c8y.Client) *MicroserviceFetcher {
	return &MicroserviceFetcher{
		client: client,
	}
}

func (f *MicroserviceFetcher) getByID(id string) ([]fetcherResultSet, error) {
	app, resp, err := f.client.Application.GetApplication(
		WithDisabledDryRunContext(f.client),
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
func (f *MicroserviceFetcher) getByName(name string) ([]fetcherResultSet, error) {
	col, _, err := f.client.Application.GetApplications(
		WithDisabledDryRunContext(f.client),
		&c8y.ApplicationOptions{
			PaginationOptions: *c8y.NewPaginationOptions(2000),

			Type: c8y.ApplicationTypeMicroservice,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch microservices")
	}

	pattern, err := regexp.Compile("^" + regexp.QuoteMeta(name) + "$")

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

// findMicroservices returns microservices given either an id or search text
// @values: An array of ids, or names (with wildcards)
// @lookupID: Lookup the data if an id is given. If a non-id text is given, the result will always be looked up.
func FindMicroservices(client *c8y.Client, values []string, lookupID bool, format string) ([]entityReference, error) {
	f := NewMicroserviceFetcher(client)

	formattedValues, err := lookupEntity(f, values, lookupID, format)

	if err != nil {
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
