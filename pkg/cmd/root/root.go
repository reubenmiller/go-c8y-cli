package root

import (
	"github.com/MakeNowJust/heredoc"
	aliasCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alias"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdRoot(f *cmdutil.Factory, version, buildDate string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "c8y2",
		Short: "Cumulocity command line interface",
		Long:  `A command line interface to interact with Cumulocity REST API. Ideal for quick prototyping, exploring the REST API and for Platform maintainers/power users`,

		SilenceErrors: true,
		SilenceUsage:  true,
		Example: heredoc.Doc(`
			$ c8y devices list
			$ c8y devices list --type "myDevice" | c8y devices update --data "myValue=1"
			$ c8y operations list --device myDeviceName
		`),
		Annotations: map[string]string{
			"help:feedback": heredoc.Doc(`
				Open an issue using 'c8y issue create -R github.com/cli/cli'
			`),
		},
	}

	cmd.SetOut(f.IOStreams.Out)
	cmd.SetErr(f.IOStreams.ErrOut)

	// Child commands
	cmd.AddCommand(aliasCmd.NewCmdAlias(f))

	return cmd
}
