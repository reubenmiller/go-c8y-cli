// v2-based microservice log-level list: a run-once command (the name flag is
// required, not piped) that lists the configured loggers of a microservice via
// Microservices.Loggers.List.
package list

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to List log levels of microservice
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List log levels of microservice",
		Long: `List all log levels of microservice.
(This only works for Spring Boot microservices based on Cumulocity Java Microservice SDK)
`,
		Example: heredoc.Doc(`
$ c8y microservices loglevels list --name my-microservice
List log levels of microservice
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Microservice name (required)")
	_ = cmd.MarkFlagRequired("name")

	completion.WithOptions(
		cmd,
		completion.WithMicroservice("name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty("loggers"),
		flags.WithPowershellName("Get-MicroserviceLogLevelCollection"),
		flags.WithOutputType("application/json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		name := in.String("name")
		return func(ctx context.Context) output.Seq {
			return output.FromIterator(client.Microservices.Loggers.List(ctx, name).Items())
		}, nil
	})
}
