// v2-based reset user password: the id flag drives iteration (pipe or --id); the
// body sets the new password (or, when none is given, requests a password-reset
// email) and the change is applied via Users.Update (PUT on the user). For
// Cumulocity users the id is the username, passed straight through.
package resetuserpassword

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

// ResetUserPasswordCmd command
type ResetUserPasswordCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewResetUserPasswordCmd creates a command to Reset user password
func NewResetUserPasswordCmd(f *cmdutil.Factory) *ResetUserPasswordCmd {
	ccmd := &ResetUserPasswordCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "resetUserPassword",
		Short: "Reset user password",
		Long: `The password can be reset either by issuing a password reset email (default), or be specifying a new password.

Note: In more recent Cumulocity versions,  you can't set a fixed password for another user.
`,
		Example: heredoc.Doc(`
$ c8y users resetUserPassword --id "myuser"
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
	cmd.Flags().String("newPassword", "", "New user password. Min: 6, max: 32 characters. Only Latin1 chars allowed")

	completion.WithOptions(
		cmd,
		completion.WithUser("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Reset-UserPassword"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.user+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ResetUserPasswordCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("newPassword", "password"),
		flags.WithRequiredTemplateString(`
{sendPasswordResetEmail: !std.objectHas(self, 'password')}`),
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
