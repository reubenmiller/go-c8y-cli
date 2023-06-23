package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithUser user completion
func WithUser(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			items, _, err := client.User.GetUsers(
				c8y.WithDisabledDryRunContext(context.Background()),
				&c8y.UserOptions{
					PaginationOptions: *c8y.NewPaginationOptions(2000),
				},
			)

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items.Users {
				if toComplete == "" || MatchString(pattern, item.Username) || MatchString(pattern, item.FirstName) || MatchString(pattern, item.LastName) {
					details := []string{}
					if item.Email != "" {
						details = append(details, "email: "+item.Email)
					}
					if item.FirstName != "" || item.LastName != "" {
						details = append(details, fmt.Sprintf("name: %s %s", item.FirstName, item.LastName))
					}
					values = append(values, fmt.Sprintf("%s\t%s", item.ID, strings.Join(details, " | ")))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
