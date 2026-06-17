// v2-based firmware patch delete: the id flag drives iteration (pipe or --id). A
// piped plain managed-object id is deleted directly; a patch version name (with
// --firmware) is resolved scoped to that firmware by Patches.Delete.
// --forceCascade (default true) removes the related binary.
package delete

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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/firmware/firmwarepatches"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DeleteCmd command
type DeleteCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDeleteCmd creates a command to Delete firmware package version patch
func NewDeleteCmd(f *cmdutil.Factory) *DeleteCmd {
	ccmd := &DeleteCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete firmware package version patch",
		Long:  `Delete an existing firmware package version patch`,
		Example: heredoc.Doc(`
$ c8y firmware patches delete --id 12345
Delete a firmware patch and related binary

$ c8y firmware patches delete --id 12345 --forceCascade=false
Delete a firmware patch (but keep the related binary)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Firmware patch id or name (required) (accepts pipeline)")
	cmd.Flags().String("firmware", "", "Firmware id or name (used to help completion be more accurate)")
	cmd.Flags().Bool("forceCascade", true, "Remove version and any related binaries")

	completion.WithOptions(
		cmd,
		completion.WithFirmwarePatch("id", "firmware", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithFirmware("firmware", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Remove-FirmwarePatch"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteCmd) RunE(cmd *cobra.Command, args []string) error {
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
		opt := firmwarepatches.DeleteOptions{ForceCascade: in.Bool("forceCascade")}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Repository.Firmware.Patches.Delete(ctx, ref, opt)
			})
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
