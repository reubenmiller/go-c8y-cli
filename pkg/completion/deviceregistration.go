package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/devices/registration"
	"github.com/spf13/cobra"
)

// WithDeviceRegistrationRequest device registration request completion
func WithDeviceRegistrationRequest(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			pattern := "*" + toComplete + "*"
			items, err := client.Devices.Registration.List(
				context.Background(),
				registration.ListOptions{
					PaginationOptions: pageSize(100),
				},
			).Collect()

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}

			for _, item := range items {
				if toComplete == "" || MatchString(pattern, item.ID()) {
					values = append(values, fmt.Sprintf("%s\t%s | creationTime: %s", item.ID(), item.Status(), item.Get("creationTime").String()))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
