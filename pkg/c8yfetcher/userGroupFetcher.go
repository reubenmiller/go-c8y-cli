package c8yfetcher

import (
	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type UserGroupFetcher struct {
	client *c8y.Client
	*DefaultFetcher
}

func NewUserGroupFetcher(client *c8y.Client) *UserGroupFetcher {
	return &UserGroupFetcher{
		client: client,
	}
}

func (f *UserGroupFetcher) getByID(id string) ([]fetcherResultSet, error) {
	group, resp, err := f.client.User.GetGroup(
		WithDisabledDryRunContext(f.client),
		id,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 1)
	results[0] = fetcherResultSet{
		ID:    group.GetID(),
		Name:  group.Name,
		Value: *resp.JSON,
	}
	return results, nil
}

func (f *UserGroupFetcher) getByName(name string) ([]fetcherResultSet, error) {
	groups, _, err := f.client.User.GetGroups(
		WithDisabledDryRunContext(f.client),
		&c8y.GroupOptions{
			PaginationOptions: *c8y.NewPaginationOptions(2000),
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 0)

	for i, group := range groups.Groups {
		// Does pattern match name (using wildcards)
		if isMatch, _ := matcher.MatchWithWildcards(group.Name, name); !isMatch {
			continue
		}
		results = append(results, fetcherResultSet{
			ID:    group.GetID(),
			Name:  group.Name,
			Self:  group.Self,
			Value: groups.Items[i],
		})
	}

	return results, nil
}
