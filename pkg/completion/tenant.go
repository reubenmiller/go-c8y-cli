package completion

import (
	"context"
	"fmt"
	"strings"

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
				// Hide error, and just use the current tenant id
				values := make([]string, 0)
				if client.TenantName != "" {
					values = append(values, client.TenantName+"\tcurrent tenant")
				}
				// values = append(values, fmt.Sprintf("error\t%s", err))
				return values, cobra.ShellCompDirectiveDefault
			}
			values := []string{client.TenantName + "\tcurrent tenant"}
			for _, tenant := range tenants.Tenants {
				details := []string{}
				if tenant.Company != "" {
					details = append(details, fmt.Sprintf("company: %s", tenant.Company))
				}
				if tenant.Domain != "" {
					details = append(details, fmt.Sprintf("domain: %s", tenant.Domain))
				}
				values = append(values, fmt.Sprintf("%s\t%s", tenant.ID, strings.Join(details, " | ")))
			}
			return values, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}
