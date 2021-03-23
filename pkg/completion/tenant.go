package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithTenantID tenant id completion
func WithTenantID(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			tenants, _, err := client.Tenant.GetTenants(
				context.Background(),
				c8y.NewPaginationOptions(20),
			)

			if err != nil {
				values := []string{fmt.Sprintf("unknown. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{client.TenantName}
			for _, tenant := range tenants.Tenants {
				values = append(values, tenant.ID)
			}
			return values, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}
