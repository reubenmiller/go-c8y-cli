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

var (
	PresetQueryInventory         = "query-inventory"
	PresetGetIdentity            = "get-identity"
	PresetQueryInventoryChildren = "query-inventory-children"
)

type Command struct {
	Name string `yaml:"name"`
}

func ParseCommand(r io.Reader, factory *cmdutil.Factory, rootCmd *cobra.Command) (*cobra.Command, error) {
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

	// commands
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

		// Use a preset to populate some common flags/options
		if item.HasPreset() {
			AddPredefinedGroupsFlags(subcmd, factory, item.Preset)
		}

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
			subcmd.Completion = AppendCompletionOptions(subcmd.Completion, subcmd, &param, factory)
		}

		// Misc. options
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

func AppendCompletionOptions(opts []completion.Option, cmd *CmdOptions, p *models.Parameter, factory *cmdutil.Factory) []completion.Option {
	if len(p.ValidationSet) > 0 {
		opts = append(opts, completion.WithValidateSet(p.Name, p.ValidationSet...))
	}

	if comp := GetCompletionOptions(cmd, p, factory); comp != nil {
		opts = append(opts, comp)
	}
	return opts
}

func GetCompletionOptions(cmd *CmdOptions, p *models.Parameter, factory *cmdutil.Factory) completion.Option {
	// External flags
	switch p.Completion.Type {
	case "external":
		return completion.WithExternalCompletion(p.Name, p.Completion.Command)
	}

	// Internal flags
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
	case "[]role", "[]roleself":
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
	existingFlag := cmd.Command.Flags().Lookup(p.Name)
	if existingFlag != nil {
		// TODO: Update the existing flag rather than ignoring it
		// TODO: Should an error be returned?
		return nil
	}
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

	case "[]string", "[]stringcsv", "[]devicerequest", "[]software", "[]softwareversion", "[]firmware", "[]firmwareversion", "[]firmwarepatch", "[]configuration", "[]deviceprofile", "[]deviceservice", "[]id", "[]user", "[]userself", "[]certificate":
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

	switch p.NamedLookup.Type {
	case "external":
		// TODO: Support controlling what looks like an ID what does not via a regex
		opts = append(opts, c8yfetcher.WithExternalCommandByNameFirstMatch(client, args, p.NamedLookup.Command, "", p.Name, targetProp, p.Format))
		return opts
	}

	// return early if options have already been set
	if len(opts) > 0 {
		return opts
	}

	switch p.Type {
	case "file":
		opts = append(opts, flags.WithFormDataFileAndInfoWithTemplateSupport(cmdutil.NewTemplateResolver(factory, cfg), p.Name, flags.FlagDataName)...)
	case "attachment":
		opts = append(opts, flags.WithFormDataFile(p.Name, flags.FlagDataName)...)

	case "fileContents":
		opts = append(opts, flags.WithFilePath(p.Name, targetProp, p.Value))
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

	case "application":
		opts = append(opts, c8yfetcher.WithApplicationByNameFirstMatch(client, args, p.Name, targetProp, p.Format))

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

func AddPredefinedGroupsFlags(cmd *CmdOptions, factory *cmdutil.Factory, template models.CommandPreset) {

	queryOptions := []flags.GetOption{}
	switch template.Type {
	case PresetGetIdentity:
		cmd.Spec.Method = "GET"

		if identityType := cmd.Spec.Preset.GetOption("value"); identityType != "" {
			cmd.Spec.Path = fmt.Sprintf("/identity/externalIds/%s/{name}", identityType)
			cmd.Command.Flags().String("name", "", "External identity id/name (required) (accepts pipeline)")
			cmd.Path.Options = append(
				cmd.Path.Options,
				[]flags.GetOption{
					flags.WithStringValue("name", "name", "%s"),
				}...,
			)
		} else {
			cmd.Spec.Path = "/identity/externalIds/{type}/{name}"
			cmd.Command.Flags().String("type", "c8y_Serial", "External identity type")
			cmd.Command.Flags().String("name", "", "External identity id/name (required) (accepts pipeline)")
			cmd.Path.Options = append(
				cmd.Path.Options,
				[]flags.GetOption{
					flags.WithStringValue("type", "type", "%s"),
					flags.WithStringValue("name", "name", "%s"),
				}...,
			)
		}

	case PresetQueryInventoryChildren:
		cmd.Spec.Method = "GET"
		cmd.Spec.Path = fmt.Sprintf("inventory/managedObjects/{id}/%s", cmd.Spec.Preset.GetOption("type", "childDevices"))

		cmd.Command.Flags().StringSlice("id", []string{""}, "Managed object id. (required) (accepts pipeline)")
		cmd.Command.Flags().String("query", "", "Additional query filter")
		cmd.Command.Flags().String("queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
		cmd.Command.Flags().String("orderBy", "name", "Order by. e.g. _id asc or name asc or creationTime.date desc")
		cmd.Command.Flags().String("name", "", "Filter by name")
		cmd.Command.Flags().String("type", "", "Filter by type")
		cmd.Command.Flags().Bool("agents", false, "Only include agents")
		cmd.Command.Flags().String("fragmentType", "", "Filter by fragment type")
		cmd.Command.Flags().String("owner", "", "Filter by owner")
		cmd.Command.Flags().String("availability", "", "Filter by c8y_Availability.status")
		cmd.Command.Flags().String("lastMessageDateTo", "", "Filter c8y_Availability.lastMessage to a specific date")
		cmd.Command.Flags().String("lastMessageDateFrom", "", "Filter c8y_Availability.lastMessage from a specific date")
		cmd.Command.Flags().String("creationTimeDateTo", "", "Filter creationTime.date to a specific date")
		cmd.Command.Flags().String("creationTimeDateFrom", "", "Filter creationTime.date from a specific date")
		// cmd.Command.Flags().StringSlice("group", []string{""}, "Filter by group inclusion")
		cmd.Command.Flags().Bool("skipChildrenNames", false, "Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved")
		cmd.Command.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
		cmd.Command.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")
		cmd.Command.Flags().Bool("withGroups", false, "When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.")
		cmd.Command.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")

		cmd.Path.Options = append(
			cmd.Path.Options,
			[]flags.GetOption{
				// c8yfetcher.WithDeviceByNameFirstMatch(client, args, "id", "id"),
				flags.WithStringSliceValues("id", "id", "%s"),
			}...,
		)

		c8yQueryOptions := []flags.GetOption{
			flags.WithStaticStringValue("fixed", template.GetOption("value")),
			flags.WithStringValue("query", "query", "%s"),
			flags.WithStringValue("name", "name", "(name eq '%s')"),
			flags.WithStringValue("type", "type", "(type eq '%s')"),
			flags.WithDefaultBoolValue("agents", "agents", "has(com_cumulocity_model_Agent)"),
			flags.WithStringValue("fragmentType", "fragmentType", "has(%s)"),
			flags.WithStringValue("owner", "owner", "(owner eq '%s')"),
			flags.WithStringValue("availability", "availability", "(c8y_Availability.status eq '%s')"),
			flags.WithEncodedRelativeTimestamp("lastMessageDateTo", "lastMessageDateTo", "(c8y_Availability.lastMessage le '%s')"),
			flags.WithEncodedRelativeTimestamp("lastMessageDateFrom", "lastMessageDateFrom", "(c8y_Availability.lastMessage ge '%s')"),
			flags.WithEncodedRelativeTimestamp("creationTimeDateTo", "creationTimeDateTo", "(creationTime.date le '%s')"),
			flags.WithEncodedRelativeTimestamp("creationTimeDateFrom", "creationTimeDateFrom", "(creationTime.date ge '%s')"),
			// c8yfetcher.WithDeviceGroupByNameFirstMatch(client, args, "group", "group", "bygroupid(%s)"),
		}

		// Add extensions to cumulocity query builder
		for _, p := range template.Extensions {
			c8yQueryOptions = append(c8yQueryOptions, GetOption(cmd, &p, factory, nil, nil, nil)...)
		}

		// options
		queryOptions = append(
			queryOptions,

			flags.WithBoolValue("skipChildrenNames", "skipChildrenNames", ""),
			flags.WithBoolValue("withChildren", "withChildren", ""),
			flags.WithBoolValue("withChildrenCount", "withChildrenCount", ""),
			flags.WithBoolValue("withGroups", "withGroups", ""),
			flags.WithBoolValue("withParents", "withParents", ""),

			flags.WithCumulocityQuery(
				c8yQueryOptions,
				template.GetOption("param", "query"),
			),
		)

		cmd.Runtime = append(cmd.Runtime,
			flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id", "id"),
			flags.WithPipelineAliases("id", "deviceId", "source.id", "managedObject.id", "id"),
		)

	case PresetQueryInventory:
		// Cumulocity inventory query
		cmd.Spec.Method = "GET"
		cmd.Spec.Path = "inventory/managedObjects"

		// flags
		cmd.Command.Flags().String("query", "", "Additional query filter (accepts pipeline)")
		cmd.Command.Flags().String("queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
		cmd.Command.Flags().String("orderBy", "name", "Order by. e.g. _id asc or name asc or creationTime.date desc")
		cmd.Command.Flags().String("name", "", "Filter by name")
		cmd.Command.Flags().String("type", "", "Filter by type")
		cmd.Command.Flags().Bool("agents", false, "Only include agents")
		cmd.Command.Flags().String("fragmentType", "", "Filter by fragment type")
		cmd.Command.Flags().String("owner", "", "Filter by owner")
		cmd.Command.Flags().String("availability", "", "Filter by c8y_Availability.status")
		cmd.Command.Flags().String("lastMessageDateTo", "", "Filter c8y_Availability.lastMessage to a specific date")
		cmd.Command.Flags().String("lastMessageDateFrom", "", "Filter c8y_Availability.lastMessage from a specific date")
		cmd.Command.Flags().String("creationTimeDateTo", "", "Filter creationTime.date to a specific date")
		cmd.Command.Flags().String("creationTimeDateFrom", "", "Filter creationTime.date from a specific date")
		cmd.Command.Flags().StringSlice("group", []string{""}, "Filter by group inclusion")
		cmd.Command.Flags().Bool("skipChildrenNames", false, "Don't include the child devices names in the response. This can improve the API response because the names don't need to be retrieved")
		cmd.Command.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object. Set it to false to improve query performance.")
		cmd.Command.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")
		cmd.Command.Flags().Bool("withGroups", false, "When set to true it returns additional information about the groups to which the searched managed object belongs. This results in setting the assetParents property with additional information about the groups.")
		cmd.Command.Flags().Bool("withParents", false, "Include a flat list of all parents and grandparents of the given object")

		// completions
		cmd.Completion = append(
			cmd.Completion,
			completion.WithValidateSet("availability", "AVAILABLE", "UNAVAILABLE", "MAINTENANCE"),
			completion.WithDeviceGroup("group", func() (*c8y.Client, error) { return factory.Client() }),
		)

		// TODO: Remove client, cfg arguments from flags and just use factory to enable lazy setting of the client

		c8yQueryOptions := []flags.GetOption{
			flags.WithStaticStringValue("fixed", template.GetOption("value")),
			flags.WithStringValue("query", "query", "%s"),
			flags.WithStringValue("name", "name", "(name eq '%s')"),
			flags.WithStringValue("type", "type", "(type eq '%s')"),
			flags.WithDefaultBoolValue("agents", "agents", "has(com_cumulocity_model_Agent)"),
			flags.WithStringValue("fragmentType", "fragmentType", "has(%s)"),
			flags.WithStringValue("owner", "owner", "(owner eq '%s')"),
			flags.WithStringValue("availability", "availability", "(c8y_Availability.status eq '%s')"),
			flags.WithEncodedRelativeTimestamp("lastMessageDateTo", "lastMessageDateTo", "(c8y_Availability.lastMessage le '%s')"),
			flags.WithEncodedRelativeTimestamp("lastMessageDateFrom", "lastMessageDateFrom", "(c8y_Availability.lastMessage ge '%s')"),
			flags.WithEncodedRelativeTimestamp("creationTimeDateTo", "creationTimeDateTo", "(creationTime.date le '%s')"),
			flags.WithEncodedRelativeTimestamp("creationTimeDateFrom", "creationTimeDateFrom", "(creationTime.date ge '%s')"),
			// c8yfetcher.WithDeviceGroupByNameFirstMatch(client, args, "group", "group", "bygroupid(%s)"),
		}

		// Add extensions to cumulocity query builder
		for _, p := range template.Extensions {
			c8yQueryOptions = append(c8yQueryOptions, GetOption(cmd, &p, factory, nil, nil, nil)...)
		}

		// options
		queryOptions = append(
			queryOptions,

			flags.WithBoolValue("skipChildrenNames", "skipChildrenNames", ""),
			flags.WithBoolValue("withChildren", "withChildren", ""),
			flags.WithBoolValue("withChildrenCount", "withChildrenCount", ""),
			flags.WithBoolValue("withGroups", "withGroups", ""),
			flags.WithBoolValue("withParents", "withParents", ""),

			flags.WithCumulocityQuery(
				c8yQueryOptions,
				template.GetOption("param", "q"),
			),
		)

		cmd.Runtime = append(cmd.Runtime,
			flags.WithExtendedPipelineSupport("query", "query", false, "c8y_DeviceQueryString"),
			flags.WithPipelineAliases("lastMessageDateTo", "time", "creationTime", "lastUpdated"),
			flags.WithPipelineAliases("lastMessageDateFrom", "time", "creationTime", "lastUpdated"),
			flags.WithPipelineAliases("creationTimeDateTo", "time", "creationTime", "lastUpdated"),
			flags.WithPipelineAliases("creationTimeDateFrom", "time", "creationTime", "lastUpdated"),
			flags.WithPipelineAliases("group", "source.id", "managedObject.id", "id"),

			flags.WithCollectionProperty("managedObjects"),
		)
	}

	// Add flags/completions for preset extensions
	for _, p := range template.Extensions {
		AddFlag(cmd, &p, factory)
		cmd.Completion = AppendCompletionOptions(cmd.Completion, cmd, &p, factory)
	}

	cmd.QueryParameter = append(cmd.QueryParameter, queryOptions...)
}

var (
	CommonFlagsQuery                = "query"
	CommonFlagsTemplate             = "queryTemplate"
	CommonFlagsOrderBy              = "orderBy"
	CommonFlagsName                 = "name"
	CommonFlagsType                 = "type"
	CommonFlagsAgents               = "agents"
	CommonFlagsFragmentType         = "fragmentType"
	CommonFlagsOwner                = "owner"
	CommonFlagsAvailability         = "availability"
	CommonFlagsLastMessageDateTo    = "lastMessageDateTo"
	CommonFlagsLastMessageDateFrom  = "lastMessageDateFrom"
	CommonFlagsCreationTimeDateTo   = "creationTimeDateTo"
	CommonFlagsCreationTimeDateFrom = "creationTimeDateFrom"
	CommonFlagsGroup                = "group"
	CommonFlagsSkipChildrenNames    = "skipChildrenNames"
	CommonFlagsWithChildren         = "withChildren"
	CommonFlagsWithChildrenCount    = "withChildrenCount"
	CommonFlagsWithGroups           = "withGroups"
	CommonFlagsWithParents          = "withParents"
)

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
	Options              []flags.GetOption
	IsBinary             bool
	Initialize           bool
	UploadProgressSource string
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
