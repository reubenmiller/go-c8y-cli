package version

import (
	"encoding/json"
	"runtime/debug"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
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

// ReleaseInfo release information
type ReleaseInfo struct {
	Version string `json:"version,omitempty"`
	Branch  string `json:"branch,omitempty"`
}

// RunE execute command
func (n *CmdVersion) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	release := &ReleaseInfo{
		Version: n.factory.BuildVersion,
		Branch:  n.factory.BuildBranch,
	}

	if release.Version == "" {
		info, ok := debug.ReadBuildInfo()
		if ok {
			release.Version = info.Main.Version
			release.Branch = ""

		}
	}

	responseText, err := json.Marshal(release)
	if err != nil {
		return nil
	}
	return n.factory.WriteJSONToConsole(cfg, cmd, "", responseText)
}
