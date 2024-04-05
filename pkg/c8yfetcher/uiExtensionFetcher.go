package c8yfetcher

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type UIExtensionFetcher struct {
	*CumulocityFetcher

	// Enable the owner check which will only return microservices owned by the current tenant
	// This option was added to ensure that it will not break the existing behaviour,
	// though this option could be enabled by default in the future if it proves to be more useful
	EnableOwnerCheck bool
}

func NewUIExtensionFetcher(factory *cmdutil.Factory) *UIExtensionFetcher {
	return &UIExtensionFetcher{
		CumulocityFetcher: &CumulocityFetcher{
			factory: factory,
		},
	}
}

func (f *UIExtensionFetcher) getByID(id string) ([]fetcherResultSet, error) {
	app, resp, err := f.Client().Application.GetApplication(
		c8y.WithDisabledDryRunContext(context.Background()),
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
func (f *UIExtensionFetcher) getByName(name string) ([]fetcherResultSet, error) {
	serverOptions := &c8y.ApplicationOptions{
		PaginationOptions: *c8y.NewPaginationOptions(2000),
		Type:              c8y.ApplicationTypeHosted,
	}
	serverOptions.WithHasVersions(true)
	if f.EnableOwnerCheck && f.Client().TenantName != "" {
		// Ignore microservices which don't match the owner
		// so that microservices of sub tenants don't get returned.
		serverOptions.Owner = f.Client().TenantName
	}

	col, _, err := f.Client().Application.GetApplications(
		c8y.WithDisabledDryRunContext(context.Background()),
		serverOptions,
	)

	if err != nil {
		return nil, errors.Wrap(err, "could not fetch ui extension")
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

// Find UI Extensions returns extensions given either an id or search text
// @values: An array of ids, or names (with wildcards)
// @lookupID: Lookup the data if an id is given. If a non-id text is given, the result will always be looked up.
func FindUIExtensions(factory *cmdutil.Factory, values []string, lookupID bool, format string, enableOwnerCheck bool) ([]entityReference, error) {
	f := NewUIExtensionFetcher(factory)
	f.EnableOwnerCheck = enableOwnerCheck

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
