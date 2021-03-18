package cmd

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/pkg/cmd/auditrecords/create"
	cmdGet "github.com/reubenmiller/go-c8y-cli/pkg/cmd/auditrecords/get"
	cmdList "github.com/reubenmiller/go-c8y-cli/pkg/cmd/auditrecords/list"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdAuditrecords struct {
	*subcommand.SubCommand
}

func NewSubCmdAuditrecords(f *cmdutil.Factory) *SubCmdAuditrecords {
	ccmd := &SubCmdAuditrecords{}

	cmd := &cobra.Command{
		Use:   "auditrecords",
		Short: "Cumulocity auditRecords",
		Long:  `REST endpoint to interact with Cumulocity auditRecords`,
	}

	// Subcommands
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdList.NewListCmd(f).GetCommand())
	cmd.AddCommand(cmdGet.NewGetCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
