package wait

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ywait"
	assertFactory "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/assert/factory"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

type CmdWait struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	ExpectedStatus []string
}

func NewCmdWait(f *cmdutil.Factory) *CmdWait {
	ccmd := &CmdWait{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Wait for operation",
		Long:  `Wait for an operation to complete`,
		Example: heredoc.Doc(`
			$ c8y operations wait --id 1234
			# Wait for the operation to be set to SUCCESSFUL

			$ c8y operations wait --id 1234 --duration 1m
			# Wait for the operation to be set to SUCCESSFUL and give up after 1 minute

			$ c8y operations list --device 1111 | c8y operations wait --status "FAILED" --status "SUCCESSFUL"
			# Wait for operation to be set to either FAILED or SUCCESSFUL
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Operation id (required) (accepts pipeline)")
	cmd.Flags().StringSliceVar(&ccmd.ExpectedStatus, "status", []string{"SUCCESSFUL"}, "Status to wait for. If multiple values are given, then it will be applied as an OR operation")
	cmd.Flags().String("duration", "30s", "Timeout duration. i.e. 30s or 1m (1 minute)")
	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
	)

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("status", "PENDING", "EXECUTING", "SUCCESSFUL", "FAILED"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdWait) RunE(cmd *cobra.Command, args []string) error {
	return assertFactory.RunWait(cmd, n.factory, "id", func(client *apiv2.Client, in *c8ystream.Resolver, ref string) c8ywait.Check {
		return c8ywait.OperationStatus(client, ref, n.ExpectedStatus)
	})
}
