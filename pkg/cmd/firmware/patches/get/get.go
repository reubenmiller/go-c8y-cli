// v2-based firmware patch get: the id flag drives iteration (pipe or --id). A
// piped plain managed-object id is fetched directly; a patch version name (with
// --firmware) is resolved scoped to that firmware via the patch resolver by
// Patches.Get.
package get

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/firmware/firmwarepatches"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get firmware patch
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get firmware patch",
		Long:  `Get an existing firmware patch`,
		Example: heredoc.Doc(`
$ c8y firmware patches get --firmware 11111 --id 1.0.0
Get a firmware patch
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Firmware patch id or name (required) (accepts pipeline)")
	cmd.Flags().String("firmware", "", "Firmware package id or name (used to help completion be more accurate)")
	cmd.Flags().Bool("skipChildrenNames", false, "Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")
	cmd.Flags().Bool("withGroups", false, "When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.")
	cmd.Flags().Bool("withParents", true, "Include a flat list of all parents and grandparents of the given object")

	completion.WithOptions(
		cmd,
		completion.WithFirmwarePatch("id", "firmware", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithFirmware("firmware", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Get-FirmwarePatch"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObject+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetCmd) RunE(cmd *cobra.Command, args []string) error {
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

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		ref := patchRef(in.String("id"), in.String("firmware"))
		opt := firmwarepatches.GetOptions{
			WithParents:       in.Bool("withParents"),
			WithChildren:      in.Bool("withChildren"),
			WithChildrenCount: in.Bool("withChildrenCount"),
			SkipChildrenNames: in.Bool("skipChildrenNames"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Repository.Firmware.Patches.Get(ctx, ref, opt))
		}, nil
	})
}

// patchRef builds the patch resolver reference: a plain managed-object id (or
// when no firmware scope is given) is used directly; otherwise the id is a patch
// version name scoped to the firmware (by id or name).
func patchRef(id, firmware string) string {
	if id == "" || firmware == "" || c8ydata.IsID(id) {
		return id
	}
	if c8ydata.IsID(firmware) {
		return firmwarepatches.NewRef().ByVersion(id, firmware)
	}
	return firmwarepatches.NewRef().ByVersionAndName(id, firmware)
}
