package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithNotification2SubscriptionName subscription name completion
func WithNotification2SubscriptionName(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveError
			}
			items, _, err := client.Notification2.GetSubscriptions(
				c8y.WithDisabledDryRunContext(context.Background()),
				&c8y.Notification2SubscriptionCollectionOptions{
					PaginationOptions: *c8y.NewPaginationOptions(2000),
				},
			)

			if err != nil {
				values := []string{fmt.Sprintf("unknown. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			uniqueNames := map[string]int64{}
			for _, item := range items.Subscriptions {
				if toComplete == "" || MatchString(pattern, item.Subscription) || MatchString(pattern, item.ID) {
					if v, ok := uniqueNames[item.Subscription]; ok {
						uniqueNames[item.Subscription] = v + 1
					} else {
						uniqueNames[item.Subscription] = 1

						description := "context=" + item.Context
						if item.SubscriptionFilter.TypeFilter != "" {
							description += " type=" + item.SubscriptionFilter.TypeFilter
						}
						if len(item.SubscriptionFilter.Apis) > 0 {
							description += " apis=" + strings.Join(item.SubscriptionFilter.Apis, ",")
						}
						if len(item.FragmentsToCopy) > 0 {
							description += " fragmentsToCopy=" + strings.Join(item.SubscriptionFilter.Apis, ",")
						}
						if item.Source.ID != "" {
							if item.Source.Name != "" {
								description += fmt.Sprintf(" source(id=%s,name=%s)", item.Source.ID, item.Source.Name)
							} else {
								description += fmt.Sprintf(" source(id=%s)", item.Source.ID)
							}
						}
						values = append(values, fmt.Sprintf("%s\t%s | id: %s", item.Subscription, description, item.ID))
					}
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

func WithNotification2SubscriptionId(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveError
			}
			items, _, err := client.Notification2.GetSubscriptions(
				c8y.WithDisabledDryRunContext(context.Background()),
				&c8y.Notification2SubscriptionCollectionOptions{
					PaginationOptions: *c8y.NewPaginationOptions(2000),
				},
			)

			if err != nil {
				values := []string{fmt.Sprintf("unknown. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			uniqueNames := map[string]int64{}
			for _, item := range items.Subscriptions {
				if toComplete == "" || MatchString(pattern, item.Subscription) || MatchString(pattern, item.ID) {
					if v, ok := uniqueNames[item.Subscription]; ok {
						uniqueNames[item.Subscription] = v + 1
					} else {
						uniqueNames[item.Subscription] = 1

						description := "context=" + item.Context
						if item.SubscriptionFilter.TypeFilter != "" {
							description += " type=" + item.SubscriptionFilter.TypeFilter
						}
						if len(item.SubscriptionFilter.Apis) > 0 {
							description += " apis=" + strings.Join(item.SubscriptionFilter.Apis, ",")
						}
						if len(item.FragmentsToCopy) > 0 {
							description += " fragmentsToCopy=" + strings.Join(item.SubscriptionFilter.Apis, ",")
						}
						if item.Source.ID != "" {
							if item.Source.Name != "" {
								description += fmt.Sprintf(" source(id=%s,name=%s)", item.Source.ID, item.Source.Name)
							} else {
								description += fmt.Sprintf(" source(id=%s)", item.Source.ID)
							}
						}
						values = append(values, fmt.Sprintf("%s\t%s | name: %s", item.ID, description, item.Subscription))
					}
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
