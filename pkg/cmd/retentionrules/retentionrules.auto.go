package cmd

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/retentionrules/create"
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/retentionrules/delete"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/retentionrules/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/retentionrules/list"
	cmdUpdate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/retentionrules/update"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdRetentionrules struct {
	*subcommand.SubCommand
}

func NewSubCmdRetentionrules(f *cmdutil.Factory) *SubCmdRetentionrules {
	ccmd := &SubCmdRetentionrules{}

	cmd := &cobra.Command{
		Use:   "retentionrules",
		Short: "Cumulocity retentionRules",
		Long:  `REST endpoint to interact with Cumulocity retentionRules`,
	}

	// Subcommands
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())
	cmd.AddCommand(cmdDelete.NewDeleteCmd(f).GetCommand())
	cmd.AddCommand(cmdUpdate.NewUpdateCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
