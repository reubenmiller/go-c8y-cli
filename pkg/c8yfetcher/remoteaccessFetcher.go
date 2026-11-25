package c8yfetcher

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type RemoteAccessFetcher struct {
	*CumulocityFetcher

	ManagedObjectID string
}

func NewRemoteAccessFetcher(factory *cmdutil.Factory, mo_id string) *RemoteAccessFetcher {
	return &RemoteAccessFetcher{
		CumulocityFetcher: &CumulocityFetcher{
			factory: factory,
		},
		ManagedObjectID: mo_id,
	}
}

func (f *RemoteAccessFetcher) getByID(id string) ([]fetcherResultSet, error) {
	config, resp, err := f.Client().RemoteAccess.GetConfiguration(
		c8y.WithDisabledDryRunContext(context.Background()),
		f.ManagedObjectID,
		id,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 1)
	results[0] = fetcherResultSet{
		ID:    config.ID,
		Name:  config.Name,
		Value: resp.JSON(),
	}
	return results, nil
}

// getByName returns applications matching a given using regular expression
func (f *RemoteAccessFetcher) getByName(name string) ([]fetcherResultSet, error) {
	configs, _, err := f.Client().RemoteAccess.GetConfigurations(
		c8y.WithDisabledDryRunContext(context.Background()),
		f.ManagedObjectID,
		nil,
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

	for _, app := range configs {
		if pattern.MatchString(app.Name) {
			results = append(results, fetcherResultSet{
				ID:    app.ID,
				Name:  app.Name,
				Value: app,
			})
		}

	}

	return results, nil
}

// FindRemoteAccessConfigurations returns remote access configurations given either an id or search text
// @values: An array of ids, or names (with wildcards)
// @lookupID: Lookup the data if an id is given. If a non-id text is given, the result will always be looked up.
func FindRemoteAccessConfigurations(factory *cmdutil.Factory, values []string, lookupID bool, format string, mo_id string) ([]entityReference, error) {
	f := NewRemoteAccessFetcher(factory, mo_id)

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
