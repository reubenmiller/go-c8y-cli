package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithUIPlugin UI extension completion
func WithUIPlugin(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			hasVersions := true
			items, _, err := client.Application.GetApplications(
				c8y.WithDisabledDryRunContext(context.Background()),
				&c8y.ApplicationOptions{
					Type:              c8y.ApplicationTypeHosted,
					HasVersions:       &hasVersions,
					PaginationOptions: *c8y.NewPaginationOptions(2000),
				},
			)

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items.Applications {
				if toComplete == "" || MatchString(pattern, item.Name) || MatchString(pattern, item.ID) {
					values = append(values, fmt.Sprintf("%s\t%s | id: %s", item.Name, item.Type, item.ID))
				}
			}

			// Lookup shared extensions
			// Shared extensions
			items, _, err = client.Application.GetApplications(
				c8y.WithDisabledDryRunContext(context.Background()),
				&c8y.ApplicationOptions{
					Type:              c8y.ApplicationTypeHosted,
					HasVersions:       &hasVersions,
					Availability:      c8y.ApplicationAvailabilityShared,
					PaginationOptions: *c8y.NewPaginationOptions(2000),
				},
			)
			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			for _, item := range items.Applications {
				if toComplete == "" || MatchString(pattern, item.Name) || MatchString(pattern, item.ID) {
					values = append(values, fmt.Sprintf("%s\t%s | id: %s", item.Name, item.Type, item.ID))
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

func formatUIPluginValues(items *c8y.ApplicationCollection, toComplete string) []string {
	values := []string{}
	pattern := "*" + toComplete + "*"

	withVersions := strings.Contains(toComplete, "@")

	for i, app := range items.Applications {

		if withVersions {
			// Show name@version
			for _, appVersion := range app.ApplicationVersions {
				value := fmt.Sprintf("%s@%s", app.Name, appVersion.Version)
				if toComplete == "" || MatchString(pattern, value) || MatchString(pattern, app.ID) {
					tags := strings.Join(appVersion.Tags, ",")
					if tags == "" {
						tags = "-"
					}
					values = append(values, fmt.Sprintf("%s\t%s | id: %s, version: %s, tags: [%s]", value, app.Type, app.ID, appVersion.Version, tags))
				}
			}
		} else {
			// Show just names
			if toComplete == "" || MatchString(pattern, app.Name) || MatchString(pattern, app.ID) {
				// tags := strings.Join(appVersion.Tags, ",")
				// if tags == "" {
				// 	tags = "-"
				// }
				description := items.Items[i].Get("manifest.description").String()
				if description == "" {
					description = "<no description>"
				}
				totalVersions := len(app.ApplicationVersions)
				values = append(values, fmt.Sprintf("%s\t%s | id: %s, versions: %d, desc: %s", app.Name, app.Type, app.ID, totalVersions, description))
			}
		}
	}
	return values
}

// WithUIPluginWithVersions UI plugin with version completion
// Values are returned in the format of name@version
func WithUIPluginWithVersions(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			values := []string{}

			hasVersions := true

			// Extensions in current tenant
			items, _, err := client.Application.GetApplications(
				c8y.WithDisabledDryRunContext(context.Background()),
				&c8y.ApplicationOptions{
					Type:              c8y.ApplicationTypeHosted,
					HasVersions:       &hasVersions,
					PaginationOptions: *c8y.NewPaginationOptions(2000),
				},
			)
			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values = append(values, formatUIPluginValues(items, toComplete)...)

			// Shared extensions
			items, _, err = client.Application.GetApplications(
				c8y.WithDisabledDryRunContext(context.Background()),
				&c8y.ApplicationOptions{
					Type:              c8y.ApplicationTypeHosted,
					HasVersions:       &hasVersions,
					Availability:      c8y.ApplicationAvailabilityShared,
					PaginationOptions: *c8y.NewPaginationOptions(2000),
				},
			)
			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values = append(values, formatUIPluginValues(items, toComplete)...)
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// Complete UI extension versions
func WithUIPluginVersion(flagVersion string, flagExtension string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagVersion, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			values := []string{}

			// Lookup by name
			names, err := GetFlagStringValues(cmd, flagExtension)
			name := ""
			if len(names) > 0 {
				name = names[0]
			}

			extensionID := ""
			if err == nil && name != "" {
				if c8ydata.IsID(name) {
					extensionID = name
				} else {
					// Lookup by name
					col, _, err := client.UIExtension.GetExtensions(context.Background(), &c8y.ExtensionOptions{
						Name:              name,
						HasVersions:       true,
						PaginationOptions: *c8y.NewPaginationOptions(100),
					})
					if err != nil {
						return values, cobra.ShellCompDirectiveNoFileComp
					}
					if len(col.Applications) > 0 {
						extensionID = col.Applications[0].ID
					}
				}
			}

			if extensionID == "" {
				return values, cobra.ShellCompDirectiveNoFileComp
			}

			// Get versions
			col, _, err := client.ApplicationVersions.GetVersions(context.Background(), extensionID, &c8y.ApplicationVersionsOptions{
				PaginationOptions: *c8y.NewPaginationOptions(100),
			})
			if err != nil {
				return values, cobra.ShellCompDirectiveNoFileComp
			}

			pattern := "*" + toComplete + "*"

			for _, item := range col.Versions {
				if toComplete == "" || MatchString(pattern, item.Version) || MatchString(pattern, item.BinaryID) {
					values = append(values, fmt.Sprintf("%s\t%s | id: %s", item.Version, strings.Join(item.Tags, ","), item.BinaryID))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
