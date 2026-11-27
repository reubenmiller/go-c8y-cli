package plugins

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

type CmdList struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	Application string
}

func ListNewCmd(f *cmdutil.Factory) *CmdList {
	ccmd := &CmdList{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed UI plugins",
		Long: heredoc.Doc(`
			List UI plugins which are installed in an application
		`),
		Example: heredoc.Doc(`
			$ c8y ui applications plugins list --application devicemanagement
			List the ui plugins which are installed
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.Application, "application", "", "Application (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithHostedApplication("application", ccmd.factory.ClientV2),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("application", "application", false, "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdList) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	// v2.59.0 sent the requests of this command without a page size
	r.NoDefaultPageSize()
	if err := r.InputFlag("application"); err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		app := c8ystream.NameOrID(in.String("application"))
		if app == "" {
			// As in v2.59.0 no application is not an error: nothing is
			// listed (also nothing under --raw) and no request is sent
			return func(context.Context) any {
				return nil
			}, nil
		}
		return func(ctx context.Context) any {
			// One document per plugin, also with --raw: v2.59.0 wrote each
			// plugin on its own, so no plugin prints nothing (not "[]")
			return output.Seq(client.UIApplicationPlugins.ListPlugins(ctx, app).Documents())
		}, nil
	})
}
