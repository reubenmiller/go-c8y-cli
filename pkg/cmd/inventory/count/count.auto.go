// Code generated from specification version 1.0.0: DO NOT EDIT
package count

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

// CountCmd command
type CountCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCountCmd creates a command to Get managed object count
func NewCountCmd(f *cmdutil.Factory) *CountCmd {
	ccmd := &CountCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Get managed object count",
		Long:  `Retrieve the total number of managed objects (e.g. devices, assets, etc.) registered in your tenant, or a subset based on queries.`,
		Example: heredoc.Doc(`
$ c8y inventory count
Get count of managed objects

$ c8y inventory count --text myname
Get count of managed objects matching text (using Cumulocity text search algorithm)

$ c8y inventory count --type "c8y_Sensor"
Get count of managed objects with a specific type value

$ c8y inventory count --type "c8y_Sensor" --owner "device_mylinuxbox01"
Get count of managed objects with a specific type value and owner

$ c8y inventory count --fragmentType "c8y_IsDevice"
Get total number of devices
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("ids", []string{""}, "List of ids.")
	cmd.Flags().String("type", "", "ManagedObject type. (accepts pipeline)")
	cmd.Flags().String("fragmentType", "", "ManagedObject fragment type.")
	cmd.Flags().String("owner", "", "List of managed objects that are owned by the given username.")
	cmd.Flags().String("text", "", "managed objects containing a text value starting with the given text (placeholder {text}). Text value is any alphanumeric string starting with a latin letter (A-Z or a-z).")
	cmd.Flags().String("childAdditionId", "", "Search for a specific child addition and list all the groups to which it belongs.")
	cmd.Flags().String("childAssetId", "", "Search for a specific child asset and list all the groups to which it belongs.")
	cmd.Flags().StringSlice("childDeviceId", []string{""}, "Search for a specific child device and list all the groups to which it belongs.")

	completion.WithOptions(
		cmd,
		completion.WithDevice("childDeviceId", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,

		flags.WithExtendedPipelineSupport("type", "type", false, "type"),
		flags.WithPipelineAliases("childDeviceId", "deviceId", "source.id", "managedObject.id", "id"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CountCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithStringSliceCSV("ids", "ids", ""),
		flags.WithStringValue("type", "type"),
		flags.WithStringValue("fragmentType", "fragmentType"),
		flags.WithStringValue("owner", "owner"),
		flags.WithStringValue("text", "text"),
		flags.WithStringValue("childAdditionId", "childAdditionId"),
		flags.WithStringValue("childAssetId", "childAssetId"),
		c8yfetcher.WithDeviceByNameFirstMatch(n.factory, args, "childDeviceId", "childDeviceId"),
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
	path := flags.NewStringTemplate("inventory/managedObjects/count")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
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
