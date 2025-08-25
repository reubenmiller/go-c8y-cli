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
func (n *RegisterBasicCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithStaticStringValue("authType", "BASIC"),
		flags.WithStaticStringValue("isAgent", "true"),
		flags.WithStringValue("external-type", "external-type"),
		flags.WithStringValue("iccid", "iccid"),
		flags.WithStringValue("password", "password"),
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
		{CSVHeader: "CREDENTIALS", Properties: WithPasswordOrDefault("password"), Output: WithValue("password")},
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
