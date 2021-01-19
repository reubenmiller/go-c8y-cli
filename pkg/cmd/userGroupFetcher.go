package cmd

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
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

// getFormattedGroupSlice returns the user group id and name
// returns raw strings, lookuped values, and errors
func getFormattedGroupSlice(cmd *cobra.Command, args []string, name string) ([]string, []string, error) {
	f := newUserGroupFetcher(client)

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
		Logger.Debug("Flag is missing", err)
	}

	values = ParseValues(append(values, args...))

	formattedValues, err := lookupEntity(f, values, false)

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
