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

// RegisterExternalCACmd command
type RegisterExternalCACmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewRegisterExternalCACmd creates a command to Register device
func NewRegisterExternalCACmd(f *cmdutil.Factory) *RegisterExternalCACmd {
	ccmd := &RegisterExternalCACmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "register-external-ca",
		Short: "Register device with an x509 certificate from an external certificate authority",
		Long: heredoc.Doc(`
			The certificate should already be issued to the device by the external CA and the CA certificate
			should be added to Cumulocity as a Trusted Certificate.

			This is required when the Trusted Certificate has the autoRegistration setting disabled, so it will
			require a device to be registered with the platform before the certificate is accepted by Cumulocity
		`),
		Example: heredoc.Doc(`
			$ c8y deviceregistration register-external-ca --id "ASDF098SD1J10912UD92JDLCNCU8"
			Register a new device which which is signed by an external ca

			$ echo -e "device1\ndevice2" | c8y deviceregistration register-external-ca --type linux --template "{name: input.value}"
			Register 2 devices, and set the names based on their external id (using CERTIFICATES auth)
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
func (n *RegisterExternalCACmd) RunE(cmd *cobra.Command, args []string) error {
	return run(cmd, n.factory, registration.AuthTypeCertificates, noSecret,
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("type", "type"),
		flags.WithStaticStringValue("isAgent", "true"),
		flags.WithStringValue("external-type", "external-type"),
		flags.WithStringValue("iccid", "iccid"),
		flags.WithStringValue("tenant", "tenant"),
		flags.WithStringValue("group", "group"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
	)
}
