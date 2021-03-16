// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// DeleteManagedObjectChildAssetReferenceCmd command
type DeleteManagedObjectChildAssetReferenceCmd struct {
	*baseCmd
}

// NewDeleteManagedObjectChildAssetReferenceCmd creates a command to Delete child asset reference
func NewDeleteManagedObjectChildAssetReferenceCmd() *DeleteManagedObjectChildAssetReferenceCmd {
	ccmd := &DeleteManagedObjectChildAssetReferenceCmd{}
	cmd := &cobra.Command{
		Use:   "unassignAssetFromGroup",
		Short: "Delete child asset reference",
		Long:  `Delete child asset reference`,
		Example: heredoc.Doc(`
$ c8y inventoryReferences unassignAssetFromGroup --group 12345 --childDevice 22553
Unassign a child device from its parent device
        `),
		PreRunE: validateDeleteMode,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Asset id (required)")
	cmd.Flags().StringSlice("childDevice", []string{""}, "Child device (accepts pipeline)")
	cmd.Flags().StringSlice("childGroup", []string{""}, "Child device group")
	addProcessingModeFlag(cmd)

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("childDevice", "reference", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("group")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

// RunE executes the command
func (n *DeleteManagedObjectChildAssetReferenceCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
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
		WithDeviceGroupByNameFirstMatch(args, "group", "group"),
		WithDeviceByNameFirstMatch(args, "childDevice", "reference"),
		WithDeviceGroupByNameFirstMatch(args, "childGroup", "reference"),
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
		IgnoreAccept: cliConfig.IgnoreAcceptHeader(),
		DryRun:       cliConfig.DryRun(),
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
