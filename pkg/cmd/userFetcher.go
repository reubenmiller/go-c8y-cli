package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y-cli/pkg/matcher"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
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

// getFormattedUserSlice returns the user id and username
// returns raw strings, lookuped values, and errors
func getFormattedUserSlice(cmd *cobra.Command, args []string, name string) ([]string, []string, error) {
	f := newUserFetcher(client)

	if !cmd.Flags().Changed(name) {
		// TODO: Read from os.PIPE
		pipedInput, err := getPipe()
		if err != nil {
			Logger.Debug("No pipeline input detected")
		} else {
			Logger.Debugf("PIPED Input: %s\n", pipedInput)
			return nil, nil, nil
		}
	}

	values, err := cmd.Flags().GetStringSlice(name)
	if err != nil {
		Logger.Warning("Flag is missing", err)
	}

	values = ParseValues(append(values, args...))

	formattedValues, err := lookupEntity(f, values, true)

	if err != nil {
		Logger.Warningf("Failed to fetch entities. %s", err)
		return values, nil, err
	}

	results := []string{}

	invalidLookups := []string{}
	for _, item := range formattedValues {
		if item.ID != "" {
			if item.Name != "" {
				results = append(results, fmt.Sprintf("%s|%s", item.ID, item.Name))
			} else {
				results = append(results, item.ID)
			}
		} else {
			if item.Name != "" {
				invalidLookups = append(invalidLookups, item.Name)
			}
		}
	}

	var errors error

	if len(invalidLookups) > 0 {
		errors = fmt.Errorf("no results %v", invalidLookups)
	}

	return values, results, errors
}

// getFormattedUserLinkSlice returns the user id and username
// returns raw strings, lookuped values, and errors
func getFormattedUserLinkSlice(cmd *cobra.Command, args []string, name string) ([]string, []string, error) {
	f := newUserFetcher(client)

	if !cmd.Flags().Changed(name) {
		// TODO: Read from os.PIPE
		pipedInput, err := getPipe()
		if err != nil {
			Logger.Debug("No pipeline input detected")
		} else {
			Logger.Debugf("PIPED Input: %s\n", pipedInput)
			return nil, nil, nil
		}
	}

	values, err := cmd.Flags().GetStringSlice(name)
	if err != nil {
		Logger.Warning("Flag is missing", err)
	}

	values = ParseValues(append(values, args...))

	formattedValues, err := lookupEntity(f, values, true)

	if err != nil {
		Logger.Warningf("Failed to fetch entities. %s", err)
		return values, nil, err
	}

	results := []string{}

	invalidLookups := []string{}
	for _, item := range formattedValues {
		var selfLink string
		// Try to retrieve self link
		if data, ok := item.Data.Value.(gjson.Result); ok {
			if value := data.Get("self"); value.Exists() {
				selfLink = value.Str
			}
		}

		if selfLink != "" {
			if item.Name != "" {
				results = append(results, fmt.Sprintf("%s|%s", selfLink, item.Name))
			} else {
				results = append(results, selfLink)
			}
		} else {
			if item.Name != "" {
				invalidLookups = append(invalidLookups, item.Name)
			}
		}
	}

	var errors error

	if len(invalidLookups) > 0 {
		errors = fmt.Errorf("no results %v", invalidLookups)
	}

	return values, results, errors
}
