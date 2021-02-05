package cmd

import (
	"context"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type roleFetcher struct {
	client *c8y.Client
}

func newRoleFetcher(client *c8y.Client) *roleFetcher {
	return &roleFetcher{
		client: client,
	}
}

func (f *roleFetcher) getByID(id string) ([]fetcherResultSet, error) {
	role, resp, err := client.User.GetRole(
		context.Background(),
		id,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 1)
	results[0] = fetcherResultSet{
		ID:    role.ID,
		Name:  role.Name,
		Self:  role.Self,
		Value: *resp.JSON,
	}
	return results, nil
}

func (f *roleFetcher) getByName(name string) ([]fetcherResultSet, error) {
	roles, _, err := client.User.GetRoles(
		context.Background(),
		&c8y.RoleOptions{
			PaginationOptions: *c8y.NewPaginationOptions(100),
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, len(roles.Roles))

	for i, user := range roles.Roles {
		if isMatch, _ := matcher.MatchWithWildcards(user.Name, name); !isMatch {
			continue
		}
		results = append(results, fetcherResultSet{
			ID:    user.ID,
			Name:  user.Name,
			Self:  user.Self,
			Value: roles.Items[i],
		})
	}

	return results, nil
}
