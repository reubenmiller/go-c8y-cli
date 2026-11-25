package completion

import (
	"context"
	"errors"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/spf13/cobra"
)

var ErrNotFound = errors.New("not found")

func getSupportedSeries(cmd *cobra.Command, flagNameDevice string, client *apiv2.Client, toComplete string, formatFunc func(string) string) ([]string, error) {
	deviceID, err := cmd.Flags().GetString(flagNameDevice)
	if err != nil {
		// try getting a slice (only the first one is supported)
		if v, err := cmd.Flags().GetStringSlice(flagNameDevice); err == nil && len(v) > 0 {
			deviceID = v[0]
		}
	}

	if deviceID == "" {
		return nil, nil
	}

	if !c8ydata.IsID(deviceID) {
		matchingDevices, err := findDevicesByName(client, deviceID, 1)
		if err != nil {
			return nil, err
		}
		if len(matchingDevices) == 0 {
			return nil, ErrNotFound
		}
		deviceID = matchingDevices[0].ID()
	}

	pattern := "*" + toComplete + "*"
	series, err := client.ManagedObjects.ListSupportedSeries(context.Background(), deviceID).Unwrap()
	if err != nil {
		return nil, err
	}
	values := []string{}
	for _, item := range series.Values() {
		if toComplete == "" || MatchString(pattern, item) {
			fValue := formatFunc(item)
			if fValue != "" {
				values = append(values, fValue)
			}
		}
	}

	return values, nil
}

// WithDeviceMeasurementSeries supported measurement series completion (requires device)
func WithDeviceMeasurementSeries(flagName string, flagNameDevice string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			values, err := getSupportedSeries(cmd, flagNameDevice, client, toComplete, func(v string) string { return v })
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			return values, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}

// WithDeviceMeasurementValueFragmentType supported measurement value fragment types completion (requires device)
func WithDeviceMeasurementValueFragmentType(flagName string, flagNameDevice string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			values, err := getSupportedSeries(cmd, flagNameDevice, client, toComplete, func(v string) string {
				parts := strings.SplitN(v, ".", 2)
				if len(parts) > 0 {
					return parts[0]
				}
				return ""
			})
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			return values, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}

// WithDeviceMeasurementValueFragmentSeries supported measurement value fragment types completion (requires device)
func WithDeviceMeasurementValueFragmentSeries(flagName string, flagNameDevice string, flagNameValueFragmentType string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			valueFragmentType := ""
			if v, err := cmd.Flags().GetString(flagNameValueFragmentType); err == nil {
				valueFragmentType = v
			}

			values, err := getSupportedSeries(cmd, flagNameDevice, client, valueFragmentType+"."+toComplete, func(v string) string {
				parts := strings.SplitN(v, ".", 2)
				if len(parts) >= 2 {
					return parts[1]
				}
				return ""
			})
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			return values, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}
