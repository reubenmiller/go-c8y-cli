package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithFeature feature key completion
func WithFeature(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			features, _, err := client.Features.GetFeatures(
				c8y.WithDisabledDryRunContext(context.Background()),
			)

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}

			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range features {
				if toComplete == "" || MatchString(pattern, item.Key) {
					details := []string{}
					if item.Phase != "" {
						details = append(details, "phase: "+fmt.Sprintf("% -15s", item.Phase))
					}

					if item.Strategy != "" {
						details = append(details, "strategy: "+fmt.Sprintf("% -7s", item.Strategy))
					}
					// include so the completions descriptions are unique, otherwise
					// some shells like zsh try to group them
					details = append(details, "key: "+item.Key)
					values = append(values, fmt.Sprintf("%s\t%s", item.Key, strings.Join(details, " | ")))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
