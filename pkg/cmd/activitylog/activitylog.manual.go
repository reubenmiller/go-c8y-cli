package activitylog

import (
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/activitylog/list"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdActivityLog struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdActivityLog {
	ccmd := &SubCmdActivityLog{}

	cmd := &cobra.Command{
		Use:   "activitylog",
		Short: "Activity log commands",
		Long:  `Get information about the activity log`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewCmdList(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
