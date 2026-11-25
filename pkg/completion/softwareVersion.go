package completion

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/spf13/cobra"
)

// findSoftwareVersions returns the versions (with their parents) of a software package,
// referenced by id or name, whose version matches the given pattern.
func findSoftwareVersions(client *apiv2.Client, software string, versionPattern string) ([]jsonmodels.ManagedObject, error) {
	if !c8ydata.IsID(software) {
		// Lookup software via name
		packages, err := queryManagedObjects(client, softwareByNameQuery(software), 2)
		if err != nil {
			return nil, err
		}
		if len(packages) == 0 {
			return nil, ErrNotFound
		}
		software = packages[0].ID()
	}

	return listManagedObjects(client, managedobjects.ListOptions{
		Query:             fmt.Sprintf("$filter=(c8y_Software.version eq '%s') and bygroupid(%s) $orderby=c8y_Software.version,creationTime", versionPattern, software),
		GetOptions:        managedobjects.GetOptions{WithParents: true},
		PaginationOptions: pageSize(100),
	})
}

// WithSoftwareVersion software version completion (requires category)
func WithSoftwareVersion(flagVersion string, flagNameSoftware string, clientFunc ClientFunc) Option {
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

			items, err := findSoftwareVersions(client, softwareName, versionPattern)

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}

			for _, item := range items {
				version := item.Get("c8y_Software.version").String()
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
