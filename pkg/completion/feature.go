package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// WithFeature feature key completion
func WithFeature(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			features, err := client.Features.List(context.Background()).Collect()

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}

			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range features {
				key := item.Key()
				if toComplete == "" || MatchString(pattern, key) {
					details := []string{}
					if phase := item.Phase(); phase != "" {
						details = append(details, "phase: "+fmt.Sprintf("% -15s", phase))
					}

					if strategy := item.Strategy(); strategy != "" {
						details = append(details, "strategy: "+fmt.Sprintf("% -7s", strategy))
					}
					// include so the completions descriptions are unique, otherwise
					// some shells like zsh try to group them
					details = append(details, "key: "+key)
					values = append(values, fmt.Sprintf("%s\t%s", key, strings.Join(details, " | ")))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
