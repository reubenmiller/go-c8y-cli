package c8yfetcher

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type RoleFetcher struct {
	*CumulocityFetcher
	*IDNameFetcher
}

func NewRoleFetcher(factory *cmdutil.Factory) *RoleFetcher {
	return &RoleFetcher{
		CumulocityFetcher: &CumulocityFetcher{
			factory: factory,
		},
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

	role, resp, err := f.Client().User.GetRole(
		WithDisabledDryRunContext(f.Client()),
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
		Value: resp.JSON(),
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
	roles, _, err := f.Client().User.GetRoles(
		WithDisabledDryRunContext(f.Client()),
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
