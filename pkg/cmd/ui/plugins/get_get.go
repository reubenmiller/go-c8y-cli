// v2-based UI plugin get: the id flag drives iteration (pipe or --id); each
// reference (a plugin id or name/contextPath, resolved internally by the SDK) is
// fetched via UIPlugins.Get.

package plugins

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get UI plugin
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get UI plugin",
		Long:  `Get a UI plugin`,
		Example: heredoc.Doc(`
$ c8y ui plugins get --id 1234
Get ui plugin by id

$ c8y ui plugins get --id myplugin
Get ui plugin by name
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Plugin (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithUIPlugin("id", ccmd.factory.ClientV2),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", false),
		flags.WithPipelineAliases("id", "id"),
		flags.WithPowershellName("Get-UIPlugin"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.application+json", ""),
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

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		ref := c8ystream.NameOrID(in.String("id"))
		return func(ctx context.Context) any {
			return client.UIPlugins.Get(ctx, ref)
		}, nil
	})
}
