// v2-based current tenant get: returns the tenant associated with the current
// session via Tenants.Current.Get. Runs once (no pipeline); --withParent adds
// the parent tenant to the result.
package get

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/currenttenant"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get current tenant
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get current tenant",
		Long:  `Get the current tenant associated with the current session`,
		Example: heredoc.Doc(`
$ c8y currenttenant get
Get the current tenant (based on your current credentials)

$ c8y currenttenant get --withParent
Get the current tenant including the parent tenant (based on your current credentials)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().Bool("withParent", false, "When set to true, the returned result will contain the parent of the current tenant")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("", "", false),
		flags.WithPowershellName("Get-CurrentTenant"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.currentTenant+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := currenttenant.GetOptions{
			WithParent: in.Bool("withParent"),
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Tenants.Current.Get(ctx, opt))
		}, nil
	})
}
