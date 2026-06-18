// v2-based microservice log-level set: a run-once command (name/loggerName are
// required, not piped) that sets a logger's configured level via
// Microservices.Loggers.Set (POST {configuredLevel: <logLevel>}).
package set

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

// SetCmd command
type SetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewSetCmd creates a command to Set log level of microservice
func NewSetCmd(f *cmdutil.Factory) *SetCmd {
	ccmd := &SetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set log level of microservice",
		Long: `Set configured log level for a package (incl. sub-packages), or a specific class.
(This only works for Spring Boot microservices based on Cumulocity Java Microservice SDK)
`,
		Example: heredoc.Doc(`
$ c8y microservices loglevels set --name my-microservice --loggerName org.example --logLevel DEBUG
Set log level of microservice for a package

$ c8y microservices loglevels set --name my-microservice --loggerName org.example.microservice.ClassName --logLevel TRACE
Set log level of microservice for a specific class
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Microservice name (required)")
	cmd.Flags().String("loggerName", "", "Name of the logger: Qualified name of package or class (required)")
	cmd.Flags().String("logLevel", "", "Log level: TRACE | DEBUG | INFO | WARN | ERROR | OFF (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("loggerName")
	_ = cmd.MarkFlagRequired("logLevel")

	completion.WithOptions(
		cmd,
		completion.WithMicroservice("name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithMicroserviceLoggers("loggerName", "name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("logLevel", "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "OFF"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithPowershellName("Set-MicroserviceLogLevel"),
		flags.WithOutputType("application/json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *SetCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("logLevel", "configuredLevel"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("configuredLevel"),
	)
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
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.MicroserviceLogger] {
				return client.Microservices.Loggers.Set(ctx, name, loggerName, body)
			})
		}, nil
	})
}
