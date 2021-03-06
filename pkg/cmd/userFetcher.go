package cmd

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type userFetcher struct {
	client *c8y.Client
}

func newUserFetcher(client *c8y.Client) *userFetcher {
	return &userFetcher{
		client: client,
	}
}

func (f *userFetcher) getByID(id string) ([]fetcherResultSet, error) {
	user, resp, err := client.User.GetUser(
		context.Background(),
		id,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 1)
	results[0] = fetcherResultSet{
		ID:    user.ID,
		Name:  user.Username,
		Value: *resp.JSON,
	}
	return results, nil
}

func (f *userFetcher) getByName(name string) ([]fetcherResultSet, error) {
	users, _, err := client.User.GetUsers(
		context.Background(),
		&c8y.UserOptions{
			Username:          strings.ReplaceAll(name, "*", ""),
			PaginationOptions: *c8y.NewPaginationOptions(5),
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, len(users.Users))

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
