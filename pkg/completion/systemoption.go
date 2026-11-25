package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/systemoptions"
	"github.com/spf13/cobra"
)

// WithSystemOptionCategory system option category completion
func WithSystemOptionCategory(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			pattern := "*" + toComplete + "*"
			items, err := client.Tenants.SystemOptions.List(
				context.Background(),
				systemoptions.ListOptions{PaginationOptions: pageSize(200)},
			).Collect()

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			keys := make(map[string]interface{})
			for _, item := range items {
				category := item.Category()
				if toComplete == "" || MatchString(pattern, category) {
					if _, ok := keys[category]; !ok {
						values = append(values, category)
						keys[category] = struct{}{}
					}
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// WithSystemOptionKey system option key completion (requires category)
func WithSystemOptionKey(flagName string, flagNameCategory string, clientFunc ClientFunc) Option {
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
			items, err := client.Tenants.SystemOptions.List(
				context.Background(),
				systemoptions.ListOptions{PaginationOptions: pageSize(200)},
			).Collect()

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			keys := make(map[string]interface{})
			for _, item := range items {
				if item.Category() != category {
					continue
				}
				key := item.Key()
				if toComplete == "" || MatchString(pattern, key) {
					if _, ok := keys[key]; !ok {
						values = append(values, key)
						keys[key] = struct{}{}
					}
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
