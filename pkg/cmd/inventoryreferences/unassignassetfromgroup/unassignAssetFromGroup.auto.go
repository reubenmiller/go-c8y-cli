// Code generated from specification version 1.0.0: DO NOT EDIT
package unassignassetfromgroup

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// UnassignAssetFromGroupCmd command
type UnassignAssetFromGroupCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
	Config  func() (*config.Config, error)
	Client  func() (*c8y.Client, error)
}

// NewUnassignAssetFromGroupCmd creates a command to Unassign asset from group
func NewUnassignAssetFromGroupCmd(f *cmdutil.Factory) *UnassignAssetFromGroupCmd {
	ccmd := &UnassignAssetFromGroupCmd{
		factory: f,
		Config:  f.Config,
		Client:  f.Client,
	}
	cmd := &cobra.Command{
		Use:   "unassignAssetFromGroup",
		Short: "Unassign asset from group",
		Long:  `Unassign/delete an asset (device or group) from a group`,
		Example: heredoc.Doc(`
$ c8y inventoryReferences unassignAssetFromGroup --device 12345 --childDevice 22553
Unassign a child device from its parent device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Asset id (required)")
	cmd.Flags().StringSlice("childDevice", []string{""}, "Child device (accepts pipeline)")
	cmd.Flags().StringSlice("childGroup", []string{""}, "Child device group")

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),

		flags.WithExtendedPipelineSupport("childDevice", "reference", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("group")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UnassignAssetFromGroupCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.Config()
	if err != nil {
		return err
	}
	client, err := n.Client()
	if err != nil {
		return err
	}
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithProcessingModeValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects/{group}/childAssets/{reference}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "group", "group"),
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "childDevice", "reference"),
		c8yfetcher.WithDeviceGroupByNameFirstMatch(client, args, "childGroup", "reference"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "DELETE",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.DryRun(),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
