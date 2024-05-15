// Code generated from specification version 1.0.0: DO NOT EDIT
package create_webssh

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

// CreateWebsshCmd command
type CreateWebsshCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateWebsshCmd creates a command to Create web ssh configuration
func NewCreateWebsshCmd(f *cmdutil.Factory) *CreateWebsshCmd {
	ccmd := &CreateWebsshCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create-webssh",
		Short: "Create web ssh configuration",
		Long: `Create web ssh configuration
`,
		Example: heredoc.Doc(`
$ c8y remoteaccess configurations create-webssh
Create a webssh configuration

$ c8y remoteaccess configurations create-webssh --hostname 127.0.0.1 --port 2222
Create a webssh configuration with a custom hostname and port
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device (accepts pipeline)")
	cmd.Flags().String("name", "webssh", "Connection name")
	cmd.Flags().String("hostname", "", "Hostname")
	cmd.Flags().Int("port", 0, "Port")
	cmd.Flags().String("credentialsType", "USER_PASS", "Credentials type")
	cmd.Flags().String("privateKey", "", "Private ssh key")
	cmd.Flags().String("publicKey", "", "Public ssh key")
	cmd.Flags().String("username", "", "Username")
	cmd.Flags().String("password", "", "Username")
	cmd.Flags().String("protocol", "SSH", "Protocol")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("credentialsType", "USER_PASS"),
		completion.WithValidateSet("protocol", "PASSTHROUGH", "SSH"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),

		flags.WithExtendedPipelineSupport("device", "device", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateWebsshCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("hostname", "hostname"),
		flags.WithIntValue("port", "port"),
		flags.WithStringValue("credentialsType", "credentialsType"),
		flags.WithStringValue("privateKey", "privateKey"),
		flags.WithStringValue("publicKey", "publicKey"),
		flags.WithStringValue("username", "username"),
		flags.WithStringValue("password", "password"),
		flags.WithStringValue("protocol", "protocol"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("hostname", "port", "protocol", "name"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("/service/remoteaccess/devices/{device}/configurations")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithDeviceByNameFirstMatch(n.factory, args, "device", "device"),
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
