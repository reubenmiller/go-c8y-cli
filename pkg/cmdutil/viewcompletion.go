package cmdutil

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/spf13/cobra"
)

// WithViewCompletion view completion
func WithViewCompletion(flagName string, dataviewFunc func() (*dataview.DataView, error)) completion.Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			view, err := dataviewFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			pattern := "*" + toComplete + "*"
			items, err := view.GetViews(pattern)

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}

			values := []string{
				config.ViewsOff + "\tDisable view. Display all properties",
				config.ViewsAuto + "\tAuto detect view",
			}
			for _, item := range items {
				values = append(values, fmt.Sprintf("%s\t%v | file: %s", item.Name, item.Columns, item.FileName))
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
