package c8yfetcher

import (
	"regexp"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type ApplicationFetcher struct {
	client *c8y.Client
	*DefaultFetcher
}

func NewApplicationFetcher(client *c8y.Client) *ApplicationFetcher {
	return &ApplicationFetcher{
		client: client,
	}
}

func (f *ApplicationFetcher) getByID(id string) ([]fetcherResultSet, error) {
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
func (f *ApplicationFetcher) getByName(name string) ([]fetcherResultSet, error) {
	col, _, err := f.client.Application.GetApplications(
		WithDisabledDryRunContext(f.client),
		&c8y.ApplicationOptions{
			PaginationOptions: *c8y.NewPaginationOptions(2000),
		},
	)
	if err != nil {
		return nil, err
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
