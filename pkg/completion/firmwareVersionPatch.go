package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithFirmwarePatch firmware patch version completion (requires category)
func WithFirmwarePatch(flagVersion string, flagNameFirmware string, clientFunc func() (*c8y.Client, error)) Option {
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
				Query:             fmt.Sprintf("(type eq '%s') and has(c8y_Patch) and (c8y_Firmware.version eq '%s')", versionType, versionPattern),
				WithParents:       true,
				PaginationOptions: *c8y.NewPaginationOptions(100),
			}

			firmwareNames, err := cmd.Flags().GetStringSlice(flagNameFirmware)
			firmwareName := ""
			if len(firmwareNames) > 0 {
				firmwareName = firmwareNames[0]
			}

			if err == nil && firmwareName != "" {
				// Filter firmware versions by firmware
				if c8ydata.IsID(firmwareName) {
					opt.Query = fmt.Sprintf(
						"(type eq '%s') and (not(has(c8y_Patch))) and (name eq '%v') and (bygroupid(%s))",
						versionType, versionPattern, firmwareName,
					)
				} else {
					// Lookup by name
					packages, _, err := client.Inventory.GetManagedObjects(
						context.Background(),
						&c8y.ManagedObjectOptions{
							Query: fmt.Sprintf("(type eq '%s') and name eq '%s'", "c8y_Firmware", firmwareName),
						},
					)
					if err == nil && len(packages.ManagedObjects) > 0 {
						opt.Query = fmt.Sprintf(
							"(type eq '%s') and (has(c8y_Patch)) and (c8y_Firmware.version eq '%v') and (bygroupid(%s))",
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
				dependency := ""
				if v := items.Items[i].Get("c8y_Firmware.version"); v.Exists() {
					version = v.String()
				}
				if v := items.Items[i].Get("c8y_Patch.dependency"); v.Exists() {
					dependency = v.String()
				}
				if version == "" {
					continue
				}
				if toComplete == "" || MatchString(versionPattern, version) {

					versionParent := ""
					if v := items.Items[i].Get("additionParents.references.0.managedObject.name"); v.Exists() {
						versionParent = v.String()
					}
					values = append(values, fmt.Sprintf("%s\t%s (dependency: %s) | id: %s", version, versionParent, dependency, item.ID))
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
