// Code generated from specification version 1.0.0: DO NOT EDIT
package unassign

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// UnassignCmd command
type UnassignCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUnassignCmd creates a command to Delete child addition reference
func NewUnassignCmd(f *cmdutil.Factory) *UnassignCmd {
	ccmd := &UnassignCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "unassign",
		Short: "Delete child addition reference",
		Long:  `Delete child addition reference`,
		Example: heredoc.Doc(`
$ c8y devices additions unassign --device 12345 --child 22553
Unassign a child device from its parent device
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "ManagedObject id (required)")
	cmd.Flags().StringSlice("child", []string{""}, "Child addition reference (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDevice("child", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),

		flags.WithExtendedPipelineSupport("child", "child", true, "deviceId", "source.id", "managedObject.id", "id"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("device")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UnassignCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
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
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
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
	body := mapbuilder.NewInitializedMapBuilder(false)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects/{device}/childAdditions/{child}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "device", "device"),
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "child", "child"),
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
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
