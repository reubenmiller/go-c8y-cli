package completion

import (
	"context"
	"fmt"

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
