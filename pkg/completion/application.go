package completion

import (
	"context"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/applications"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

const (
	applicationTypeHosted       = "HOSTED"
	applicationTypeMicroservice = "MICROSERVICE"
)

// listApplications returns one page of the applications matching the options
func listApplications(client *apiv2.Client, opt applications.ListOptions) ([]jsonmodels.Application, error) {
	if opt.PageSize == 0 {
		opt.PaginationOptions = pageSize(2000)
	}
	return client.Applications.List(context.Background(), opt).Collect()
}

// isOwnedByOtherTenant reports whether the application is owned by a tenant other than the client's tenant
func isOwnedByOtherTenant(client *apiv2.Client, app jsonmodels.Application) bool {
	if client.Auth().Tenant == "" {
		return false
	}
	owner := app.OwnerTenantID()
	return owner != "" && owner != client.Auth().Tenant
}

// formatApplication formats an application as "name<TAB>type | id: <id>"
func formatApplication(item jsonmodels.Application) string {
	return fmt.Sprintf("%s\t%s | id: %s", item.Name(), item.Type(), item.ID())
}

// WithApplication application completion
func WithApplication(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, err := listApplications(client, applications.ListOptions{})

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items {
				if toComplete == "" || MatchString(pattern, item.Name()) || MatchString(pattern, item.ID()) {
					values = append(values, formatApplication(item))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// WithApplicationContext application context completion
func WithApplicationContext(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, err := listApplications(client, applications.ListOptions{})

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items {
				contextPath := item.ContextPath()
				if toComplete == "" || MatchString(pattern, contextPath) || MatchString(pattern, item.Name()) || MatchString(pattern, item.ID()) {
					values = append(values, fmt.Sprintf("%s\t%s | type: %s  id: %s", contextPath, item.Name(), item.Type(), item.ID()))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// WithHostedApplication hosted application completion
func WithHostedApplication(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, err := listApplications(client, applications.ListOptions{})

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items {
				if !strings.EqualFold(item.Type(), applicationTypeHosted) {
					continue
				}

				// Ignore if not hosted in the current application
				if isOwnedByOtherTenant(client, item) {
					continue
				}

				if toComplete == "" || MatchString(pattern, item.Name()) || MatchString(pattern, item.ID()) {
					values = append(values, formatApplication(item))
				}
			}

			// If no results, then included applications that are not owned by the current tenant
			if len(values) == 0 {
				for _, item := range items {
					if !strings.EqualFold(item.Type(), applicationTypeHosted) {
						continue
					}

					if toComplete == "" || MatchString(pattern, item.Name()) || MatchString(pattern, item.ID()) {
						values = append(values, formatApplication(item))
					}
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// WithApplicationWithVersions applications with versions completion
func WithApplicationWithVersions(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, err := listApplications(client, applications.ListOptions{
				HasVersions: "true",
			})

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items {
				// Ignore if not hosted in the current application
				if isOwnedByOtherTenant(client, item) {
					continue
				}

				if toComplete == "" || MatchString(pattern, item.Name()) || MatchString(pattern, item.ID()) {
					values = append(values, formatApplication(item))
				}
			}

			// If no results, then included applications that are not owned by the current tenant
			if len(values) == 0 {
				for _, item := range items {
					if toComplete == "" || MatchString(pattern, item.Name()) || MatchString(pattern, item.ID()) {
						values = append(values, formatApplication(item))
					}
				}
			}

			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// WithMicroservice microservice completion
func WithMicroservice(flagName string, clientFunc ClientFunc) Option {
	return func(cmd *cobra.Command) *cobra.Command {
		_ = cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			client, err := clientFunc()
			if err != nil {
				return []string{err.Error()}, cobra.ShellCompDirectiveDefault
			}
			items, err := listApplications(client, applications.ListOptions{})

			if err != nil {
				values := []string{fmt.Sprintf("error. %s", err)}
				return values, cobra.ShellCompDirectiveError
			}
			values := []string{}
			pattern := "*" + toComplete + "*"
			for _, item := range items {
				if !strings.EqualFold(item.Type(), applicationTypeMicroservice) {
					continue
				}
				if toComplete == "" || MatchString(pattern, item.Name()) || MatchString(pattern, item.ID()) {
					values = append(values, formatApplication(item))
				}
			}
			return values, cobra.ShellCompDirectiveNoFileComp
		})
		return cmd
	}
}

// WithMicroserviceLoggers microservice logger completion (requires the microservice name)
func WithMicroserviceLoggers(flagName string, flagNameMicroserviceName string, clientFunc ClientFunc) Option {
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
			result := client.Microservices.Loggers.List(context.Background(), microserviceName)
			if result.Err != nil {
				values := []string{fmt.Sprintf("error. %s", result.Err)}
				return values, cobra.ShellCompDirectiveError
			}

			pattern := "*" + toComplete + "*"

			// The loggers collection is a map keyed by the logger name
			gjson.ParseBytes(result.Data.Bytes()).ForEach(func(key, value gjson.Result) bool {
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

func getMicroserviceByName(client *apiv2.Client, name string) (string, error) {
	apps, err := client.Applications.List(context.Background(), applications.ListOptions{
		Name:              name,
		Type:              applicationTypeMicroservice,
		PaginationOptions: pageSize(1),
	}).Collect()
	if err != nil {
		return "", err
	}

	for _, app := range apps {
		return app.ID(), nil
	}
	return "", ErrNotFound
}

func getMicroserviceInstances(cmd *cobra.Command, flagApplicationID string, client *apiv2.Client, toComplete string, formatFunc func(string) string) ([]string, error) {
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
	items, err := listManagedObjects(client, managedobjects.ListOptions{
		Type:              "c8y_Application_" + deviceID,
		PaginationOptions: pageSize(1),
	})

	if err != nil {
		return nil, err
	}
	values := []string{}
	for _, item := range items {
		if v := item.Get("c8y_Status.instances"); v.IsObject() {
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
func WithMicroserviceInstance(flagName string, flagApplicationID string, clientFunc ClientFunc) Option {
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
