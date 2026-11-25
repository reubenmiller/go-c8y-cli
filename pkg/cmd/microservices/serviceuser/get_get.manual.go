package serviceuser

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdGet struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdGet(f *cmdutil.Factory) *CmdGet {
	ccmd := &CmdGet{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get microservice service user",
		Long: `Get the service user associated to a microservice
`,
		Example: heredoc.Doc(`
$ c8y microservices serviceuser get --id 12345
Get application service user by app id

$ c8y microservices serviceuser get --id myapp
Get application service user by app name
        `),
		PreRunE: nil,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Microservice id (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithMicroservice("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true, "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdGet) RunE(cmd *cobra.Command, args []string) error {
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
			return client.Microservices.ListServiceUsers(ctx, ref)
		}, nil
	})
}
