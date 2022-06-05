package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yquerycmd"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get software collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get software collection",
		Long:  `Get a collection of software packages (managedObjects) based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y software list
Get a list of software packages

$ c8y software list --name "python3*"
Get a list of software packages starting with "python3"
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			handler := c8yquerycmd.NewInventoryQueryRunner(
				cmd,
				args,
				ccmd.factory,
				flags.WithStaticStringValue("dummy", "(type eq 'c8y_Software')"),
				flags.WithStringValue("name", "name", "(name eq '%s')"),
				flags.WithStringValue("deviceType", "deviceType", "(c8y_Filter.type eq '%s')"),
				flags.WithStringValue("description", "description", "(description eq '%s')"),
			)
			return handler()
		},
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Software name filter")
	cmd.Flags().String("description", "", "Software description filter")
	cmd.Flags().String("deviceType", "", "Software device type filter")

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithCommonCumulocityQueryOptions(),
		flags.WithExtendedPipelineSupport("query", "query", false),
		flags.WithCollectionProperty("managedObjects"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
