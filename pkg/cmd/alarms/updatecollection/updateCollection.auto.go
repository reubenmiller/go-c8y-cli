// Code generated from specification version 1.0.0: DO NOT EDIT
package updatecollection

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// UpdateCollectionCmd command
type UpdateCollectionCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUpdateCollectionCmd creates a command to Update alarm collection
func NewUpdateCollectionCmd(f *cmdutil.Factory) *UpdateCollectionCmd {
	ccmd := &UpdateCollectionCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "updateCollection",
		Short: "Update alarm collection",
		Long:  `Update the status of a collection of alarms by using a filter. Currently only the status of alarms can be changed`,
		Example: heredoc.Doc(`
$ c8y alarms updateCollection --device 12345 --status ACTIVE --newStatus ACKNOWLEDGED
Update the status of all active alarms on a device to ACKNOWLEDGED
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject that the alarm originated from (accepts pipeline)")
	cmd.Flags().String("status", "", "The status of the alarm: ACTIVE, ACKNOWLEDGED or CLEARED. If status was not appeared, new alarm will have status ACTIVE. Must be upper-case.")
	cmd.Flags().String("severity", "", "The severity of the alarm: CRITICAL, MAJOR, MINOR or WARNING. Must be upper-case.")
	cmd.Flags().Bool("resolved", false, "When set to true only resolved alarms will be removed (the one with status CLEARED), false means alarms with status ACTIVE or ACKNOWLEDGED.")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of alarm occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of alarm occurrence.")
	cmd.Flags().String("newStatus", "", "New status to be applied to all of the matching alarms (required)")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("status", "ACTIVE", "ACKNOWLEDGED", "CLEARED"),
		completion.WithValidateSet("severity", "CRITICAL", "MAJOR", "MINOR", "WARNING"),
		completion.WithValidateSet("newStatus", "ACTIVE", "ACKNOWLEDGED", "CLEARED"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),

		flags.WithExtendedPipelineSupport("device", "source", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("alarms"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("newStatus")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
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
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "device", "source"),
		flags.WithStringValue("status", "status"),
		flags.WithStringValue("severity", "severity"),
		flags.WithBoolValue("resolved", "resolved", ""),
		flags.WithEncodedRelativeTimestamp("dateFrom", "dateFrom", ""),
		flags.WithEncodedRelativeTimestamp("dateTo", "dateTo", ""),
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
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		flags.WithStringValue("newStatus", "status"),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("alarm/alarms")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "PUT",
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
