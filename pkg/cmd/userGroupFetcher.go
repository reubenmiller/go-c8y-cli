package cmd

import (
	"context"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type userGroupFetcher struct {
	client *c8y.Client
}

func newUserGroupFetcher(client *c8y.Client) *userGroupFetcher {
	return &userGroupFetcher{
		client: client,
	}
}

func (f *userGroupFetcher) getByID(id string) ([]fetcherResultSet, error) {
	group, resp, err := client.User.GetGroup(
		context.Background(),
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

func (f *userGroupFetcher) getByName(name string) ([]fetcherResultSet, error) {
	groups, _, err := client.User.GetGroups(
		context.Background(),
		&c8y.GroupOptions{
			PaginationOptions: *c8y.NewPaginationOptions(2000),
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, len(groups.Groups))

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
