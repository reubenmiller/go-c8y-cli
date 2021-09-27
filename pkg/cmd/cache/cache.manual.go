package cache

import (
	cmdDelete "github.com/reubenmiller/go-c8y-cli/pkg/cmd/cache/delete"
	cmdRenew "github.com/reubenmiller/go-c8y-cli/pkg/cmd/cache/renew"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdCache struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdCache {
	ccmd := &SubCmdCache{}

	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Local cache management",
		Long:  `Commands to manage the local cache, i.e. to cleanup the cache or refresh it`,
	}

	// Subcommands
	cmd.AddCommand(cmdDelete.NewCmdDelete(f).GetCommand())
	cmd.AddCommand(cmdRenew.NewCmdRenew(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
