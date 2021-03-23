package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithDeviceGroup device group completion
func WithDeviceGroup(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			pattern := "*" + toComplete + "*"
			opt := &c8y.ManagedObjectOptions{
				Query:             fmt.Sprintf("(name eq '%s') and has(%s)", pattern, "c8y_IsDeviceGroup"),
				PaginationOptions: *c8y.NewPaginationOptions(100),
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
			for _, item := range items.ManagedObjects {
				if toComplete == "" || MatchString(pattern, item.Name) || MatchString(pattern, item.ID) {
					values = append(values, fmt.Sprintf("%s\t%s (id=%s)", item.Name, item.Type, item.ID))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
