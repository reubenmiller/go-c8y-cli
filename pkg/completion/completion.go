package completion

import (
	"context"
	"fmt"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/spf13/cobra"
)

// ClientFunc returns the go-c8y v2 client which the completion lookups use.
// Commands pass their factory's ClientV2 method (cmdutil.Factory.ClientV2), so
// completion is built exactly like the client of the c8ystream Runner.
type ClientFunc = func() (*apiv2.Client, error)

// GetFlagStringValues get string slice from either a string slice or string flag
func GetFlagStringValues(cmd *cobra.Command, name string) ([]string, error) {
	items, err := cmd.Flags().GetStringSlice(name)

	if err != nil {
		// fallback to string
		item, strErr := cmd.Flags().GetString(name)

		if strErr != nil {
			return nil, strErr
		}
		if item != "" {
			items = append(items, item)
		}
	}
	return items, nil
}

// pageSize returns the pagination options for a single page of the given size
func pageSize(size int) pagination.PaginationOptions {
	return pagination.PaginationOptions{PageSize: size}
}

// listManagedObjects returns one page of the managed objects matching the options.
// Completion never runs in dry-run mode, as the v2 client only enables it via
// the context.
func listManagedObjects(client *apiv2.Client, opt managedobjects.ListOptions) ([]jsonmodels.ManagedObject, error) {
	return client.ManagedObjects.List(context.Background(), opt).Collect()
}

// queryManagedObjects returns one page of the managed objects matching the inventory query.
func queryManagedObjects(client *apiv2.Client, query string, size int) ([]jsonmodels.ManagedObject, error) {
	return listManagedObjects(client, managedobjects.ListOptions{
		Query:             query,
		PaginationOptions: pageSize(size),
	})
}

// findDevicesByName returns the devices whose name matches the given name (wildcards are supported).
func findDevicesByName(client *apiv2.Client, name string, size int) ([]jsonmodels.ManagedObject, error) {
	return queryManagedObjects(client, fmt.Sprintf("(name eq '%s') and has(%s)", name, "c8y_IsDevice"), size)
}

// formatManagedObject formats a managed object as "name<TAB>type | id: <id>"
func formatManagedObject(item jsonmodels.ManagedObject) string {
	return fmt.Sprintf("%s\t%s | id: %s", item.Name(), item.Type(), item.ID())
}

// managedObjectCompletion completes managed objects which match the query
// built from the wildcard pattern ("*<toComplete>*").
func managedObjectCompletion(clientFunc ClientFunc, query func(pattern string) string) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, err := clientFunc()
		if err != nil {
			return []string{err.Error()}, cobra.ShellCompDirectiveDefault
		}

		pattern := "*" + toComplete + "*"
		items, err := queryManagedObjects(client, query(pattern), 100)
		if err != nil {
			values := []string{fmt.Sprintf("error. %s", err)}
			return values, cobra.ShellCompDirectiveError
		}
		values := []string{}
		for _, item := range items {
			if toComplete == "" || MatchString(pattern, item.Name()) || MatchString(pattern, item.ID()) {
				values = append(values, formatManagedObject(item))
			}
		}
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}
