package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithUIExtension UI extension completion
func WithUIExtension(flagName string, clientFunc func() (*c8y.Client, error)) Option {
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
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// Complete UI extension versions
func WithUIExtensionVersion(flagVersion string, flagExtension string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagVersion, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			values := []string{}

			// Lookup firmware by name
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

			// Get firmware versions
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
