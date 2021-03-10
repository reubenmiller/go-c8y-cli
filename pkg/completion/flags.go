package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// Option adds flags to a given command
type Option func(*cobra.Command) *cobra.Command

// WithOptions applies given options to the command
func WithOptions(cmd *cobra.Command, opts ...Option) *cobra.Command {
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

// WithValidateSet adds a completion function with the given values
func WithValidateSet(flagName string, values ...string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return values, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}

// WithLazyRequired marks a flag as required but does not enforce it.
func WithLazyRequired(flagName string, values ...string) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.Flags().SetAnnotation(flagName, cobra.BashCompOneRequiredFlag, []string{"false"})
		return cmd
	}
}

// WithTenantID tenant id completion
func WithTenantID(flagName string, client *c8y.Client) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			tenants, _, err := client.Tenant.GetTenants(
				context.Background(),
				c8y.NewPaginationOptions(20),
			)

			if err != nil {
				values := []string{fmt.Sprintf("unknown. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{"default"}
			for _, tenant := range tenants.Tenants {
				values = append(values, tenant.ID)
			}
			return values, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}

type CobraCompletion func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

func WithTenantCompletion(flagName string, client *c8y.Client) CobraCompletion {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// if len(args) != 0 {
		// 	return []string{"no args"}, cobra.ShellCompDirectiveNoFileComp
		// }

		// inputValue, err := cmd.Flags().GetString(flagName)

		tenants, _, err := client.Tenant.GetTenants(
			context.Background(),
			c8y.NewPaginationOptions(20),
		)

		if err != nil {
			values := []string{fmt.Sprintf("unknown. %s", err)}
			return values, cobra.ShellCompDirectiveNoFileComp
		}
		values := []string{"default"}
		for _, tenant := range tenants.Tenants {
			values = append(values, tenant.ID)
		}
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}
