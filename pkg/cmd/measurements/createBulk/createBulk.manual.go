package createBulk

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

// CreateBulkCmd command
type CreateBulkCmd struct {
	*subcommand.SubCommand

	batchSize int

	factory *cmdutil.Factory
}

// CreateBulkCmd creates a command to Create multiple measurements via the bulk measurement api
func NewCreateBulkCmd(f *cmdutil.Factory) *CreateBulkCmd {
	ccmd := &CreateBulkCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "createBulk",
		Short: "Create bulk measurements",
		Long: heredoc.Doc(`
			Create new measurements using bulk api

			Collects inputs and groups them into batches and send them to the server.

			Notes:
			* The template is applied to each piped measurement before being grouped
		`),
		Example: heredoc.Doc(`
$ c8y measurements list -p 1000 --device 11111  | c8y measurements createBulk --device 22222 --batchSize 100
Copy measurements from one device to another, but create measurement in batches of 100

$ c8y measurements list -p 20 --valueFragmentType c8y_Temperature --valueFragmentSeries T \
	| c8y measurements createBulk \
		--type testme \
		--template "{c8y_Temperature+:{T+:{value: input.value.c8y_Temperature.T.value + 100}}}"
Copy measurements from one device to another modifying the measurements slightly but adding 100 to the value
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "The ManagedObject which is the source of this measurement. (accepts pipeline)")
	cmd.Flags().String("time", "", "Time of the measurement. Defaults to current timestamp.")
	cmd.Flags().String("type", "", "The most specific type of this entire measurement.")
	cmd.Flags().IntVar(&ccmd.batchSize, "batchSize", 10, "Batch size. Number of measurements per request to send")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "source.id", false, "deviceId", "source.id", "managedObject.id", "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateBulkCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithStaticStringValue("Content-Type", "application/vnd.com.nsn.cumulocity.measurementcollection+json"),
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

	// Wrap body around an nested array. It will be used
	// when the body is marshalled
	body.ArraySize = n.batchSize
	body.ArrayPrefix = `{"measurements":[`
	body.ArraySuffix = `]}`

	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		c8yfetcher.WithDeviceByNameFirstMatch(client, args, "device", "source.id"),
		flags.WithRelativeTimestamp("time", "time", ""),
		flags.WithStringValue("type", "type"),
		flags.WithDefaultTemplateString(`{time: _.Now('0s')} + (if std.isObject(input.value) then input.value else {source:{id:input.value}}) + {id:: '', 'self':: ''}`),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("type", "time", "source.id"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("measurement/measurements")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
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
