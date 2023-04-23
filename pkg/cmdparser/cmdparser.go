package cmdparser

import (
	"fmt"
	"io"
	"strconv"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/integration/models"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Command struct {
	Name string `yaml:"name"`
}

func ParseCommand(r io.Reader, factory *cmdutil.Factory) (*cobra.Command, error) {
	spec := &models.Specification{}
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, spec)
	if err != nil {
		return nil, err
	}

	if spec.Group.Skip {
		return nil, nil
	}

	cmd := &cobra.Command{
		Use:   spec.Group.Name,
		Short: spec.Group.Description,
		Long:  spec.Group.DescriptionLong,
	}

	for _, item := range spec.Commands {
		if item.ShouldIgnore() {
			continue
		}

		subcmd := NewCommandWithOptions(&cobra.Command{
			Use:     item.Name,
			Short:   item.Description,
			Long:    item.GetDescriptionLong(),
			Example: item.GetExamples(),
			Hidden:  item.IsHidden(),
		}, item)

		flagNames := make(map[string]interface{})
		for _, param := range item.GetAllParameters() {
			// Ignore duplicated flags
			if _, ok := flagNames[param.Name]; ok {
				continue
			}
			flagNames[param.Name] = 0

			if err := AddFlag(subcmd, &param, factory); err != nil {
				return nil, err
			}

			if param.AcceptsPipeline() {
				subcmd.Runtime = append(subcmd.Runtime, flags.WithExtendedPipelineSupport(param.Name, param.GetTargetProperty(), param.IsRequired(), param.PipelineAliases...))
				subcmd.Runtime = append(subcmd.Runtime, flags.WithPipelineAliases(param.Name, param.PipelineAliases...))
			}

			if len(param.ValidationSet) > 0 {
				subcmd.Completion = append(subcmd.Completion, completion.WithValidateSet(param.Name, param.ValidationSet...))
			}

			// Add completions
			if comp := GetCompletionOptions(subcmd, &param, factory); comp != nil {
				subcmd.Completion = append(subcmd.Completion, comp)
			}
		}

		// Misc. options
		// TODO: Check if pipeline should always be added if no pipeline argument is supported
		// flags.WithExtendedPipelineSupport("", "", false),

		if item.CollectionProperty != "" {
			subcmd.Runtime = append(subcmd.Runtime, flags.WithCollectionProperty(item.CollectionProperty))
		}

		if item.SemanticMethod != "" {
			subcmd.Runtime = append(subcmd.Runtime, flags.WithSemanticMethod(item.SemanticMethod))
		}

		if item.IsDeprecated() {
			subcmd.Runtime = append(subcmd.Runtime, flags.WithDeprecationNotice(item.Deprecated))
		}

		if item.SupportsProcessingMode() {
			subcmd.Runtime = append(subcmd.Runtime, flags.WithProcessingMode())
		}

		// Add template/data support by default
		if item.SupportsTemplates() {
			subcmd.Runtime = append(
				subcmd.Runtime,
				flags.WithData(),
				factory.WithTemplateFlag(subcmd.Command),
			)
		}

		cmd.AddCommand(subcmd.NewRuntimeCommand(factory).SubCommand.GetCommand())
	}

	return cmd, nil
}

func MapCommandAPI(cmd *CmdOptions, param *models.Parameter, typeName string) {
	switch typeName {
	case "string":
		StringArg(cmd, param)
	case "integer":
		if v, err := strconv.ParseInt(param.Default, 10, 64); err == nil {
			cmd.Command.Flags().Int64(param.Name, v, param.Description)
		}
		cmd.Command.Flags().Int64(param.Name, 0, param.Description)
	}
}

