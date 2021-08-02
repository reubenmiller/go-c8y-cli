package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithSoftwareVersion software version completion (requires category)
func WithSoftwareVersion(flagVersion string, flagNameSoftware string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagVersion, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			// query
			versionType := "c8y_SoftwareBinary"
			versionPattern := "*" + toComplete + "*"
			opt := &c8y.ManagedObjectOptions{
				// Only filter by version name
				Query:             fmt.Sprintf("(type eq '%s') and (not(has(c8y_Patch))) and (c8y_Software.version eq '%s')", versionType, versionPattern),
				WithParents:       true,
				PaginationOptions: *c8y.NewPaginationOptions(100),
			}

			// opt.Query = strings.ReplaceAll(opt.Query, " ", "%20")

			softwareNames, err := cmd.Flags().GetStringSlice(flagNameSoftware)
			softwareName := ""
			if len(softwareNames) > 0 {
				softwareName = softwareNames[0]
			}

			if err == nil && softwareName != "" {
				if c8ydata.IsID(softwareName) {
					opt.Query = fmt.Sprintf(
						"(type eq '%s') and (not(has(c8y_Patch))) and (c8y_Software.version eq '%v') and (bygroupid(%s))",
						versionType, versionPattern, softwareName,
					)
				} else {
					// Lookup by name
					softwarePackages, _, err := client.Inventory.GetManagedObjects(
						context.Background(),
						&c8y.ManagedObjectOptions{
							Query: fmt.Sprintf("(type eq '%s') and name eq '%s'", "c8y_Software", softwareName),
						},
					)
					if err == nil && len(softwarePackages.ManagedObjects) > 0 {
						opt.Query = fmt.Sprintf(
							"(type eq '%s') and (not(has(c8y_Patch))) and (c8y_Software.version eq '%v') and (bygroupid(%s))",
							versionType, versionPattern, softwarePackages.ManagedObjects[0].ID)
					}
				}
			}

			items, _, err := client.Inventory.GetManagedObjects(
				context.Background(),
				opt,
			)

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			// values = append(values, fmt.Sprintf("count: %v", len(items.ManagedObjects)))
			for i, item := range items.ManagedObjects {
				version := ""
				if v := items.Items[i].Get("c8y_Software.version"); v.Exists() {
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
