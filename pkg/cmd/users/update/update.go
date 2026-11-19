// v2-based user update: the id flag drives iteration (pipe or --id) and the body
// builder (profile fields + --data/--template) is evaluated per item, then both
// feed Users.Update. For Cumulocity users the id is the username, passed straight
// through.
package update

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/users"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// UpdateCmd command
type UpdateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUpdateCmd creates a command to Update user
func NewUpdateCmd(f *cmdutil.Factory) *UpdateCmd {
	ccmd := &UpdateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update user",
		Long:  `Update properties, reset password or enable/disable for a user in a tenant`,
		Example: heredoc.Doc(`
$ c8y users update --id "myuser" --firstName "Simon"
Update a user
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "User id (required) (accepts pipeline)")
	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().String("firstName", "", "User first name")
	cmd.Flags().String("lastName", "", "User last name")
	cmd.Flags().String("displayName", "", "The user's display name in Cumulocity")
	cmd.Flags().String("phone", "", "User phone number. Format: '+[country code][number]', has to be a valid MSISDN")
	cmd.Flags().String("email", "", "User email address")
	cmd.Flags().Bool("enabled", false, "User activation status (true/false)")
	cmd.Flags().String("password", "", "User password. Min: 6, max: 32 characters. Only Latin1 chars allowed")
	cmd.Flags().Bool("shouldResetPassword", false, "User must reset password on next login")
	cmd.Flags().Bool("sendPasswordResetEmail", false, "Send password reset email to the user instead of setting a password")
	cmd.Flags().Bool("newsletter", false, "Indicates whether the user is subscribed to the newsletter or not")
	cmd.Flags().String("customProperties", "", "Custom properties to be added to the user")

	completion.WithOptions(
		cmd,
		completion.WithUser("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("sendPasswordResetEmail", "true", "false"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Update-User"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.user+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("firstName", "firstName"),
		flags.WithStringValue("lastName", "lastName"),
		flags.WithStringValue("displayName", "displayName"),
		flags.WithStringValue("phone", "phone"),
		flags.WithStringValue("email", "email"),
		flags.WithBoolValue("enabled", "enabled", ""),
		flags.WithStringValue("password", "password"),
		flags.WithBoolValue("shouldResetPassword", "shouldResetPassword", ""),
		flags.WithBoolValue("sendPasswordResetEmail", "sendPasswordResetEmail", ""),
		flags.WithBoolValue("newsletter", "newsletter", ""),
		flags.WithDataValue("customProperties", "customProperties"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		tenant := in.String("tenant")
		if tenant == "" {
			tenant = n.factory.GetTenant()
		}
		opt := users.UpdateOptions{
			Tenant: tenant,
			ID:     users.UserRef(in.String("id")),
		}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.User] {
				return client.Users.Update(ctx, opt, body)
			})
		}, nil
	})
}
