package completion

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// WithDeviceCertificate trusted device certificate completion
func WithDeviceCertificate(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, _, err := client.DeviceCertificate.GetCertificates(
				c8y.WithDisabledDryRunContext(context.Background()),
				client.TenantName,
				&c8y.DeviceCertificateCollectionOptions{
					PaginationOptions: *c8y.NewPaginationOptions(100),
				},
			)

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items.Certificates {
				if toComplete == "" || MatchString(pattern, item.Name) || MatchString(pattern, item.Fingerprint) {
					values = append(values, fmt.Sprintf("%s\t%s | id: %s", item.Name, item.Issuer, item.Fingerprint))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}
