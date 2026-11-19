// v2-based application update: the id flag drives iteration (pipe or --id) and the
// body builder (name/key/availability/... + --data/--template) is evaluated per
// item, then both feed Applications.Update (the id/name reference is resolved
// internally by the SDK).
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

// NewUpdateCmd creates a command to Update application
func NewUpdateCmd(f *cmdutil.Factory) *UpdateCmd {
	ccmd := &UpdateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update application",
		Long:  `Update an existing application`,
		Example: heredoc.Doc(`
$ c8y applications update --id 12345 --availability MARKET
Update an application's availability
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Application id (required) (accepts pipeline)")
	cmd.Flags().String("name", "", "Name of application")
	cmd.Flags().String("key", "", "Shared secret of application")
	cmd.Flags().String("availability", "", "Access level for other tenants. Possible values are : MARKET, PRIVATE (default)")
	cmd.Flags().String("contextPath", "", "contextPath of the hosted application")
	cmd.Flags().String("resourcesUrl", "", "URL to application base directory hosted on an external server")
	cmd.Flags().String("resourcesUsername", "", "authorization username to access resourcesUrl")
	cmd.Flags().String("resourcesPassword", "", "authorization password to access resourcesUrl")
	cmd.Flags().String("externalUrl", "", "URL to the external application")

	completion.WithOptions(
		cmd,
		completion.WithApplication("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("availability", "MARKET", "PRIVATE"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPipelineAliases("id", "id"),
		flags.WithPowershellName("Update-Application"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.application+json", ""),
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
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("key", "key"),
		flags.WithStringValue("availability", "availability"),
		flags.WithStringValue("contextPath", "contextPath"),
		flags.WithStringValue("resourcesUrl", "resourcesUrl"),
		flags.WithStringValue("resourcesUsername", "resourcesUsername"),
		flags.WithStringValue("resourcesPassword", "resourcesPassword"),
		flags.WithStringValue("externalUrl", "externalUrl"),
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
		ref := c8ystream.NameOrID(in.String("id"))
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Application] {
				return client.Applications.Update(ctx, ref, body)
			})
		}, nil
	})
}
