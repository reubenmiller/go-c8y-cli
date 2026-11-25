package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/userroles"
	"github.com/spf13/cobra"
)

// WithUserRole user role completion
func WithUserRole(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, err := client.UserRoles.List(
				context.Background(),
				userroles.ListOptions{
					PaginationOptions: pageSize(2000),
				},
			).Collect()

			if err != nil {
				values := []string{fmt.Sprintf("unknown. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items {
				if toComplete == "" || MatchString(pattern, item.Name()) || MatchString(pattern, item.ID()) {
					values = append(values, item.ID())
				}
			}
			return values, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}
