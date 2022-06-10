package fragments

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ywaiter"
	assertFactory "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/assert/factory"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/desiredstate"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type AssertFragment struct{}

func (a *AssertFragment) GetStateHandler(cmd *cobra.Command, client *c8y.Client) desiredstate.StateDefiner {
	fragments, _ := cmd.Flags().GetStringSlice("fragments")
	return &c8ywaiter.InventoryState{
		Client:    client,
		Fragments: fragments,
	}
}

func (a *AssertFragment) GetValue(v interface{}, input interface{}) []byte {
	if mo, ok := v.(*c8y.ManagedObject); ok {
		return []byte(mo.Item.Raw)
	}
	return nil
}

type CmdFragments struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdFragments(f *cmdutil.Factory) *CmdFragments {
	ccmd := &CmdFragments{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "fragments",
		Short: "Assert fragments on a device",
		Long:  `Assert fragments on a device by polling until a condition is met or a timeout is reached`,
		Example: heredoc.Doc(`
			$ c8y devices assert fragments --id 1234 --fragment "c8y_Mobile.iccd"
			# Fragments for the managed object to have a non-null c8y_Mobile.iccd fragment

			$ c8y devices assert fragments --id 1234 --fragment '!c8y_Mobile'
			# Fragments for the managed object to not have a c8y_Mobile fragment

			$ c8y devices assert fragments --id 1234 --fragment 'name=^\d+-\w+$'
			# Fragments for the managed object name fragment to match the regular expression '^\d+-\w+'

			$ c8y devices assert fragments --id 1234 --fragment 'name=^$' --fragment c8y_IsDevice
			# Fragments for the managed object name fragment to match the regular expression '^\d+-\w+'

			$ c8y devices assert fragments --id 1234 --duration 1m --interval 10s
			# Fragments for the operation to be set to SUCCESSFUL and give up after 1 minute

			$ c8y devices list --device 1111 | c8y operations fragments --status "FAILED" --status "SUCCESSFUL"
			# Fragments for operation to be set to either FAILED or SUCCESSFUL
		`),
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("fragments", nil, "Fragments to fragments for. If multiple values are given, then it will be applied as an OR operation")

	completion.WithOptions(
		cmd,
	)

	assertFragment := &AssertFragment{}
	ccmd.SubCommand = subcommand.NewSubCommand(assertFactory.NewAssertDeviceCmdFactory(cmd, f, assertFragment))

	return ccmd
}
