package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// WithApplication application completion
func WithApplication(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, _, err := client.Application.GetApplications(
				context.Background(),
				&c8y.ApplicationOptions{
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

// WithApplicationContext application context completion
func WithApplicationContext(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, _, err := client.Application.GetApplications(
				context.Background(),
				&c8y.ApplicationOptions{
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
				if toComplete == "" || MatchString(pattern, item.ContextPath) || MatchString(pattern, item.Name) || MatchString(pattern, item.ID) {
					values = append(values, fmt.Sprintf("%s\t%s | type: %s  id: %s", item.ContextPath, item.Name, item.Type, item.ID))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// WithHostedApplication hosted application completion
func WithHostedApplication(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, _, err := client.Application.GetApplications(
				context.Background(),
				&c8y.ApplicationOptions{
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
				if !strings.EqualFold(item.Type, "HOSTED") {
					continue
				}

				// Ignore if not hosted in the current application
				if client.TenantName != "" {
					if item.Owner != nil && item.Owner.Tenant != nil && item.Owner.Tenant.ID != "" {
						if item.Owner.Tenant.ID != client.TenantName {
							continue
						}
					}
				}

				if toComplete == "" || MatchString(pattern, item.Name) || MatchString(pattern, item.ID) {
					values = append(values, fmt.Sprintf("%s\t%s | id: %s", item.Name, item.Type, item.ID))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// WithMicroservice microservice completion
func WithMicroservice(flagName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, _, err := client.Application.GetApplications(
				context.Background(),
				&c8y.ApplicationOptions{
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
				if !strings.EqualFold(item.Type, "MICROSERVICE") {
					continue
				}
				if toComplete == "" || MatchString(pattern, item.Name) || MatchString(pattern, item.ID) {
					values = append(values, fmt.Sprintf("%s\t%s | id: %s", item.Name, item.Type, item.ID))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// WithMicroservice completion
func WithMicroserviceLoggers(flagName string, flagNameMicroserviceName string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			microserviceName, err := cmd.Flags().GetString(flagNameMicroserviceName)
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			values := []string{}
			resp, err := client.SendRequest(context.Background(), c8y.RequestOptions{
				Method: "GET",
				Path:   "/service/" + microserviceName + "/loggers",
			})

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}

			pattern := "*" + toComplete + "*"

			resp.JSON("loggers").ForEach(func(key, value gjson.Result) bool {
				if toComplete == "" || MatchString(pattern, key.String()) {
					values = append(values, key.String())
				}
				return true
			})

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

func getMicroserviceByName(client *c8y.Client, name string) (string, error) {
	apps, _, err := client.Application.GetApplicationsByName(
		context.Background(),
		name,
		&c8y.ApplicationOptions{
			Type:              c8y.ApplicationTypeMicroservice,
			PaginationOptions: *c8y.NewPaginationOptions(1),
		},
	)
	if err != nil {
		return "", err
	}

	for _, app := range apps.Applications {
		return app.ID, nil
	}
	return "", ErrNotFound
}

func getMicroserviceInstances(cmd *cobra.Command, flagApplicationID string, client *c8y.Client, toComplete string, formatFunc func(string) string) ([]string, error) {
	deviceID, err := cmd.Flags().GetString(flagApplicationID)
	if err != nil {
		// try getting a slice (only the first one is supported)
		if v, err := cmd.Flags().GetStringSlice(flagApplicationID); err == nil && len(v) > 0 {
			deviceID = v[0]
		}
	}

	if deviceID == "" {
		return nil, nil
	}

	if !c8ydata.IsID(deviceID) {
		deviceID, err = getMicroserviceByName(client, deviceID)

		if err != nil {
			return nil, err
		}

		if deviceID == "" {
			return nil, ErrNotFound
		}
	}

	pattern := "*" + toComplete + "*"
	items, _, err := client.Inventory.GetManagedObjects(
		context.Background(),
		&c8y.ManagedObjectOptions{
			Type:              "c8y_Application_" + deviceID,
			PaginationOptions: *c8y.NewPaginationOptions(1),
		},
	)

	if err != nil {
		return nil, err
	}
	values := []string{}
	for i := range items.ManagedObjects {
		if v := items.Items[i].Get("c8y_Status.instances"); v.IsObject() {
			v.ForEach(func(key, value gjson.Result) bool {
				if toComplete == "" || MatchString(pattern, key.Str) {
					fValue := formatFunc(key.Str)
					if fValue != "" {
						values = append(values, fValue)
					}
				}
				return true
			})
		}
	}

	return values, nil
}

// WithMicroserviceInstance microservice instances completion
func WithMicroserviceInstance(flagName string, flagApplicationID string, clientFunc func() (*c8y.Client, error)) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}

			instances, err := getMicroserviceInstances(cmd, flagApplicationID, client, toComplete, func(s string) string { return s })
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			return instances, cobra.ShellCompDirectiveDefault
		})
		return cmd
	}
}
