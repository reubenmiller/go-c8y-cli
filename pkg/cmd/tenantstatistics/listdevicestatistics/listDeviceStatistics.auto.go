// Code generated from specification version 1.0.0: DO NOT EDIT
package listdevicestatistics

import (
	"fmt"
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

// ListDeviceStatisticsCmd command
type ListDeviceStatisticsCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListDeviceStatisticsCmd creates a command to Retrieve device statistics
func NewListDeviceStatisticsCmd(f *cmdutil.Factory) *ListDeviceStatisticsCmd {
	ccmd := &ListDeviceStatisticsCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "listDeviceStatistics",
		Short: "Retrieve device statistics",
		Long:  `Retrieve device statistics from a specific tenant (by a given ID). Either daily or monthly.`,
		Example: heredoc.Doc(`
$ c8y statistics device list
Get tenant summary statistics for the current tenant
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("date", "-1d", "Date of the queried day. By monthly, the day will be ignored.")
	cmd.Flags().String("type", "daily", "Aggregation type. e.g. daily or monthly (required)")
	cmd.Flags().String("tenant", "", "Tenant id. Defaults to current tenant (based on credentials)")
	cmd.Flags().StringSlice("deviceId", []string{""}, "The ID of the device to search for. (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("type", "daily", "monthly"),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDevice("deviceId", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,

		flags.WithExtendedPipelineSupport("deviceId", "deviceId", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("type")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListDeviceStatisticsCmd) RunE(cmd *cobra.Command, args []string) error {
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
		c8yfetcher.WithDeviceByNameFirstMatch(n.factory, args, "deviceId", "deviceId"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(query)

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
	path := flags.NewStringTemplate("/tenant/statistics/device/{tenant}/{type}/{date}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithRelativeDate(false, "date", "date", ""),
		flags.WithStringValue("type", "type"),
		flags.WithStringDefaultValue(client.TenantName, "tenant", "tenant"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "GET",
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
