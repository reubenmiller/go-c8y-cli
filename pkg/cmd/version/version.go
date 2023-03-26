package version

import (
	"runtime/debug"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type CmdVersion struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdVersion(f *cmdutil.Factory) *CmdVersion {
	ccmd := &CmdVersion{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show command version",
		Long:  `Version number of c8y`,
		RunE:  ccmd.RunE,
	}

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE execute command
func (n *CmdVersion) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	release := map[string]interface{}{
		"version": n.factory.BuildVersion,
		"branch":  n.factory.BuildBranch,
	}

	if n.factory.BuildVersion == "" {
		info, ok := debug.ReadBuildInfo()
		if ok {
			release["version"] = info.Main.Version
			release["branch"] = "(unknown)"
		}
	}

	return n.factory.WriteJSONToConsole(cfg, cmd, "", release)
}
