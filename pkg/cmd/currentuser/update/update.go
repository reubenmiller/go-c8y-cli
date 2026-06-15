// v2-based current user update: builds the body (name/contact/password +
// --data/--template) and updates the current user via Users.CurrentUser.Update.
// Runs once (no pipeline).
package update

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
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

// NewUpdateCmd creates a command to Update current user
func NewUpdateCmd(f *cmdutil.Factory) *UpdateCmd {
	ccmd := &UpdateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update current user",
		Long:  `Update properties or settings of your user such as first/last name, email or password`,
		Example: heredoc.Doc(`
$ c8y currentuser update --lastName "Smith"
Update the current user's last name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("firstName", "", "User first name")
	cmd.Flags().String("lastName", "", "User last name")
	cmd.Flags().String("phone", "", "User phone number. Format: '+[country code][number]', has to be a valid MSISDN")
	cmd.Flags().String("email", "", "User email address")
	cmd.Flags().String("enabled", "", "User activation status (true/false)")
	cmd.Flags().String("password", "", "User password. Min: 6, max: 32 characters. Only Latin1 chars allowed")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("", "", false),
		flags.WithPowershellName("Update-CurrentUser"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.currentUser+json", ""),
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

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("firstName", "firstName"),
		flags.WithStringValue("lastName", "lastName"),
		flags.WithStringValue("phone", "phone"),
		flags.WithStringValue("email", "email"),
		flags.WithStringValue("enabled", "enabled"),
		flags.WithStringValue("password", "password"),
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
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.User] {
				return client.Users.CurrentUser.Update(ctx, body)
			})
		}, nil
	})
}
