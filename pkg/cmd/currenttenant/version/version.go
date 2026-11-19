// v2-based current-tenant version: returns the platform (backend) version of the
// current tenant via Tenants.Current.GetVersion (the "system"/"version" system
// option). Runs once (no pipeline).
package version

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// VersionCmd command
type VersionCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewVersionCmd creates a command to Get tenant version
func NewVersionCmd(f *cmdutil.Factory) *VersionCmd {
	ccmd := &VersionCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Get tenant version",
		Long:  `Get tenant platform (backend) version`,
		Example: heredoc.Doc(`
$ c8y currenttenant version
Get the platform version of the current tenant
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithPowershellName("Get-TenantVersion"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.option+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *VersionCmd) RunE(cmd *cobra.Command, args []string) error {
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
			return c8ystream.FromResult(client.Tenants.Current.GetVersion(ctx))
		}, nil
	})
}
