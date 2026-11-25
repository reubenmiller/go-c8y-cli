package completion

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/spf13/cobra"
)

// WithFirmwareVersion firmware version completion (requires category)
func WithFirmwareVersion(flagVersion string, flagNameFirmware string, clientFunc ClientFunc) Option {
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
			opt := managedobjects.ListOptions{
				// Only filter by version name
				Query:             fmt.Sprintf("$filter=(type eq '%s') and (not(has(c8y_Patch))) and (c8y_Firmware.version eq '%s') $orderby=c8y_Firmware.version,creationTime", versionType, versionPattern),
				GetOptions:        managedobjects.GetOptions{WithParents: true},
				PaginationOptions: pageSize(100),
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
					packages, err := listManagedObjects(client, managedobjects.ListOptions{
						Query: fmt.Sprintf("$filter=(type eq '%s') and name eq '%s' $orderby=name,creationTime", "c8y_Firmware", firmwareName),
					})
					if err == nil && len(packages) > 0 {
						opt.Query = fmt.Sprintf(
							"$filter=(type eq '%s') and (not(has(c8y_Patch))) and (c8y_Firmware.version eq '%v') and (bygroupid(%s)) $orderby=c8y_Firmware.version,creationTime",
							versionType, versionPattern, packages[0].ID())
					}
				}
			}

			items, err := listManagedObjects(client, opt)

			if err != nil {
				values = append(values, fmt.Sprintf("error. %s", err))
				return values, cobra.ShellCompDirectiveError
			}

			for _, item := range items {
				version := item.Get("c8y_Firmware.version").String()
				if version == "" {
					continue
				}
				if toComplete == "" || MatchString(versionPattern, version) {
					versionParent := item.Get("additionParents.references.0.managedObject.name").String()
					values = append(values, fmt.Sprintf("%s\t%s | id: %s", version, versionParent, item.ID()))
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
