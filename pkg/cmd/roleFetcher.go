package cmd

import (
	"strings"

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
		WithDisabledDryRunContext(f.client),
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
	// check if already resolved, so we can safe a lookup
	if strings.Contains(name, "/roles/") {
		return []fetcherResultSet{
			{
				ID:   name,
				Name: name,
				Self: name,
			},
		}, nil
	}
	roles, _, err := client.User.GetRoles(
		WithDisabledDryRunContext(f.client),
		&c8y.RoleOptions{
			PaginationOptions: *c8y.NewPaginationOptions(100),
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by name")
	}

	results := make([]fetcherResultSet, len(roles.Roles))

	for i, role := range roles.Roles {
		if isMatch, _ := matcher.MatchWithWildcards(role.Name, name); !isMatch {
			continue
		}
		results = append(results, fetcherResultSet{
			ID:    role.ID,
			Name:  role.Name,
			Self:  role.Self,
			Value: roles.Items[i],
		})
	}

	return results, nil
}
