package exists

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ywait"
	assertFactory "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/assert/factory"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/spf13/cobra"
)


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
		Short: "Assert existence of a tenant",
		Long: heredoc.Doc(`
			Assert that a tenant exists or not and pass input untouched

			If the assertion is true, then the input value (stdin or an explicit argument value) will be passed untouched to stdout.
			This is useful if you want to filter a list of tenants by whether they exist or not in the platform, and use the results
			in some downstream command (in the pipeline)

			By default, a failed assertion will not set the exit code to a non-zero value. If you want a non-zero exit code
			in such as case then use the --strict option.
		`),
		Example: heredoc.Doc(`
			$ c8y tenants assert exists --tenant t12345
			# => t12345 (if the tenant exists)
			# => <no response> (if the tenant does not exist)
			# Assert the tenant exists

			$ echo "t12345" | c8y tenants assert exists
			# Pass the piped input only on if the tenant exists

			$ echo -e "t11111\nt22222" | c8y tenants assert exists --not
			# Only select the tenant ids which do not exist

			$ echo t12345 | c8y tenants assert exists --strict
			# Return non-zero exit code if tenant t12345 does not exist
		`),
	}

	cmd.Flags().Bool("not", false, "Negate the match")

	ccmd.SubCommand = subcommand.NewSubCommand(assertFactory.New(cmd, f, assertFactory.Assert{
		Target: assertFactory.Tenant,
		Check: func(client *apiv2.Client, in *c8ystream.Resolver, ref string) c8ywait.Check {
			return c8ywait.TenantExists(client, ref, in.Bool("not"))
		},
	}))

	return ccmd
}
