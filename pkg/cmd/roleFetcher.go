package cmd

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y-cli/pkg/matcher"
	"github.com/spf13/cobra"
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

// getFormattedRoleSlice returns the user id and username
// returns raw strings, lookuped values, and errors
func getFormattedRoleSlice(cmd *cobra.Command, args []string, name string) ([]string, []string, error) {
	f := newRoleFetcher(client)

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

// getFormattedRoleSelfSlice returns the user id and username
// returns raw strings, lookuped values, and errors
func getFormattedRoleSelfSlice(cmd *cobra.Command, args []string, name string) ([]string, []string, error) {
	f := newRoleFetcher(client)

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
		Logger.Error("Flag is missing", err)
	}

	values = ParseValues(append(values, args...))

	formattedValues, err := lookupEntity(f, values, true)

	if err != nil {
		Logger.Errorf("Failed to fetch entities. %s", err)
		return values, nil, err
	}

	results := []string{}

	invalidLookups := []string{}
	for _, item := range formattedValues {
		selfLink := item.Data.Self

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
