package exists

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ywaiter"
	assertFactory "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/assert/factory"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/desiredstate"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type AssertExists struct{}

func (a *AssertExists) GetStateHandler(cmd *cobra.Command, client *c8y.Client) desiredstate.StateDefiner {
	negate, err := cmd.Flags().GetBool("not")
	_ = err
	return &c8ywaiter.InventoryExistance{
		Client: client,
		Negate: negate,
	}
}

func (a *AssertExists) GetValue(v interface{}, input interface{}) []byte {
	if raw, ok := input.([]byte); ok {
		return raw
	}
	return nil
}

type CmdExists struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdExists(f *cmdutil.Factory) *CmdExists {
	ccmd := &CmdExists{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "exists",
		Short: "Assert existance of a managed object",
		Long: heredoc.Doc(`
			Assert that a managed objects exists or not and pass input untouched

			If the assertion is true, then the input value (stdin or an explicit argument value) will be passed untouched to stdout.
			This is useful if you want to filter a list of managed objects by whether they exist or not in the platform, and use the results
			in some downstream command (in the pipeline)

			By default, a failed assertion will not set the exit code to a non-zero value. If you want a non-zero exit code
			in such as case then use the --strict option.
		`),
		Example: heredoc.Doc(`
			$ c8y inventory assert exists --id 1234
			# => 1234 (if the ID exists)
			# => <no response> (if the ID does not exist)
			# Assert the managed object exists
			
			$ echo "1111" | c8y inventory assert exists
			# Pass the piped input only on if the ids exists as a managed object
			
			$ echo -e "1111\n2222" | c8y inventory assert exists --not
			# Only select the managed object ids which do not exist
			
			$ echo 1 | c8y inventory assert exists --strict
			# Return non-zero exit code if a managed object id=1 does not exist
		`),
	}

	cmd.Flags().Bool("not", false, "Negate the match")

	assertExists := &AssertExists{}
	ccmd.SubCommand = subcommand.NewSubCommand(assertFactory.NewAssertCmdFactory(cmd, f, assertExists))

	return ccmd
}
