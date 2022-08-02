package c8yfetcher

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type HostedApplicationFetcher struct {
	client *c8y.Client
	*DefaultFetcher
}

func NewHostedApplicationFetcher(client *c8y.Client) *HostedApplicationFetcher {
	return &HostedApplicationFetcher{
		client: client,
	}
}

func (f *HostedApplicationFetcher) getByID(id string) ([]fetcherResultSet, error) {
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
		Value: resp.JSON(),
	}
	return results, nil
}

// getByName returns applications matching a given using regular expression
func (f *HostedApplicationFetcher) getByName(name string) ([]fetcherResultSet, error) {
	col, _, err := f.client.Application.GetApplications(
		WithDisabledDryRunContext(f.client),
		&c8y.ApplicationOptions{
			PaginationOptions: *c8y.NewPaginationOptions(2000),

			Type: c8y.ApplicationTypeHosted,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch applications")
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
		if app.Type == "HOSTED" && pattern.MatchString(app.Name) {

			// Ignore applications which don't match the owner
			// so that it can overwrite existing applications such as cockpit and devicemanagement.
			// Otherwise it will always match the in-built apps
			if f.client.TenantName != "" {
				if app.Owner != nil && app.Owner.Tenant != nil && app.Owner.Tenant.ID != "" {
					if app.Owner.Tenant.ID != f.client.TenantName {
						continue
					}
				}
			}

			results = append(results, fetcherResultSet{
				ID:    app.ID,
				Name:  app.Name,
				Value: col.Items[i],
			})
		}

	}

	return results, nil
}

// FindHostedApplications returns hosted applications given either an id or search text
// @values: An array of ids, or names (with wildcards)
// @lookupID: Lookup the data if an id is given. If a non-id text is given, the result will always be looked up.
func FindHostedApplications(client *c8y.Client, values []string, lookupID bool, format string) ([]entityReference, error) {
	f := NewHostedApplicationFetcher(client)

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
