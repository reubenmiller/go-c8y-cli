package tokens

import (
	cmdCreate "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/notification2/tokens/create"
	cmdUnsubscribe "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/notification2/tokens/unsubscribe"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdTokens struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdTokens {
	ccmd := &SubCmdTokens{}

	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "Cumulocity notification2 tokens",
		Long: `In order to receive subscribed notifications, a consumer application or microservice
must obtain an authorization token that provides proof that the holder is allowed to
receive subscribed notifications.
`,
	}

	// Subcommands
	cmd.AddCommand(cmdCreate.NewCreateCmd(f).GetCommand())
	cmd.AddCommand(cmdUnsubscribe.NewUnsubscribeCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
