// v2-based firmware version list: the firmware flag drives iteration (pipe or
// --firmware); each firmware reference (id or name) is resolved via
// Repository.Firmware.ResolveID, then drives a typed Versions.ListAll call. The
// query excludes patches (not(has(c8y_Patch))); the c8y_FirmwareBinary type and
// bygroupid scoping are applied by the SDK service.
package list

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/firmware/firmwareversions"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get firmware package version collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get firmware package version collection",
		Long:  `Get a collection of firmware package versions (managedObjects) based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y firmware versions list --firmware 12345
Get a list of firmware package versions

$ c8y firmware list | c8y firmware versions list
Get a list of firmware and their versions (using pipeline)

$ c8y firmware versions get --id 12345 | c8y firmware versions list
Get all versions of a firmware using an existing version object
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("query", "", "Additional query filter")
	cmd.Flags().String("queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
	cmd.Flags().String("orderBy", "creationTime.date desc", "Order by. e.g. _id asc or name asc or creationTime.date desc")
	cmd.Flags().String("firmware", "", "Firmware package id or name (required) (accepts pipeline)")
	cmd.Flags().String("version", "", "Filter by version")
	cmd.Flags().String("url", "", "Filter by url")
	cmd.Flags().Bool("skipChildrenNames", false, "Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")
	cmd.Flags().Bool("withGroups", false, "When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.")
	cmd.Flags().Bool("withParents", true, "Include parent references")

	completion.WithOptions(
		cmd,
		completion.WithFirmware("firmware", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("firmware", "firmware", true, "additionParents.references.0.managedObject.id", "id"),
		flags.WithCollectionProperty("managedObjects"),
		flags.WithPowershellName("Get-FirmwareVersionCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectCollection+json", "application/vnd.com.nsn.cumulocity.managedObject+json"),
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

	if err := r.InputFlag("firmware"); err != nil {
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

	queryTemplate := cmd.Flag("queryTemplate").Value.String()
	rawOutput := r.Config.RawOutput()
	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		// Resolve the firmware reference to an id under a real (dry-disabled)
		// context so the bygroupid filter is correct even under --dry.
		fwID, err := client.Repository.Firmware.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("firmware")), nil)
		if err != nil {
			return nil, err
		}

		// Exclude patches and apply the optional url / piped query filters; the
		// type and version scoping are applied by the SDK service.
		parts := []string{"not(has(c8y_Patch))"}
		if raw := in.String("query"); raw != "" {
			parts = append(parts, applyQueryTemplate(queryTemplate, raw))
		} else if queryTemplate != "" {
			parts = append(parts, queryTemplate)
		}
		if u := in.String("url"); u != "" {
			parts = append(parts, fmt.Sprintf("(c8y_Firmware.url eq '%s')", u))
		}

		opt := firmwareversions.ListOptions{
			FirmwareID: fwID,
			Version:    in.String("version"),
			Query:      strings.Join(parts, " and "),
		}
		opt.SkipChildrenNames = in.Bool("skipChildrenNames")
		opt.WithChildren = in.Bool("withChildren")
		opt.WithChildrenCount = in.Bool("withChildrenCount")
		opt.WithParents = in.Bool("withParents")
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.Repository.Firmware.Versions.ListAll), nil
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
