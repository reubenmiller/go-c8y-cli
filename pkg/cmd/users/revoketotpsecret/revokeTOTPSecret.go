// v2-based revoke TOTP secret: the id flag drives iteration (pipe or --id); each
// user's TOTP (TFA) secret is revoked via Users.RevokeTOTPSecret. For Cumulocity
// users the id is the username, passed straight through. Success yields no output.
package revoketotpsecret

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/users"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// RevokeTOTPSecretCmd command
type RevokeTOTPSecretCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewRevokeTOTPSecretCmd creates a command to Revoke a user's TOTP (TFA) secret
func NewRevokeTOTPSecretCmd(f *cmdutil.Factory) *RevokeTOTPSecretCmd {
	ccmd := &RevokeTOTPSecretCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "revokeTOTPSecret",
		Short: "Revoke a user's TOTP (TFA) secret",
		Long: `Revoke/delete a user's TOTP (TFA) secret to force them to setup TFA again.

This is required when the user loses their TFA configuration, or it is compromised.
`,
		Example: heredoc.Doc(`
$ c8y users revokeTOTPSecret --id "myuser"
Revoke a user's TOTP (TFA) secret
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.DeleteModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("tenant", "", "Tenant")
	cmd.Flags().StringSlice("id", []string{""}, "User id (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithTenantID("tenant", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithUser("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPipelineAliases("tenant", "tenant", "owner.tenant.id"),
		flags.WithPowershellName("Remove-UserTOTPSecret"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *RevokeTOTPSecretCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
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
		opt := users.RevokeTOTPSecretOptions{
			Tenant: tenant,
			ID:     users.UserRef(in.String("id")),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Users.RevokeTOTPSecret(ctx, opt)
			})
		}, nil
	})
}
