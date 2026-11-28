package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/trustedcertificates"
	"github.com/spf13/cobra"
)

// WithDeviceCertificate trusted device certificate completion
func WithDeviceCertificate(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, err := client.TrustedCertificates.List(
				context.Background(),
				trustedcertificates.ListOptions{
					TenantID:          client.Auth().Tenant,
					PaginationOptions: pageSize(100),
				},
			).Collect()

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items {
				if toComplete == "" || MatchString(pattern, item.Name()) || MatchString(pattern, item.Fingerprint()) {
					values = append(values, fmt.Sprintf("%s\t%s | id: %s", item.Name(), item.Issuer(), item.Fingerprint()))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
