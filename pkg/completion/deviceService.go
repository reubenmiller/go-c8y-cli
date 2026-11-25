package completion

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/spf13/cobra"
)

// WithDeviceService device service completion (requires device)
func WithDeviceService(flagService string, flagDevice string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagService, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveError
			}

			// query
			serviceNamePattern := "*" + toComplete + "*"

			deviceNames, err := GetFlagStringValues(cmd, flagDevice)
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveError
			}

			// stop if no device is given otherwise there will be too many service name conflicts
			if len(deviceNames) == 0 {
				return []string{fmt.Errorf("device is required. %w", ErrNotFound).Error()}, cobra.ShellCompDirectiveError
			}

			deviceName := ""
			if len(deviceNames) > 0 {
				deviceName = deviceNames[0]
			}

			var deviceID string
			if deviceName == "" {
				// do nothing
				deviceID = ""
			} else if c8ydata.IsID(deviceName) {
				// no lookup required
				deviceID = deviceName
			} else {
				// lookup via name
				devices, err := findDevicesByName(client, deviceName, 100)
				if err != nil {
					return []string{err.Error()}, cobra.ShellCompDirectiveError
				}
				if len(devices) == 0 {
					return []string{"DeviceNotFound"}, cobra.ShellCompDirectiveError
				}
				deviceID = devices[0].ID()
				deviceName = devices[0].Name()
			}

			query := fmt.Sprintf("type eq 'c8y_Service' and name eq '%s' and bygroupid(%s)", serviceNamePattern, deviceID)

			items, err := listManagedObjects(client, managedobjects.ListOptions{
				Query:             query,
				GetOptions:        managedobjects.GetOptions{WithParents: deviceName == ""},
				PaginationOptions: pageSize(100),
			})

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}

			serviceParent := deviceName

			for _, item := range items {
				if v := item.Get("additionParents.references.0.managedObject.name"); v.Exists() {
					serviceParent = v.String()
				}
				values = append(values, fmt.Sprintf("%s\t%s | id: %s", item.Name(), serviceParent, item.ID()))
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
