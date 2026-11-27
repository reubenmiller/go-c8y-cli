package c8ystream

import (
	"context"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/matcher"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/users"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
)

// FirstUserMatch returns the first user whose name matches the pattern. Like
// v2.59.0 it asks the server for one page of 5 users, filtered by the pattern
// without its wildcards. The error of a failed request is a core.LookupError.
func FirstUserMatch(ctx context.Context, client *apiv2.Client, pattern string) (jsonmodels.User, bool, error) {
	res := client.Users.List(ctx, users.ListOptions{
		Username:          strings.ReplaceAll(pattern, "*", ""),
		PaginationOptions: pagination.PaginationOptions{PageSize: 5},
	})
	if res.Err != nil {
		return jsonmodels.User{}, false, core.WrapLookup("user", pattern, res.Err)
	}
	for user, err := range res.Items() {
		if err != nil {
			return jsonmodels.User{}, false, core.WrapLookup("user", pattern, err)
		}
		if matched, _ := matcher.MatchWithWildcards(user.ID(), pattern); matched {
			return user, true, nil
		}
	}
	return jsonmodels.User{}, false, nil
}

// UserWithWildcards returns the user name to use for a user given as a flag
// value. As in v2.59.0 a name with a wildcard ("peter*") is looked up and
// resolves to the first user which matches (no match is an unresolved name,
// see UnresolvedNameError); any other value is returned as is.
func UserWithWildcards(ctx context.Context, client *apiv2.Client, value string) (string, error) {
	if !strings.Contains(value, "*") {
		return value, nil
	}
	user, found, err := FirstUserMatch(ctx, client, value)
	if err != nil {
		return "", err
	}
	if !found {
		return "", UnresolvedNameError(value)
	}
	return user.ID(), nil
}
