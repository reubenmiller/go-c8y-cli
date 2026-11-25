package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants"
	"github.com/spf13/cobra"
)

// WithTenantID tenant id completion
func WithTenantID(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			currentTenant := client.Auth.Tenant
			items, err := client.Tenants.List(
				context.Background(),
				tenants.ListOptions{PaginationOptions: pageSize(20)},
			).Collect()

			if err != nil {
				// Hide error, and just use the current tenant id
				values := make([]string, 0)
				if currentTenant != "" {
					values = append(values, currentTenant+"\tcurrent tenant")
				}
				return values, cobra.ShellCompDirectiveDefault
			}
			values := []string{currentTenant + "\tcurrent tenant"}
			for _, tenant := range items {
				details := []string{}
				if company := tenant.Company(); company != "" {
					details = append(details, fmt.Sprintf("company: %s", company))
				}
				if domain := tenant.Domain(); domain != "" {
					details = append(details, fmt.Sprintf("domain: %s", domain))
				}
				values = append(values, fmt.Sprintf("%s\t%s", tenant.ID(), strings.Join(details, " | ")))
			}
			return values, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}
