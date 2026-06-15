// v2-based current user logout: logs out the current user via Users.Logout,
// invalidating the session token (OAUTH2_INTERNAL). Runs once; yields no output.
package logout

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// LogoutCmd command
type LogoutCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewLogoutCmd creates a command to Logout current user
func NewLogoutCmd(f *cmdutil.Factory) *LogoutCmd {
	ccmd := &LogoutCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout current user",
		Long:  `Logout the current user. This will invalidate the token associated with the user when using OAUTH2_INTERNAL`,
		Example: heredoc.Doc(`
$ c8y currentuser logout
Log out the current user
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("", "", false),
		flags.WithPowershellName("Invoke-UserLogout"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *LogoutCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitStatus(ctx, func(ctx context.Context) op.Result[core.NoContent] {
				return client.Users.Logout(ctx)
			})
		}, nil
	})
}
