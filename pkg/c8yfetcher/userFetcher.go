package c8yfetcher

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type UserFetcher struct {
	*CumulocityFetcher
	*IDNameFetcher
}

func NewUserFetcher(factory *cmdutil.Factory) *UserFetcher {
	return &UserFetcher{
		CumulocityFetcher: &CumulocityFetcher{
			factory: factory,
		},
	}
}

func (f *UserFetcher) getByID(id string) ([]fetcherResultSet, error) {
	user, resp, err := f.Client().User.GetUser(
		WithDisabledDryRunContext(f.Client()),
		id,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 1)
	results[0] = fetcherResultSet{
		ID:    user.ID,
		Name:  user.Username,
		Value: resp.JSON(),
	}
	return results, nil
}

func (f *UserFetcher) getByName(name string) ([]fetcherResultSet, error) {
	users, _, err := f.Client().User.GetUsers(
		WithDisabledDryRunContext(f.Client()),
		&c8y.UserOptions{
			Username:          strings.ReplaceAll(name, "*", ""),
			PaginationOptions: *c8y.NewPaginationOptions(5),
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 0)

	for i, user := range users.Users {
		if isMatch, _ := matcher.MatchWithWildcards(user.ID, name); !isMatch {
			continue
		}
		results = append(results, fetcherResultSet{
			ID:    user.ID,
			Name:  user.Username,
			Self:  user.Self,
			Value: users.Items[i],
		})
	}

	return results, nil
}
