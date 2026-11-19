// v2-based inventory additions list (deprecated; use `inventory children list
// --childType addition`): the id flag drives iteration; each parent (resolved as
// a managed object) has its child additions streamed via the ChildAdditions
// service, plucking the nested managed objects.
package list

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/children/childref"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects/child"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get child addition collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get child addition collection",
		Long:  `Get a collection of managedObject child additions`,
		Example: heredoc.Doc(`
$ c8y inventory additions list --id 12345
Get a list of the child additions of an existing managed object
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Managed object id. (required) (accepts pipeline)")
	cmd.Flags().String("query", "", "Additional query filter")
	cmd.Flags().String("queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
	cmd.Flags().String("orderBy", "", "Order by. e.g. _id asc or name asc or creationTime.date desc")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("references.#.managedObject"),
		flags.WithPowershellName("Get-ChildAdditionCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json", "application/vnd.com.nsn.cumulocity.managedObject+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	common, err := r.Config.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}
	rawOutput := r.Config.RawOutput()

	queryTemplate := cmd.Flag("queryTemplate").Value.String()
	orderBy := cmd.Flag("orderBy").Value.String()

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		parentID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("id")), nil)
		if err != nil {
			return nil, err
		}

		q := model.NewInventoryQuery()
		if raw := in.String("query"); raw != "" {
			q.AddFilterPart(applyQueryTemplate(queryTemplate, raw))
		} else if queryTemplate != "" {
			q.AddFilterPart(queryTemplate)
		}
		q.AddOrderBy(orderBy)

		opt := child.ListOptions{Query: q.Build()}
		opt.WithChildren = in.Bool("withChildren")
		opt.WithChildrenCount = in.Bool("withChildrenCount")
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
		}
		return c8ystream.ListCall(rawOutput, opt, childref.ListAll(client, childref.TypeAddition, parentID)), nil
	})
}

// applyQueryTemplate formats a raw query value with the --queryTemplate (a %s
// template). An empty template returns the value unchanged.
func applyQueryTemplate(template, value string) string {
	if template == "" {
		return value
	}
	return fmt.Sprintf(template, value)
}
