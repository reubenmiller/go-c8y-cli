package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/notification2"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/spf13/cobra"
)

// notification2Completion completes notification2 subscriptions (one entry per
// subscription name), formatting each with format(subscription, description).
func notification2Completion(clientFunc ClientFunc, format func(item jsonmodels.Notification2Subscription, description string) string) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, err := clientFunc()
		if err != nil {
			return []string{err.Error()}, cobra.ShellCompDirectiveError
		}
		items, err := client.Notification2.List(
			context.Background(),
			notification2.ListOptions{
				PaginationOptions: pageSize(2000),
			},
		).Collect()

		if err != nil {
			values := []string{fmt.Sprintf("unknown. %s", err)}
			return values, cobra.ShellCompDirectiveError
		}
		values := []string{}
		pattern := "*" + toComplete + "*"
		uniqueNames := map[string]int64{}
		for _, item := range items {
			name := item.Subscription()
			if toComplete == "" || MatchString(pattern, name) || MatchString(pattern, item.ID()) {
				if v, ok := uniqueNames[name]; ok {
					uniqueNames[name] = v + 1
				} else {
					uniqueNames[name] = 1

					description := "context=" + item.Context()
					if typeFilter := item.SubscriptionFilterTypeFilter(); typeFilter != "" {
						description += " type=" + typeFilter
					}
					if apis := item.SubscriptionFilterApis(); len(apis) > 0 {
						description += " apis=" + strings.Join(apis, ",")
					}
					if fragments := item.FragmentsToCopy(); len(fragments) > 0 {
						description += " fragmentsToCopy=" + strings.Join(fragments, ",")
					}
					if sourceID := item.SourceID(); sourceID != "" {
						if sourceName := item.SourceName(); sourceName != "" {
							description += fmt.Sprintf(" source(id=%s,name=%s)", sourceID, sourceName)
						} else {
							description += fmt.Sprintf(" source(id=%s)", sourceID)
						}
					}
					values = append(values, format(item, description))
				}
			}
		}
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}

// WithNotification2SubscriptionName subscription name completion
func WithNotification2SubscriptionName(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, notification2Completion(clientFunc, func(item jsonmodels.Notification2Subscription, description string) string {
			return fmt.Sprintf("%s\t%s | id: %s", item.Subscription(), description, item.ID())
		}))
		return cmd
	}
}

// WithNotification2SubscriptionId subscription id completion
func WithNotification2SubscriptionId(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, notification2Completion(clientFunc, func(item jsonmodels.Notification2Subscription, description string) string {
			return fmt.Sprintf("%s\t%s | name: %s", item.ID(), description, item.Subscription())
		}))
		return cmd
	}
}
