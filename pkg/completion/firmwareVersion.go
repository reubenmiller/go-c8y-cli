package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithFirmwareVersion firmware version completion (requires category)
func WithFirmwareVersion(flagVersion string, flagNameFirmware string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagVersion, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			values := []string{}

			// query
			versionType := "c8y_FirmwareBinary"
			versionPattern := "*" + toComplete + "*"
			opt := &c8y.ManagedObjectOptions{
				// Only filter by version name
				Query:             fmt.Sprintf("$filter=(type eq '%s') and (not(has(c8y_Patch))) and (c8y_Firmware.version eq '%s') $orderby=c8y_Firmware.version,creationTime", versionType, versionPattern),
				WithParents:       true,
				PaginationOptions: *c8y.NewPaginationOptions(100),
			}

			firmwareNames, err := GetFlagStringValues(cmd, flagNameFirmware)
			firmwareName := ""
			if len(firmwareNames) > 0 {
				firmwareName = firmwareNames[0]
			}

			if err == nil && firmwareName != "" {
				// Filter firmware versions by firmware
				if c8ydata.IsID(firmwareName) {
					opt.Query = fmt.Sprintf(
						"$filter=(type eq '%s') and (not(has(c8y_Patch))) and (c8y_Firmware.version eq '%v') and (bygroupid(%s)) $orderby=c8y_Firmware.version,creationTime",
						versionType, versionPattern, firmwareName,
					)
				} else {
					// Lookup by name
					packages, _, err := client.Inventory.GetManagedObjects(
						context.Background(),
						&c8y.ManagedObjectOptions{
							Query: fmt.Sprintf("$filter=(type eq '%s') and name eq '%s' $orderby=name,creationTime", "c8y_Firmware", firmwareName),
						},
					)
					if err == nil && len(packages.ManagedObjects) > 0 {
						opt.Query = fmt.Sprintf(
							"$filter=(type eq '%s') and (not(has(c8y_Patch))) and (c8y_Firmware.version eq '%v') and (bygroupid(%s)) $orderby=c8y_Firmware.version,creationTime",
							versionType, versionPattern, packages.ManagedObjects[0].ID)
					}
				}
			}

			items, _, err := client.Inventory.GetManagedObjects(
				context.Background(),
				opt,
			)

			if err != nil {
				values = append(values, fmt.Sprintf("error. %s", err))
				return values, cobra.ShellCompDirectiveError
			}

			for i, item := range items.ManagedObjects {
				version := ""
				if v := items.Items[i].Get("c8y_Firmware.version"); v.Exists() {
					version = v.String()
				}
				if version == "" {
					continue
				}
				if toComplete == "" || MatchString(versionPattern, version) {

					versionParent := ""
					if v := items.Items[i].Get("additionParents.references.0.managedObject.name"); v.Exists() {
						versionParent = v.String()
					}
					values = append(values, fmt.Sprintf("%s\t%s | id: %s", version, versionParent, item.ID))
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
