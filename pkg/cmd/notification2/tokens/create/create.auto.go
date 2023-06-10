// Code generated from specification version 1.0.0: DO NOT EDIT
package create

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// CreateCmd command
type CreateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateCmd creates a command to Create a token
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a token",
		Long:  `Create a token to use for subscribing to notifications`,
		Example: heredoc.Doc(`
$ c8y notification2 tokens create --name testSubscription --subscriber testSubscriber --expiresInMinutes 1440
Create a new token for a subscription which is valid for 1 day

$ c8y notification2 tokens create --name testSubscription --subscriber testSubscriber --expiresInMinutes 30
Create a new token which is valid for 30 minutes
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("subscriber", "", "The subscriber name which the client wishes to be identified with. (accepts pipeline)")
	cmd.Flags().String("name", "", "The subscription name. This value must match the same that was used when the subscription was created.")
	cmd.Flags().Int("expiresInMinutes", 1440, "The token expiration duration.")
	cmd.Flags().Bool("shared", false, "Subscription is shared amongst multiple subscribers. >= 1016.x")
	cmd.Flags().String("type", "", "The subscription type. Currently the only supported type is notification .Other types may be added in future.")
	cmd.Flags().Bool("signed", false, "If true, the token will be securely signed by the Cumulocity IoT platform. >= 1016.x")
	cmd.Flags().Bool("nonPersistent", false, "If true, indicates that the created token refers to the non-persistent variant of the named subscription. >= 1016.x")

	completion.WithOptions(
		cmd,
		completion.WithNotification2SubscriptionName("name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("type", "notification"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("subscriber", "subscriber", false, "id"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithOverrideValue("subscriber", "subscriber"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("subscriber", "subscriber"),
		flags.WithStringValue("name", "subscription"),
		flags.WithIntValue("expiresInMinutes", "expiresInMinutes"),
		flags.WithBoolValue("shared", "shared", ""),
		flags.WithStringValue("type", "type"),
		flags.WithBoolValue("signed", "signed", ""),
		flags.WithBoolValue("nonPersistent", "nonPersistent", ""),
		flags.WithDefaultTemplateString(`
{subscriber: 'goc8ycli'}
`),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("subscriber", "subscription"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("notification2/token")
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
