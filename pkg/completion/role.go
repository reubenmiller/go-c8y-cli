package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithUserRole user role completion
func WithUserRole(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, _, err := client.User.GetRoles(
				context.Background(),
				&c8y.RoleOptions{
					PaginationOptions: *c8y.NewPaginationOptions(2000),
				},
			)

			if err != nil {
				values := []string{fmt.Sprintf("unknown. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items.Roles {
				if toComplete == "" || MatchString(pattern, item.Name) || MatchString(pattern, item.ID) {
					values = append(values, item.ID)
				}
			}
			return values, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}
