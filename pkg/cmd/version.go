package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type versionCmd struct {
	*baseCmd
}

func newVersionCmd() *versionCmd {
	ccmd := &versionCmd{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of c8y",
		Long:  `All software has versions. This is c8y`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Cumulocity command line tool\n%s -- %s\n", buildVersion, buildBranch)
		},
	}

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}
