package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/users"
	"github.com/spf13/cobra"
)

// WithUser user completion
func WithUser(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			items, err := client.Users.List(
				context.Background(),
				users.ListOptions{
					PaginationOptions: pageSize(2000),
				},
			).Collect()

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items {
				firstName, lastName := item.FirstName(), item.LastName()
				if toComplete == "" || MatchString(pattern, item.UserName()) || MatchString(pattern, firstName) || MatchString(pattern, lastName) {
					details := []string{}
					if email := item.Email(); email != "" {
						details = append(details, "email: "+email)
					}
					if firstName != "" || lastName != "" {
						details = append(details, fmt.Sprintf("name: %s %s", firstName, lastName))
					}
					values = append(values, fmt.Sprintf("%s\t%s", item.ID(), strings.Join(details, " | ")))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
