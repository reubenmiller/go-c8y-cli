package c8yfetcher

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type RemoteAccessFetcher struct {
	*CumulocityFetcher

	ManagedObjectID string
}

func DetectRemoteAccessConfiguration(client *c8y.Client, mo_id string, name string) (*c8y.RemoteAccessConfiguration, error) {
	configs, _, err := client.RemoteAccess.GetConfigurations(
		c8y.WithDisabledDryRunContext(context.Background()),
		mo_id,
		&c8y.RemoteAccessCollectionOptions{
			PaginationOptions: *c8y.NewPaginationOptions(50),
		})

	if err != nil {
		return nil, err
	}

	passthroughConfigs := make([]c8y.RemoteAccessConfiguration, 0)

	for _, config := range configs {
		if strings.EqualFold(config.Protocol, c8y.RemoteAccessProtocolPassthrough) {
			passthroughConfigs = append(passthroughConfigs, config)
		}
	}

	if name != "" {
		// Exact match
		for _, config := range passthroughConfigs {
			if strings.EqualFold(config.Name, name) {
				return &config, nil
			}
		}

		// Regex match
		pattern, patternErr := matcher.ConvertWildcardToRegex(name)
		if patternErr == nil {
			// Match by regex
			for _, config := range passthroughConfigs {
				if pattern.MatchString(config.Name) {
					return &config, nil
				}
			}
		}

		return nil, fmt.Errorf("configuration not found. name=%s", name)
	}

	if len(passthroughConfigs) == 0 {
		return nil, fmt.Errorf("configuration not found. name=%s", name)
	}

	//
	// Guess the configuration
	//

	// There is only one, so let's use that
	if len(passthroughConfigs) == 1 {
		return &passthroughConfigs[0], nil
	}

	// Use first match which uses port 22
	for _, config := range passthroughConfigs {
		if config.Port == 22 {
			return &config, nil
		}
	}

	// Use first match with ssh in its name
	for _, config := range passthroughConfigs {
		if strings.Contains(strings.ToLower(config.Name), "ssh") {
			return &config, nil
		}
	}

	// Finally just try the first
	return &passthroughConfigs[0], nil
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
