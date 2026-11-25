package certificate_authority

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type SubCmdCertificate_authority struct {
	*subcommand.SubCommand
}

func NewSubCommand(f *cmdutil.Factory) *SubCmdCertificate_authority {
	ccmd := &SubCmdCertificate_authority{}

	cmd := &cobra.Command{
		Use:   "certificate-authority",
		Short: "(PREVIEW FEATURE) Cumulocity certificate authority",
		Long: heredoc.Doc(`
			The Cumulocity certificate authority must be first enabled in your tenant
			by the Cumulocity feature toggles. Please contact support if you wish to
			use the preview of this feature.
		`),
	}

	// Subcommands
	cmd.AddCommand(NewGetCmd(f).GetCommand())
	cmd.AddCommand(NewCreateCmd(f).GetCommand())
	cmd.AddCommand(NewUpdateCmd(f).GetCommand())
	cmd.AddCommand(NewDeleteCmd(f).GetCommand())

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}
