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

// NewListCmd creates a command to Get firmware collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get firmware collection",
		Long:  `Get a collection of firmware packages (managedObjects) based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y firmware list
Get a list of firmware packages
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			handler := c8yquerycmd.NewInventoryQueryRunner(
				cmd,
				args,
				ccmd.factory,
				flags.WithStaticStringValue("dummy", "(type eq 'c8y_Firmware')"),
				flags.WithStringValue("name", "name", "(name eq '%s')"),
				flags.WithStringValue("deviceType", "deviceType", "(c8y_Filter.type eq '%s')"),
				flags.WithStringValue("description", "description", "(description eq '%s')"),
			)
			return handler()
		},
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Firmware name filter")
	cmd.Flags().String("description", "", "Firmware description filter")
	cmd.Flags().String("deviceType", "", "Firmware device type filter")

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
