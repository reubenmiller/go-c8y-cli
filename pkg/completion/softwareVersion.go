package completion

import (
	"context"
	"fmt"

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
			versionPattern := "*" + toComplete + "*"

			softwareNames, err := GetFlagStringValues(cmd, flagNameSoftware)
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			softwareName := ""
			if len(softwareNames) > 0 {
				softwareName = softwareNames[0]
			}

			items, _, err := client.Software.GetSoftwareVersionsByName(
				context.Background(),
				softwareName,
				versionPattern,
				true,
				c8y.NewPaginationOptions(100),
			)

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}

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
