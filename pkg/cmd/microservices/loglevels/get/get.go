// v2-based microservice log-level get: a run-once command (name and loggerName
// are required, not piped) that fetches a single logger's configured level via
// Microservices.Loggers.Get.
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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get log level of microservice
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get log level of microservice",
		Long: `Get configured log level for a package (incl. sub-packages), or a specific class.
(This only works for Spring Boot microservices based on Cumulocity Java Microservice SDK)
`,
		Example: heredoc.Doc(`
$ c8y microservices loglevels get --name my-microservice --loggerName org.example
Get log level of microservice for a package

$ c8y microservices loglevels get --name my-microservice --loggerName org.example.microservice.ClassName
Get log level of microservice for a specific class
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Microservice name (required)")
	cmd.Flags().String("loggerName", "", "Name of the logger: Qualified name of package or class (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("loggerName")

	completion.WithOptions(
		cmd,
		completion.WithMicroservice("name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithMicroserviceLoggers("loggerName", "name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithPowershellName("Get-MicroserviceLogLevel"),
		flags.WithOutputType("application/json", ""),
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
		name := in.String("name")
		loggerName := in.String("loggerName")
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.Microservices.Loggers.Get(ctx, name, loggerName))
		}, nil
	})
}
