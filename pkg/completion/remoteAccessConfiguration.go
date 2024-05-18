package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithRemoteAccessConfiguration remote access configuration completion (requires device)
func WithRemoteAccessConfiguration(flagConfiguration string, flagDevice string, clientFunc func() (*c8y.Client, error)) Option {
	return findRemoteAccessConfigurations(flagConfiguration, flagDevice, "", clientFunc)
}

// WithRemoteAccessPassthroughConfiguration complete passthrough remote access completions (requires device)
func WithRemoteAccessPassthroughConfiguration(flagConfiguration string, flagDevice string, clientFunc func() (*c8y.Client, error)) Option {
	return findRemoteAccessConfigurations(flagConfiguration, flagDevice, c8y.RemoteAccessProtocolPassthrough, clientFunc)
}

func findRemoteAccessConfigurations(flagConfiguration string, flagDevice string, configType string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagConfiguration, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveError
			}

			deviceNames, err := GetFlagStringValues(cmd, flagDevice)
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveError
			}

			// stop if no device is given otherwise there will be too many service name conflicts
			if len(deviceNames) == 0 {
				return []string{fmt.Errorf("device is required. %w", ErrNotFound).Error()}, cobra.ShellCompDirectiveError
			}

			if deviceNames[0] == "" {
				return []string{fmt.Errorf("device is required. %w", ErrNotFound).Error()}, cobra.ShellCompDirectiveError
			}
			deviceName := deviceNames[0]

			var deviceID string
			if c8y.IsID(deviceName) {
				// no lookup required
				deviceID = deviceName
			} else {
				// lookup via name
				items, _, err := client.Inventory.GetDevicesByName(
					c8y.WithDisabledDryRunContext(context.Background()),
					deviceName,
					c8y.NewPaginationOptions(100),
				)
				if err != nil {
					return []string{err.Error()}, cobra.ShellCompDirectiveError
				}
				if len(items.ManagedObjects) == 0 {
					return []string{"DeviceNotFound"}, cobra.ShellCompDirectiveError
				}
				deviceID = items.ManagedObjects[0].ID
				// deviceName = items.ManagedObjects[0].Name
			}
			// technically we could save a request and just use the managed object data structure (.c8y_RemoteAccessList)
			// but it is more correct to get the list of connections from the service
			configs, _, err := client.RemoteAccess.GetConfigurations(
				c8y.WithDisabledDryRunContext(context.Background()),
				deviceID,
				&c8y.RemoteAccessCollectionOptions{
					PaginationOptions: *c8y.NewPaginationOptions(100),
				},
			)
			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}

			values := []string{}
			namePattern, err := matcher.ConvertWildcardToRegex("*" + toComplete + "*")
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveError
			}
			for _, item := range configs {
				if configType != "" && !strings.EqualFold(item.Protocol, configType) {
					continue
				}
				if namePattern.MatchString(item.Name) || strings.HasPrefix(item.ID, toComplete) {
					values = append(values, fmt.Sprintf("%s\t%s %s:%d | id: %s", item.Name, item.Protocol, item.Hostname, item.Port, item.ID))
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
