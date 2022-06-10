package statistics

import (
	cmdList "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/statistics/list"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdStatistics struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdStatistics {
	ccmd := &SubCmdStatistics{}

	cmd := &cobra.Command{
		Use:   "statistics",
		Short: "Cumulocity device statistics (for a single tenant) statistics",
		Long: `Device statistics are collected for each inventory object with at least one measurement, event or alarm. There are no additional checks if the inventory object is marked as device using the c8y_IsDevice fragment. When the first measurement, event or alarm is created for a specific inventory object, Cumulocity IoT is always considering this as a device and starts counting.

Device statistics are counted with daily and monthly rate. All requests are considered when counting device statistics, no matter which processing mode is used.
`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
