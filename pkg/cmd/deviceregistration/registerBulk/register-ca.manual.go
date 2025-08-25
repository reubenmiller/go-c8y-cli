package registerBulk

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// RegisterCumulocityCACmd command
type RegisterCumulocityCACmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewRegisterCumulocityCACmd creates a command to Register device via the Cumulocity CA
func NewRegisterCumulocityCACmd(f *cmdutil.Factory) *RegisterCumulocityCACmd {
	ccmd := &RegisterCumulocityCACmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "register-ca",
		Short: "Register device with an x509 certificate from the Cumulocity Certificate Authority (private preview feature)",
		Long: heredoc.Doc(`
			Register a device using the Cumulocity Certificate Authority to enable the device to request a device certificate
			securely using EST. This will be supported by thin-edge.io out-of-the-box.

			This feature requires the private preview feature toggle, "certificate-authority"
		`),
		Example: heredoc.Doc(`
			$ c8y deviceregistration register-ca --id "ASDF098SD1J10912UD92JDLCNCU8"
			Register a new device using BASIC authentication and generate a random password (printed on the console)

			$ c8y deviceregistration register-ca --id "ASDF098SD1J10912UD92JDLCNCU8" --one-time-password "RqzwJeTusABlk4)KmtIc"
			Register a new device and provide the one-time-password to be used for enrollment

			$ echo -e "device1\ndevice2" | c8y deviceregistration register-ca --type linux --template "{name: input.value}"
			Register 2 devices, and set the names based on their external id
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
	cmd.Flags().String("one-time-password", "", "One Time Password used for initial enrollment. Leave blank for a randomly generated password")
	cmd.Flags().String("tenant", "", "The ID of the tenant for which the registration is executed (only allowed for the management tenant)")
	cmd.Flags().String("group", "", "The path in the groups hierarchy where the device is added. PATH contains the name of each group separated by /, that is: main_group/sub_group/.../last_sub_group. If a group does not exist, the import creates the group")

	completion.WithOptions(
		cmd,
		completion.WithRootDeviceGroup("group", func() (*c8y.Client, error) { return f.Client() }),
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
func (n *RegisterCumulocityCACmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	c8yclient, err := n.factory.Client()
	if err != nil {
		return err
	}

	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder(true)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithOverrideValue("id", "id"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("type", "type"),
		flags.WithStaticStringValue("authType", "CERTIFICATES"),
		flags.WithStaticStringValue("isAgent", "true"),
		flags.WithStringValue("external-type", "external-type"),
		flags.WithStringValue("iccid", "iccid"),
		flags.WithStringValue("one-time-password", "password"),
		flags.WithStringValue("tenant", "tenant"),
		flags.WithStringValue("group", "group"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithStringValue("id", "id", ""),
	)

	if err != nil {
		return cmderrors.NewUserError(err)
	}

	var iter iterator.Iterator
	if inputIterators.Total > 0 {
		iter = mapbuilder.NewMapBuilderIterator(body)
	} else {
		iter = iterator.NewBoundIterator(mapbuilder.NewMapBuilderIterator(body), 1)
	}

	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}
	commonOptions.DisableResultPropertyDetection()

	mappings := []PayloadMapping{
		{CSVHeader: "ID", Properties: WithValue("id"), Output: WithValue("externalId")},
		{CSVHeader: "AUTH_TYPE", Properties: WithValue("authType"), Output: WithValue("authType")},
		{CSVHeader: "ENROLLMENT_OTP", Properties: WithPasswordOrDefault("password"), Output: WithValue("password")},
		{CSVHeader: "NAME", Properties: WithValue("name", "id"), Output: WithValue("name")},
		{CSVHeader: "TYPE", Properties: WithValue("type"), Output: WithValue("type")},
		{CSVHeader: "IDTYPE", Properties: WithValue("external-type"), Output: WithValue("externalType")},
		{CSVHeader: "ICCID", Properties: WithValue("iccid")},
		{CSVHeader: "TENANT", Properties: WithValue("tenant"), Output: WithValue("tenant")},
		{CSVHeader: "PATH", Properties: WithValue("group")},
		{CSVHeader: "com_cumulocity_model_Agent.active", Properties: WithValue("isAgent")},
	}

	return n.factory.RunWithGenericWorkers(cmd, inputIterators, iter, RunBulkRegistrationJob(cmd, &RegistrationOptions{
		Config:        cfg,
		Client:        c8yclient,
		Factory:       n.factory,
		CommonOptions: commonOptions,
	}, mappings))
}