func GetCompletionOptions(cmd *CmdOptions, p *models.Parameter, factory *cmdutil.Factory) completion.Option {
	switch p.Type {
	case "application", "applicationname":
		return completion.WithApplication(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "hostedapplication":
		return completion.WithHostedApplication(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "microservice":
		return completion.WithMicroservice(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "microservicename":
		return completion.WithMicroservice(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "microserviceinstance":
		return completion.WithMicroserviceInstance(p.Name, "id", func() (*c8y.Client, error) { return factory.Client() })
	case "role":
		return completion.WithUserRole(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]devicerequest":
		return completion.WithDeviceRegistrationRequest(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]user", "[]userself":
		return completion.WithUser(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]usergroup":
		return completion.WithUserGroup(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]devicegroup":
		return completion.WithDeviceGroup(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]smartgroup":
		return completion.WithSmartGroup(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]tenant":
		return completion.WithTenantID(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "tenantname":
		return completion.WithTenantID(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]device":
		return completion.WithDevice(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]agent":
		return completion.WithAgent(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]software", "softwareName":
		return completion.WithSoftware(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]softwareversion", "softwareversionName":
		if len(p.DependsOn) > 0 {
			return completion.WithSoftwareVersion(p.Name, p.DependsOn[0], func() (*c8y.Client, error) { return factory.Client() })
		}
	case "[]firmware(name)":
		return completion.WithFirmware(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]firmwareversion", "firmwareVersionName":
		if len(p.DependsOn) > 0 {
			return completion.WithFirmwareVersion(p.Name, p.DependsOn[0], func() (*c8y.Client, error) { return factory.Client() })
		}
	case "[]firmwarepatch", "firmwarepatchName":
		if len(p.DependsOn) > 0 {
			return completion.WithFirmwarePatch(p.Name, p.DependsOn[0], func() (*c8y.Client, error) { return factory.Client() })
		}
	case "[]configuration":
		return completion.WithConfiguration(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]deviceprofile":
		return completion.WithDeviceProfile(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "[]deviceservice":
		if len(p.DependsOn) > 0 {
			return completion.WithDeviceService(p.Name, p.DependsOn[0], func() (*c8y.Client, error) { return factory.Client() })
		}
	case "[]certificate":
		return completion.WithDeviceCertificate(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "subscriptionName":
		return completion.WithNotification2SubscriptionName(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	case "subscriptionId":
		return completion.WithNotification2SubscriptionId(p.Name, func() (*c8y.Client, error) { return factory.Client() })
	}
	return nil
}

func AddFlag(cmd *CmdOptions, p *models.Parameter, factory *cmdutil.Factory) error {
	switch p.Type {
	case "string", "stringStatic", "json_custom", "directory", "softwareName", "softwareversionName", "firmwareName", "firmwareversionName", "firmwarepatchName", "binaryUploadURL", "inventoryChildType", "subscriptionName", "subscriptionId", "file", "attachment", "fileContents", "certificatefile":
		cmd.Command.Flags().StringP(p.Name, p.ShortName, p.Default, p.Description)

	case "json":
		// Ignore, as it is add by default to all PUT and POST requests

	case "datefrom", "dateto", "datetime", "date":
		cmd.Command.Flags().StringP(p.Name, p.ShortName, p.Default, p.GetDescription())
		p.PipelineAliases = append(p.PipelineAliases, "time", "creationTime", "creationTime", "lastUpdated")

	case "source":
		cmd.Command.Flags().StringP(p.Name, p.ShortName, p.Default, p.GetDescription())
		p.PipelineAliases = append(p.PipelineAliases, "id", "source.id", "managedObject.id", "deviceId")

	case "[]string", "stringcsv", "[]devicerequest", "[]software", "[]softwareversion", "[]firmware", "[]firmwareversion", "[]firmwarepatch", "[]configuration", "[]deviceprofile", "[]deviceservice", "[]id", "[]user", "[]userself", "[]certificate":
		cmd.Command.Flags().StringSliceP(p.Name, p.ShortName, []string{p.Default}, p.GetDescription())

	case "[]device", "[]agent":
		cmd.Command.Flags().StringSliceP(p.Name, p.ShortName, []string{p.Default}, p.GetDescription())
		p.PipelineAliases = append(p.PipelineAliases, "deviceId", "source.id", "managedObject.id", "id")

	case "[]devicegroup":
		cmd.Command.Flags().StringSliceP(p.Name, p.ShortName, []string{p.Default}, p.GetDescription())
		p.PipelineAliases = append(p.PipelineAliases, "source.id", "managedObject.id", "id")

	case "[]smartgroup":
		cmd.Command.Flags().StringSliceP(p.Name, p.ShortName, []string{p.Default}, p.GetDescription())
		p.PipelineAliases = append(p.PipelineAliases, "managedObject.id")

	case "[]roleself":
		cmd.Command.Flags().StringSliceP(p.Name, p.ShortName, []string{p.Default}, p.GetDescription())
		p.PipelineAliases = append(p.PipelineAliases, "self", "id")

	case "[]role", "[]usergroup":
		cmd.Command.Flags().StringSliceP(p.Name, p.ShortName, []string{p.Default}, p.GetDescription())
		p.PipelineAliases = append(p.PipelineAliases, "id")

	case "application", "applicationname", "hostedapplication", "microservice", "microserviceinstance":
		cmd.Command.Flags().StringP(p.Name, p.ShortName, p.Default, p.GetDescription())
		p.PipelineAliases = append(p.PipelineAliases, "id")

	case "microservicename":
		cmd.Command.Flags().StringP(p.Name, p.ShortName, p.Default, p.GetDescription())
		p.PipelineAliases = append(p.PipelineAliases, "name")

	case "tenant", "tenantname":
		cmd.Command.Flags().StringP(p.Name, p.ShortName, p.Default, p.GetDescription())
		p.PipelineAliases = append(p.PipelineAliases, "tenant", "owner.tenant.id")

	case "integer":
		defaultValue, err := strconv.ParseInt(p.Default, 0, 64)
		if err != nil {
			defaultValue = 0
		}
		cmd.Command.Flags().IntP(p.Name, p.ShortName, int(defaultValue), p.GetDescription())

	case "float":
		defaultValue, err := strconv.ParseFloat(p.Default, 32)
		if err != nil {
			defaultValue = 0
		}
		cmd.Command.Flags().Float32P(p.Name, p.ShortName, float32(defaultValue), p.GetDescription())

	case "boolean", "booleanDefault", "optional_fragment":
		defaultValue, err := strconv.ParseBool(p.Default)
		if err != nil {
			defaultValue = false
		}
		cmd.Command.Flags().BoolP(p.Name, p.ShortName, defaultValue, p.GetDescription())

	default:
		return fmt.Errorf("unknown flag type. name=%s, type=%s", p.Name, p.Type)
	}

	if p.IsRequired() && !p.AcceptsPipeline() {
		cmd.Command.MarkFlagRequired(p.Name)
	}

	if p.IsHidden() && !p.AcceptsPipeline() {
		cmd.Command.Flags().MarkHidden(p.Name)
	}

	return nil
}

func GetOption(cmd *CmdOptions, p *models.Parameter, factory *cmdutil.Factory, cfg *config.Config, client *c8y.Client, args []string) []flags.GetOption {
	targetProp := p.GetTargetProperty()

	opts := []flags.GetOption{}
	switch p.Type {
	case "file":
		opts = append(opts, flags.WithFormDataFileAndInfoWithTemplateSupport(cmdutil.NewTemplateResolver(factory, cfg), p.Name, "data")...)
	case "fileContents":
		opts = append(opts, flags.WithFilePath(p.Name, targetProp, p.Value))
	case "attachment":
		opts = append(opts, flags.WithFormDataFile(p.Name, "data")...)

	case "boolean":
		opts = append(opts, flags.WithBoolValue(p.Name, targetProp, p.Value))
	case "booleanDefault":
		opts = append(opts, flags.WithDefaultBoolValue(p.Name, targetProp, p.Value))
	case "optional_fragment":
		opts = append(opts, flags.WithOptionalFragment(p.Name, targetProp, p.Value))

	case "datetime":
		if p.TargetType == models.ParamPath || p.TargetType == models.ParamQueryParameter {
			opts = append(opts, flags.WithEncodedRelativeTimestamp(p.Name, targetProp, p.Format))
		} else {
			opts = append(opts, flags.WithRelativeTimestamp(p.Name, targetProp, p.Format))
		}
	case "date":
		opts = append(opts, flags.WithRelativeDate(false, p.Name, targetProp, p.Format))

	case "[]string":
		opts = append(opts, flags.WithStringSliceValues(p.Name, targetProp, p.Value))
	case "[]stringcsv":
		opts = append(opts, flags.WithStringSliceCSV(p.Name, targetProp, p.Value))

	case "inventoryChildType":
		opts = append(opts, flags.WithInventoryChildType(p.Name, targetProp, p.Format))

	case "string", "source", "tenantname", "subscriptionName", "subscriptionId", "applicationname", "microserviceinstance", "microservicename", "softwareName", "softwareversionName", "firmwareName", "firmwareversionName", "firmwarepatchName":
		opts = append(opts, flags.WithStringValue(p.Name, targetProp, p.Format))

	case "stringStatic":
		opts = append(opts, flags.WithStaticStringValue(p.Name, p.Value))
	case "integer":
		opts = append(opts, flags.WithIntValue(p.Name, targetProp, p.Format))
	case "float":
		opts = append(opts, flags.WithFloatValue(p.Name, targetProp, p.Format))

	case "json_custom":
		opts = append(opts, flags.WithDataValue(p.Name, targetProp, p.Format))
	case "binaryUploadURL":
		opts = append(opts, c8ybinary.WithBinaryUploadURL(client, factory.IOStreams.ProgressIndicator(), p.Name, targetProp, p.Format))
	case "json":
		// don't do anything because it should be manually set)

	case "tenant":
		opts = append(opts, flags.WithStringDefaultValue(client.TenantName, p.Name, targetProp, p.Format))

	case "[]id", "[]devicerequest":
		opts = append(opts, c8yfetcher.WithIDSlice(args, p.Name, targetProp, p.Format))

	case "[]software":
		opts = append(opts, c8yfetcher.WithSoftwareByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

	case "softwareDetails":
		opts = append(opts, c8yfetcher.WithSoftwareVersionData(client, "software", "version", "url", args, "", targetProp, p.Format))

	case "configurationDetails":
		opts = append(opts, c8yfetcher.WithConfigurationFileData(client, "configuration", "configurationType", "url", args, "", targetProp, p.Format))

	case "[]softwareversion":
		opts = append(opts, c8yfetcher.WithSoftwareVersionByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

	case "[]deviceservice":
		opts = append(opts, c8yfetcher.WithDeviceServiceByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

	case "certificatefile":
		opts = append(opts, flags.WithCertificateFile(p.Name, targetProp))
	case "[]certificate":
		opts = append(opts, c8yfetcher.WithCertificateByNameFirstMatch(client, args, p.Name, targetProp))

	case "[]firmware":
		opts = append(opts, c8yfetcher.WithFirmwareByNameFirstMatch(client, args, p.Name, targetProp, p.Format))
	case "[]firmwareversion":
		opts = append(opts, c8yfetcher.WithFirmwareVersionByNameFirstMatch(client, args, p.Name, targetProp, p.Format))
	case "firmwareDetails":
		opts = append(opts, c8yfetcher.WithFirmwareVersionData(client, "firmware", "version", "url", args, "", targetProp))
	case "[]firmwarepatch":
		opts = append(opts, c8yfetcher.WithFirmwarePatchByNameFirstMatch(client, args, p.Name, targetProp))

	case "[]configuration":
		opts = append(opts, c8yfetcher.WithConfigurationByNameFirstMatch(client, args, p.Name, targetProp))

	case "[]deviceprofile":
		opts = append(opts, c8yfetcher.WithDeviceProfileByNameFirstMatch(client, args, p.Name, targetProp))

	case "[]device":
		opts = append(opts, c8yfetcher.WithDeviceByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

	case "[]agent":
		opts = append(opts, c8yfetcher.WithAgentByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

	case "[]devicegroup":
		opts = append(opts, c8yfetcher.WithDeviceGroupByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

	case "[]smartgroup":
		opts = append(opts, c8yfetcher.WithSmartGroupByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

	case "[]user":
		opts = append(opts, c8yfetcher.WithUserByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

	case "[]userself":
		opts = append(opts, c8yfetcher.WithUserSelfByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

	case "[]roleself":
		opts = append(opts, c8yfetcher.WithRoleSelfByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

	case "[]role":
		opts = append(opts, c8yfetcher.WithRoleByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

	case "[]usergroup":
		opts = append(opts, c8yfetcher.WithUserGroupByNameFirstMatch(client, args, p.Name, targetProp, p.Format))
	}

	return opts
}

type CmdOptions struct {
	Spec           models.Command
	Completion     []completion.Option
	Command        *cobra.Command
	Runtime        []flags.Option
	Header         []flags.GetOption
	QueryParameter []flags.GetOption
	FormData       []flags.GetOption
	Body           BodyOptions
	Path           PathOptions
}

type BodyOptions struct {
	Options    []flags.GetOption
	IsBinary   bool
	Initialize bool
}

func (c *CmdOptions) NewRuntimeCommand(f *cmdutil.Factory) *RuntimeCmd {
	return NewRuntimeCmd(f, c)
}

func NewCommandWithOptions(cmd *cobra.Command, endpoint models.Command) *CmdOptions {
	return &CmdOptions{
		Spec:           endpoint,
		Command:        cmd,
		Runtime:        make([]flags.Option, 0),
		Completion:     make([]completion.Option, 0),
		Header:         make([]flags.GetOption, 0),
		QueryParameter: make([]flags.GetOption, 0),
		FormData:       make([]flags.GetOption, 0),
		Body: BodyOptions{
			Options: make([]flags.GetOption, 0),
		},
		Path: PathOptions{
			Options: make([]flags.GetOption, 0),
		},
	}
}

type PathOptions struct {
	Template string
	Options  []flags.GetOption
}

func StringArg(cmd *CmdOptions, param *models.Parameter) {
	cmd.Command.Flags().String(param.Name, param.Default, param.Description)
}
