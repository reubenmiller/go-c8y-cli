// v2-based application version get: the application flag drives iteration (pipe
// or --application); for each application (id or name, resolved via
// Applications.ResolveID) the selected version is fetched by --tag (ListByTag) or
// --version (ListByVersion). Use only one of the two selectors.
package get

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

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get a specific version of an application
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a specific version of an application",
		Long:  `Retrieve the selected version of an application in your tenant. To select the version, use only the version or only the tag query parameter`,
		Example: heredoc.Doc(`
$ c8y applications versions get --application 1234 --tag tag1
Get application version by tag

$ c8y applications versions get --application 1234 --version 1.0
Get application version by version name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("application", "", "Application (accepts pipeline)")
	cmd.Flags().String("version", "", "The version field of the application version")
	cmd.Flags().String("tag", "", "The tag of the application version")

	completion.WithOptions(
		cmd,
		completion.WithApplicationWithVersions("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("application", "application", false, "id", "name"),
		flags.WithCollectionProperty("-"),
		flags.WithPowershellName("Get-ApplicationVersion"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.applicationVersion+json", ""),
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

	if err := r.InputFlag("application"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		appID, err := client.Applications.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("application")), nil)
		if err != nil {
			return nil, err
		}
		version := in.String("version")
		tag := in.String("tag")
		return func(ctx context.Context) output.Seq {
			var res op.Result[jsonmodels.ApplicationVersion]
			if tag != "" {
				res = client.ApplicationVersions.ListByTag(ctx, appID, tag)
			} else {
				res = client.ApplicationVersions.ListByVersion(ctx, appID, version)
			}
			// collectionProperty "-": render the response as-is (no plucking).
			return c8ystream.FromResponse(res)
		}, nil
	})
}
