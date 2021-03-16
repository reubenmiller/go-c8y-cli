package cmd

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type versionCmd struct {
	*baseCmd
}

func NewVersionCmd() *versionCmd {
	ccmd := &versionCmd{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show command version",
		Long:  `Version number of c8y`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Cumulocity command line tool\n%s -- %s\n", buildVersion, buildBranch)
		},
	}

	cmdutil.DisableAuthCheck(cmd)
	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
