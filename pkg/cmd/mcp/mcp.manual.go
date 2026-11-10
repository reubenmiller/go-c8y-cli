package mcp

import (
	cmdConnect "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/mcp/connect"
	cmdServe "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/mcp/serve"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdMCP struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdMCP {
	ccmd := &SubCmdMCP{}

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol (MCP) server for AI agents",
		Long:  `Run an MCP server which gives AI agents controlled access to a Cumulocity tenant using go-c8y-cli commands`,
	}

	cmd.AddCommand(cmdServe.NewCmdServe(f).GetCommand())
	cmd.AddCommand(cmdConnect.NewCmdConnect(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
