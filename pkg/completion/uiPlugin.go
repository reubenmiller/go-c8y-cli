package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/applications"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/applications/versions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/spf13/cobra"
)

const applicationAvailabilityShared = "SHARED"

// listUIPlugins returns the hosted applications with versions (UI plugins) of
// the current tenant followed by the shared ones.
func listUIPlugins(client *apiv2.Client) (owned []jsonmodels.Application, shared []jsonmodels.Application, err error) {
	opt := applications.ListOptions{
		Type:        applicationTypeHosted,
		HasVersions: "true",
	}
	owned, err = listApplications(client, opt)
	if err != nil {
		return nil, nil, err
	}

	opt.Availability = applicationAvailabilityShared
	shared, err = listApplications(client, opt)
	if err != nil {
		return nil, nil, err
	}
	return owned, shared, nil
}

// WithUIPlugin UI extension completion
func WithUIPlugin(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			owned, shared, err := listUIPlugins(client)
			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range append(owned, shared...) {
				if toComplete == "" || MatchString(pattern, item.Name()) || MatchString(pattern, item.ID()) {
					values = append(values, formatApplication(item))
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

func formatUIPluginValues(items []jsonmodels.Application, toComplete string) []string {
	values := []string{}
	pattern := "*" + toComplete + "*"

	withVersions := strings.Contains(toComplete, "@")

	for _, app := range items {
		appVersions := app.Get("applicationVersions").Array()
		if withVersions {
			// Show name@version
			for _, appVersion := range appVersions {
				version := appVersion.Get("version").String()
				value := fmt.Sprintf("%s@%s", app.Name(), version)
				if toComplete == "" || MatchString(pattern, value) || MatchString(pattern, app.ID()) {
					tagValues := []string{}
					for _, tag := range appVersion.Get("tags").Array() {
						tagValues = append(tagValues, tag.String())
					}
					tags := strings.Join(tagValues, ",")
					if tags == "" {
						tags = "-"
					}
					values = append(values, fmt.Sprintf("%s\t%s | id: %s, version: %s, tags: [%s]", value, app.Type(), app.ID(), version, tags))
				}
			}
		} else {
			// Show just names
			if toComplete == "" || MatchString(pattern, app.Name()) || MatchString(pattern, app.ID()) {
				description := app.Get("manifest.description").String()
				if description == "" {
					description = "<no description>"
				}
				values = append(values, fmt.Sprintf("%s\t%s | id: %s, versions: %d, desc: %s", app.Name(), app.Type(), app.ID(), len(appVersions), description))
			}
		}
	}
	return values
}

// WithUIPluginWithVersions UI plugin with version completion
// Values are returned in the format of name@version
func WithUIPluginWithVersions(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			owned, shared, err := listUIPlugins(client)
			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := formatUIPluginValues(owned, toComplete)
			values = append(values, formatUIPluginValues(shared, toComplete)...)
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// WithUIPluginVersion UI extension version completion (requires the plugin)
func WithUIPluginVersion(flagVersion string, flagExtension string, clientFunc ClientFunc) Option {
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
					items, err := listApplications(client, applications.ListOptions{
						Name:              name,
						HasVersions:       "true",
						PaginationOptions: pageSize(100),
					})
					if err != nil {
						return values, cobra.ShellCompDirectiveNoFileComp
					}
					if len(items) > 0 {
						extensionID = items[0].ID()
					}
				}
			}

			if extensionID == "" {
				return values, cobra.ShellCompDirectiveNoFileComp
			}

			// Get versions
			items, err := client.Applications.Versions.List(context.Background(), extensionID, versions.ListOptions{
				PaginationOptions: pageSize(100),
			}).Collect()
			if err != nil {
				return values, cobra.ShellCompDirectiveNoFileComp
			}

			pattern := "*" + toComplete + "*"

			for _, item := range items {
				version, binaryID := item.Version(), item.BinaryID()
				if toComplete == "" || MatchString(pattern, version) || MatchString(pattern, binaryID) {
					values = append(values, fmt.Sprintf("%s\t%s | id: %s", version, strings.Join(item.Tags(), ","), binaryID))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
