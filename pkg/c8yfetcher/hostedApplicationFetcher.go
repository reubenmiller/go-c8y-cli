package c8yfetcher

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type HostedApplicationFetcher struct {
	*CumulocityFetcher
}

func NewHostedApplicationFetcher(factory *cmdutil.Factory) *HostedApplicationFetcher {
	return &HostedApplicationFetcher{
		CumulocityFetcher: &CumulocityFetcher{
			factory: factory,
		},
	}
}

func (f *HostedApplicationFetcher) getByID(id string) ([]fetcherResultSet, error) {
	app, resp, err := f.Client().Application.GetApplication(
		WithDisabledDryRunContext(f.Client()),
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
	col, _, err := f.Client().Application.GetApplications(
		WithDisabledDryRunContext(f.Client()),
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

	// First check for hosted applications owned by the current tenant
	for i, app := range col.Applications {
		if app.Type == "HOSTED" && pattern.MatchString(app.Name) {

			// Ignore applications which don't match the owner
			// so that it can overwrite existing applications such as cockpit and devicemanagement.
			// Otherwise it will always match the in-built apps
			if f.Client().TenantName != "" {
				if app.Owner != nil && app.Owner.Tenant != nil && app.Owner.Tenant.ID != "" {
					if app.Owner.Tenant.ID != f.Client().TenantName {
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

	// If not results are found, then also include any matches (not just in the current tenant)
	if len(results) == 0 {
		for i, app := range col.Applications {
			if app.Type == "HOSTED" && pattern.MatchString(app.Name) {
				if app.Owner != nil && app.Owner.Tenant != nil && app.Owner.Tenant.ID != "" {
					if app.Owner.Tenant.ID != f.Client().TenantName {
						continue
					}
				}
				results = append(results, fetcherResultSet{
					ID:    app.ID,
					Name:  app.Name,
					Value: col.Items[i],
				})
			}
		}
	}

	return results, nil
}

// FindHostedApplications returns hosted applications given either an id or search text
// @values: An array of ids, or names (with wildcards)
// @lookupID: Lookup the data if an id is given. If a non-id text is given, the result will always be looked up.
func FindHostedApplications(factory *cmdutil.Factory, values []string, lookupID bool, format string) ([]entityReference, error) {
	f := NewHostedApplicationFetcher(factory)

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
