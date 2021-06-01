package c8yfetcher

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type RoleFetcher struct {
	client *c8y.Client
	*IDNameFetcher
}

func NewRoleFetcher(client *c8y.Client) *RoleFetcher {
	return &RoleFetcher{
		client: client,
	}
}

func (f *RoleFetcher) getByID(id string) ([]fetcherResultSet, error) {
	if strings.Contains(id, "/roles/") {
		realID := id[strings.LastIndex(id, "/"):]
		return []fetcherResultSet{
			{
				ID:   realID,
				Name: realID,
				Self: id,
			},
		}, nil
	}

	role, resp, err := f.client.User.GetRole(
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

func (f *RoleFetcher) getByName(name string) ([]fetcherResultSet, error) {
	// check if already resolved, so we can save a lookup
	if strings.Contains(name, "/roles/") {
		id := name[strings.LastIndex(name, "/"):]
		return []fetcherResultSet{
			{
				ID:   id,
				Name: id,
				Self: name,
			},
		}, nil
	}
	roles, _, err := f.client.User.GetRoles(
		WithDisabledDryRunContext(f.client),
		&c8y.RoleOptions{
			PaginationOptions: *c8y.NewPaginationOptions(100),
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by name")
	}

	results := make([]fetcherResultSet, 0)

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
