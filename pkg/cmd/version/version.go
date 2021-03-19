package version

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type CmdVersion struct {
	*subcommand.SubCommand
}

func NewCmdVersion(f *cmdutil.Factory) *CmdVersion {
	ccmd := &CmdVersion{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show command version",
		Long:  `Version number of c8y`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Cumulocity command line tool\n%s -- %s\n", f.BuildVersion, f.BuildBranch)
		},
	}

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
