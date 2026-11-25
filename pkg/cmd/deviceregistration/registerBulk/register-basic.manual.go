package registerBulk

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/devices/registration"
	"github.com/spf13/cobra"
)

// RegisterBasicCmd command
type RegisterBasicCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewRegisterBasicCmd creates a command to Register device
func NewRegisterBasicCmd(f *cmdutil.Factory) *RegisterBasicCmd {
	ccmd := &RegisterBasicCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "register-basic",
		Short: "Register device with username/password credentials",
		Long:  `Register a new device which will authenticate with username/password credentials`,
		Example: heredoc.Doc(`
			$ c8y deviceregistration register-basic --id "ASDF098SD1J10912UD92JDLCNCU8"
			Register a new device using BASIC authentication and generate a random password (printed on the console)

			$ c8y deviceregistration register-basic --id "ASDF098SD1J10912UD92JDLCNCU8" --password "RqzwJeTusABlk4)KmtIc"
			Register a new device using a user specified password

			$ echo -e "device1\ndevice2" | c8y deviceregistration register-basic --type linux --template "{name: input.value}"
			Register 2 devices, and set the names based on their external id (using basic auth)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Device identifier. Max: 1000 characters. E.g. IMEI (required) (accepts pipeline)")
	cmd.Flags().String("name", "", "Device name. Defaults to the external id")
	cmd.Flags().String("type", "thin-edge.io", "Device type")
	cmd.Flags().String("iccid", "", "The ICCID of the device (SIM card number). If the ICCID appears in file, the import adds a fragment c8y_Mobile.iccid.")
	cmd.Flags().String("external-type", "c8y_Serial", "The type of the external ID. If IDTYPE doesn't appear in the file, the default value is used. The default value is c8y_Serial")
	cmd.Flags().String("password", "", "Device password. Leave blank for a randomly generated password")
	cmd.Flags().String("tenant", "", "The ID of the tenant for which the registration is executed (only allowed for the management tenant)")
	cmd.Flags().String("group", "", "The path in the groups hierarchy where the device is added. PATH contains the name of each group separated by /, that is: main_group/sub_group/.../last_sub_group. If a group does not exist, the import creates the group")

	completion.WithOptions(
		cmd,
		completion.WithRootDeviceGroup("group", f.ClientV2),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true, "externalId", "name", "id"),

		// Enable confirmation prompts
		flags.WithSemanticMethod("POST"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *RegisterBasicCmd) RunE(cmd *cobra.Command, args []string) error {
	return run(cmd, n.factory, registration.AuthTypeBasic, credentials,
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("type", "type"),
		flags.WithStaticStringValue("isAgent", "true"),
		flags.WithStringValue("external-type", "external-type"),
		flags.WithStringValue("iccid", "iccid"),
		flags.WithStringValue("password", "password"),
		flags.WithStringValue("tenant", "tenant"),
		flags.WithStringValue("group", "group"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
	)
}
