package completion

import (
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithTenantOptionCategory tenant option category completion
func WithTenantOptionCategory(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			pattern := "*" + toComplete + "*"
			items, _, err := client.TenantOptions.GetOptions(
				WithDisabledDryRunContext(client),
				c8y.NewPaginationOptions(200),
			)

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			keys := make(map[string]interface{})
			for _, item := range items.Options {
				if toComplete == "" || MatchString(pattern, item.Category) {
					if _, ok := keys[item.Category]; !ok {
						values = append(values, item.Category)
						keys[item.Category] = struct{}{}
					}
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// WithTenantOptionKey tenant option key completion (requires category)
func WithTenantOptionKey(flagName string, flagNameCategory string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			category, err := cmd.Flags().GetString(flagNameCategory)
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			pattern := "*" + toComplete + "*"
			items, _, err := client.TenantOptions.GetOptions(
				WithDisabledDryRunContext(client),
				c8y.NewPaginationOptions(200),
			)

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			keys := make(map[string]interface{})
			for _, item := range items.Options {
				if item.Category != category {
					continue
				}
				if toComplete == "" || MatchString(pattern, item.Key) {
					if _, ok := keys[item.Key]; !ok {
						values = append(values, item.Key)
						keys[item.Key] = struct{}{}
					}
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
