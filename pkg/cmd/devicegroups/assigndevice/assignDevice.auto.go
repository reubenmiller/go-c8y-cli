// Code generated from specification version 1.0.0: DO NOT EDIT
package assigndevice

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

// AssignDeviceCmd command
type AssignDeviceCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewAssignDeviceCmd creates a command to Assign device to group
func NewAssignDeviceCmd(f *cmdutil.Factory) *AssignDeviceCmd {
	ccmd := &AssignDeviceCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "assignDevice",
		Short:  "Assign device to group",
		Long:   `Assigns a device to a group. The device will be a childAsset of the group`,
		Hidden: true,

		Example: heredoc.Doc(`
$ c8y devicegroups assignDevice --group 12345 --newChildDevice 43234
Add a device to a group

$ c8y devicegroups assignDevice --group 12345 --newChildDevice 43234,99292,12222
Add multiple devices to a group
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Group (required)")
	cmd.Flags().StringSlice("newChildDevice", []string{""}, "New device to be added to the group as an child asset (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithDeviceGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDevice("newChildDevice", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),

		flags.WithExtendedPipelineSupport("newChildDevice", "managedObject.id", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("group", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("newChildDevice", "deviceId", "source.id", "managedObject.id", "id"),

		flags.WithCollectionProperty("managedObject"),
		flags.WithDeprecationNotice("please use 'c8y devicegroups children assign --childType asset' instead"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("group")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *AssignDeviceCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	// Runtime flag options
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
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
	body := mapbuilder.NewInitializedMapBuilder(true)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "newChildDevice", "managedObject.id"),
		cmdutil.WithTemplateValue(n.factory, cfg),
		flags.WithTemplateVariablesValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects/{id}/childAssets")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithDeviceGroupByNameFirstMatch(client, args, "group", "id"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "POST",
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
