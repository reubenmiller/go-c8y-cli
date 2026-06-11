package codegen

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// srcArg is an argument source entry together with where it came from
// (PowerShell tags body parameters with an ArgSource=body note property).
type srcArg struct {
	v         *Value
	argSource string
}

// removeSkipped filters parameters marked with skip=true.
func removeSkipped(items []srcArg) []srcArg {
	kept := make([]srcArg, 0, len(items))
	for _, item := range items {
		if strings.EqualFold(item.v.Get("skip").Str(), "true") {
			continue
		}
		kept = append(kept, item)
	}
	return kept
}

func wrapArgs(values []*Value) []srcArg {
	wrapped := make([]srcArg, 0, len(values))
	for _, v := range values {
		wrapped = append(wrapped, srcArg{v: v})
	}
	return wrapped
}

var (
	reOverrideExclude = regexp.MustCompile(`(?i)device\b|agent\b|group|devicegroup|self|application|hostedapplication|software\b|softwareName\b|deviceservice\b|softwareversion\b|softwareversionName\b|firmware\b|firmwareName\b|firmwareversion\b|firmwareversionName\b|firmwarepatch\b|firmwarepatchName\b|configuration\b|deviceprofile\b|microservice|id\[\]|devicerequest\[\]`)
	reDeviceSuffix    = regexp.MustCompile(`(?i)device$`)
	reFormDataCode    = regexp.MustCompile(`(?i)^flags\.WithFormData`)
	reOptionCode      = regexp.MustCompile(`(?i)^(flags\.|c8yfetcher\.|With|c8ybinary\.)`)
)

func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func matchesMethod(method string, candidates ...string) bool {
	for _, c := range candidates {
		if containsFold(method, c) {
			return true
		}
	}
	return false
}

func clientFactoryArg() string {
	return "func() (*c8y.Client, error) { return ccmd.factory.Client()}),"
}

// GenerateCommand ports New-C8yApiGoCommand.ps1: it renders the unformatted
// source of one <alias>.auto.go subcommand file. parentName is the parent
// group package name (lowercased, dashes replaced).
func GenerateCommand(spec *Value, parentName string) (fileName string, src []byte) {
	name := spec.Get("alias").Get("go").Str()
	nameCamel := goCamelCase(name)
	packageName := goPackageName(name)
	fileName = lowerFirst(nameCamel) + ".auto.go"

	//
	// Meta information
	//
	use := name
	description := spec.Get("description").Str()
	descriptionLong := spec.Get("descriptionLong").Str()

	var examples []string
	for _, ex := range spec.Get("examples").Get("go").Items() {
		if ex.Get("command").Truthy() {
			command := strings.TrimRightFunc(ex.Get("command").Str(), unicode.IsSpace)
			examples = append(examples, "$ "+command+"\n"+ex.Get("description").Str())
		} else {
			examples = append(examples, ex.Str())
		}
	}

	restPath := strings.ReplaceAll(spec.Get("path").Str(), " ", "%20")
	restMethod := spec.Get("method").Str()

	commandOptions := ""
	if spec.Get("hidden").Truthy() {
		commandOptions = "\n\t\tHidden: true,\n"
	}

	//
	// Arguments
	//
	var argumentSources []srcArg
	for _, p := range spec.Get("pathParameters").Items() {
		argumentSources = append(argumentSources, srcArg{v: p})
	}
	for _, item := range spec.Get("queryParameters").Items() {
		if item.Get("children").Truthy() {
			// Ignore the item, and only use the children to build the cli arguments
			for _, child := range item.Get("children").Items() {
				if !child.Get("type").StrEquals("stringStatic") {
					argumentSources = append(argumentSources, srcArg{v: child})
				}
			}
		} else {
			argumentSources = append(argumentSources, srcArg{v: item})
		}
	}
	for _, p := range spec.Get("body").Items() {
		argumentSources = append(argumentSources, srcArg{v: p, argSource: "body"})
	}
	for _, p := range spec.Get("headerParameters").Items() {
		argumentSources = append(argumentSources, srcArg{v: p})
	}
	for _, p := range spec.Get("options").Items() {
		argumentSources = append(argumentSources, srcArg{v: p})
	}

	var commandArgs []*cmdArg
	pipelineVariableName := ""
	pipelineVariableRequired := "false"
	pipelineVariableProperty := ""
	var pipelineVariableAliases []string
	collectionProperty := spec.Get("collectionProperty").Str()
	deprecationNotice := ""
	if spec.Get("deprecated").Truthy() {
		deprecationNotice = spec.Get("deprecated").Str()
	}

	// Body init
	restBodyBuilder := &strings.Builder{}
	restFormDataBuilder := &strings.Builder{}

	completionBuilder := &strings.Builder{}
	for _, iArg := range removeSkipped(argumentSources) {
		argName := iArg.v.Get("name").Str()
		argType := iArg.v.Get("type").Str()

		if iArg.v.Get("pipeline").Truthy() {
			pipelineVariableName = argName
			if iArg.v.Get("required").Truthy() {
				pipelineVariableRequired = "true"
			} else {
				pipelineVariableRequired = "false"
			}
			pipelineVariableProperty = iArg.v.Get("property").Str()
			if pipelineVariableProperty == "" {
				pipelineVariableProperty = argName
			}
			pipelineVariableAliases = nil
			for _, alias := range iArg.v.Get("pipelineAliases").Items() {
				pipelineVariableAliases = append(pipelineVariableAliases, alias.Str())
			}
			if len(pipelineVariableAliases) == 0 {
				if reDeviceSuffix.MatchString(pipelineVariableName) || strings.EqualFold(argType, "device[]") {
					pipelineVariableAliases = []string{"deviceId", "source.id", "managedObject.id", "id"}
				} else if !strings.EqualFold(pipelineVariableName, "id") {
					pipelineVariableAliases = []string{"id"}
				}
			}

			if !reOverrideExclude.MatchString(argType) {
				if matchesMethod(restMethod, "POST") && iArg.argSource == "body" {
					// Add override capability to piped arguments, so the user can still override piped data with the argument
					restBodyBuilder.WriteString("flags.WithOverrideValue(" + q(argName) + ", " + q(pipelineVariableProperty) + "),\n")
				}
			}
		}
		if vs := iArg.v.Get("validationSet"); vs.Truthy() {
			var quoted []string
			for _, item := range vs.Items() {
				quoted = append(quoted, q(item.Str()))
			}
			completionBuilder.WriteString("completion.WithValidateSet(" + q(argName) + ", " + strings.Join(quoted, ",") + "),\n")
		}

		// Special system and tenant options completions
		if containsFold(parentName, "tenantoptions") || containsFold(parentName, "systemoptions") {
			completionName := parentName + ":" + argName
			if containsFold(completionName, "tenantoptions:category") {
				completionBuilder.WriteString("completion.WithTenantOptionCategory(" + q(argName) + ", " + clientFactoryArg() + "\n")
			}
			if containsFold(completionName, "tenantoptions:key") {
				completionBuilder.WriteString("completion.WithTenantOptionKey(" + q(argName) + ", \"category\", " + clientFactoryArg() + "\n")
			}
			if containsFold(completionName, "systemoptions:category") {
				completionBuilder.WriteString("completion.WithSystemOptionCategory(" + q(argName) + ", " + clientFactoryArg() + "\n")
			}
			if containsFold(completionName, "systemoptions:key") {
				completionBuilder.WriteString("completion.WithSystemOptionKey(" + q(argName) + ", \"category\", " + clientFactoryArg() + "\n")
			}
		}

		// Special measurement series/fragments completions
		if containsFold(parentName, "measurements") {
			completionName := parentName + ":" + argName
			if containsFold(completionName, "measurements:series") {
				completionBuilder.WriteString("completion.WithDeviceMeasurementSeries(" + q(argName) + ", \"device\", " + clientFactoryArg() + "\n")
			}
			if containsFold(completionName, "measurements:valueFragmentType") {
				completionBuilder.WriteString("completion.WithDeviceMeasurementValueFragmentType(" + q(argName) + ", \"device\", " + clientFactoryArg() + "\n")
			}
			if containsFold(completionName, "measurements:valueFragmentSeries") {
				completionBuilder.WriteString("completion.WithDeviceMeasurementValueFragmentSeries(" + q(argName) + ", \"device\", \"valueFragmentType\", " + clientFactoryArg() + "\n")
			}
		}

		// Special microservices completions
		if containsFold(parentName, "loglevels") {
			completionName := parentName + ":" + argName
			if containsFold(completionName, "loglevels:loggerName") {
				completionBuilder.WriteString("completion.WithMicroserviceLoggers(" + q(argName) + ", \"name\", " + clientFactoryArg() + "\n")
			}
		}

		// Add Completions based on type
		dependsOn := ""
		if deps := iArg.v.Get("dependsOn").Items(); len(deps) > 0 {
			dependsOn = deps[0].Str()
		}
		switch {
		case equalsAnyFold(argType, "application", "applicationname"):
			completionBuilder.WriteString("completion.WithApplication(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "application_with_versions"):
			completionBuilder.WriteString("completion.WithApplicationWithVersions(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "uiplugin"):
			completionBuilder.WriteString("completion.WithUIPlugin(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "uipluginversion"):
			completionBuilder.WriteString("completion.WithUIPluginVersion(" + q(argName) + ", " + q(dependsOn) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "hostedapplication"):
			completionBuilder.WriteString("completion.WithHostedApplication(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "microservice"):
			completionBuilder.WriteString("completion.WithMicroservice(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "microservicename"):
			completionBuilder.WriteString("completion.WithMicroservice(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "microserviceinstance"):
			completionBuilder.WriteString("completion.WithMicroserviceInstance(" + q(argName) + ", \"id\", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "feature"):
			completionBuilder.WriteString("completion.WithFeature(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case equalsAnyFold(argType, "role[]", "roleself[]"):
			completionBuilder.WriteString("completion.WithUserRole(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "devicerequest[]"):
			completionBuilder.WriteString("completion.WithDeviceRegistrationRequest(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case equalsAnyFold(argType, "user[]", "userself[]"):
			completionBuilder.WriteString("completion.WithUser(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "usergroup[]"):
			completionBuilder.WriteString("completion.WithUserGroup(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "devicegroup[]"):
			completionBuilder.WriteString("completion.WithDeviceGroup(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "smartgroup[]"):
			completionBuilder.WriteString("completion.WithSmartGroup(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case equalsAnyFold(argType, "tenant", "tenantname"):
			completionBuilder.WriteString("completion.WithTenantID(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "device[]"):
			completionBuilder.WriteString("completion.WithDevice(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "agent[]"):
			completionBuilder.WriteString("completion.WithAgent(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case equalsAnyFold(argType, "software[]", "softwareName"):
			completionBuilder.WriteString("completion.WithSoftware(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case equalsAnyFold(argType, "softwareversion[]", "softwareversionName"):
			completionBuilder.WriteString("completion.WithSoftwareVersion(" + q(argName) + ", " + q(dependsOn) + ", " + clientFactoryArg() + "\n")
		case equalsAnyFold(argType, "firmware[]", "firmwareName"):
			completionBuilder.WriteString("completion.WithFirmware(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case equalsAnyFold(argType, "firmwareversion[]", "firmwareversionName"):
			completionBuilder.WriteString("completion.WithFirmwareVersion(" + q(argName) + ", " + q(dependsOn) + ", " + clientFactoryArg() + "\n")
		case equalsAnyFold(argType, "firmwarepatch[]", "firmwarepatchName"):
			completionBuilder.WriteString("completion.WithFirmwarePatch(" + q(argName) + ", " + q(dependsOn) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "configuration[]"):
			completionBuilder.WriteString("completion.WithConfiguration(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "deviceprofile[]"):
			completionBuilder.WriteString("completion.WithDeviceProfile(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "deviceservice[]"):
			completionBuilder.WriteString("completion.WithDeviceService(" + q(argName) + ", " + q(dependsOn) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "certificate[]"):
			completionBuilder.WriteString("completion.WithDeviceCertificate(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "subscriptionName"):
			completionBuilder.WriteString("completion.WithNotification2SubscriptionName(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "subscriptionId"):
			completionBuilder.WriteString("completion.WithNotification2SubscriptionId(" + q(argName) + ", " + clientFactoryArg() + "\n")
		case strings.EqualFold(argType, "remoteaccessconfiguration"):
			completionBuilder.WriteString("completion.WithRemoteAccessConfiguration(" + q(argName) + ", " + q(dependsOn) + ", " + clientFactoryArg() + "\n")
		}

		commandArgs = append(commandArgs, goArgs(argInput{
			Name:              argName,
			Type:              argType,
			OptionName:        iArg.v.Get("alias").Str(),
			Description:       iArg.v.Get("description").Str(),
			Default:           iArg.v.Get("default").Str(),
			Required:          iArg.v.Get("required").Str(),
			Hidden:            iArg.v.Get("hidden").Str(),
			Deprecated:        iArg.v.Get("deprecated").Str(),
			DeprecationNotice: iArg.v.Get("deprecationNotice").Str(),
			Pipeline:          iArg.v.Get("pipeline").Str(),
		}))
	}

	// Post actions
	postActions := &strings.Builder{}
	postActionsTotal := 0
	for _, iArg := range removeSkipped(argumentSources) {
		urlProperty := iArg.v.Get("name").Str()
		if iArg.v.Get("property").Truthy() {
			urlProperty = iArg.v.Get("property").Str()
		}
		if strings.EqualFold(iArg.v.Get("type").Str(), "binaryUploadURL") {
			postActions.WriteString("&c8ydata.AddChildAddition{Client: client, URLProperty: " + q(urlProperty) + "},\n")
			postActionsTotal++
		}
	}
	postActionOptions := ""
	if postActionsTotal > 0 {
		postActionOptions = "inputIterators.PipeOptions.PostActions = []flags.Action{\n" + postActions.String() + "}\n"
	}

	// Prepare Request
	prepareRequest := ""

	//
	// Body
	//
	getBodyContents := "body"
	isFormData := false

	if spec.Get("body").Truthy() {
		switch {
		case spec.Get("bodyContent").Get("type").StrEquals("binary"):
			getBodyContents = "body.GetFileContents()"
		case spec.Get("bodyContent").Get("type").StrEquals("formdata"):
			getBodyContents = "body"
			isFormData = true
		default:
			getBodyContents = "body"
			restBodyBuilder.WriteString("flags.WithDataFlagValue(),\n")
		}

		hasProgress := false
		for _, iArg := range removeSkipped(wrapArgs(spec.Get("body").Items())) {
			if matchesMethod(restMethod, "POST", "PUT") && !hasProgress {
				if equalsAnyFold(iArg.v.Get("type").Str(), "file", "fileContents", "attachment") {
					hasProgress = true
					prepareRequest = "PrepareRequest: c8ybinary.AddProgress(cmd, " + q(iArg.v.Get("name").Str()) + ", cfg.GetProgressBar(n.factory.IOStreams.ErrOut, n.factory.IOStreams.IsStderrTTY())),"
				}
			}

			if iArg.v.Get("options").Truthy() && iArg.v.Get("options").Get("formData").Truthy() {
				restFormDataBuilder.WriteString("Append(flags.WithFormDataProperty(" + q(iArg.v.Get("options").Get("formData").Str()) + ")).\n")
			}

			code := getValueFromFlag(iArg.v, "body")
			if code != "" {
				switch {
				case reFormDataCode.MatchString(code):
					restFormDataBuilder.WriteString("Append(" + strings.TrimRight(strings.TrimRight(code, ","), ".") + "...).\n")
				case reOptionCode.MatchString(code):
					if isFormData {
						restFormDataBuilder.WriteString("Append(" + strings.TrimRight(code, ",") + ").\n")
					} else {
						restBodyBuilder.WriteString(code + "\n")
					}
				}
			}
		}

		//
		// Activate separate body templating (if not included in -Data parameter)
		//
		if strings.EqualFold(spec.Get("bodyTemplateOptions").Get("enabled").Str(), "true") {
			commandArgs = append(commandArgs, &cmdArg{
				setFlagOptions: []string{"f.WithTemplateFlag(cmd)"},
			})
		}

		//
		// Apply a body template to the data
		//
		for _, bodyTemplate := range spec.Get("bodyTemplates").Items() {
			if bodyTemplate.Get("type").StrEquals("jsonnet") {
				// ApplyLast: true == apply template to the existing json (potentially overriding values)
				//            false == Use template as base json, and the existing json will take precedence
				if bodyTemplate.Get("applyLast").StrEquals("true") {
					restBodyBuilder.WriteString("flags.WithRequiredTemplateString(`\n" + bodyTemplate.Get("template").Str() + "`),\n")
				} else {
					restBodyBuilder.WriteString("flags.WithDefaultTemplateString(`\n" + bodyTemplate.Get("template").Str() + "`),\n")
				}
			}
		}

		//
		// Add support for user defined templates to control body
		//
		if bodyTemplatesAllowUserTemplates(spec.Get("bodyTemplates")) {
			if isFormData {
				restFormDataBuilder.WriteString("Append(cmdutil.WithTemplateValue(n.factory)).\n")
				restFormDataBuilder.WriteString("Append(flags.WithTemplateVariablesValue()).\n")
			} else {
				restBodyBuilder.WriteString("cmdutil.WithTemplateValue(n.factory),\n")
				restBodyBuilder.WriteString("flags.WithTemplateVariablesValue(),\n")
			}
		}

		if keys := spec.Get("bodyRequiredKeys"); keys.Truthy() {
			var literals []string
			for _, key := range keys.Items() {
				literals = append(literals, q(key.Str()))
			}
			restBodyBuilder.WriteString("flags.WithRequiredProperties(" + strings.Join(literals, ", ") + "),\n")
		}
	}

	//
	// Host
	//
	restHost := ""
	if !spec.Get("host").IsNull() {
		restHost = "\nHost:         replacePathParameters(" + q(spec.Get("host").Str()) + ", pathParameters),"
	}

	//
	// Path Parameters
	//
	restPathBuilder := &strings.Builder{}
	for _, p := range removeSkipped(wrapArgs(spec.Get("pathParameters").Items())) {
		if code := getValueFromFlag(p.v, "path"); code != "" {
			restPathBuilder.WriteString(code + "\n")
		}
	}

	//
	// Query parameters
	//
	restQueryBuilder := &strings.Builder{}
	cumulocityQueryBuilder := &strings.Builder{}
	restQueryBuilderPost := &strings.Builder{}

	for _, p := range removeSkipped(wrapArgs(spec.Get("queryParameters").Items())) {
		if p.v.Get("type").StrEquals("queryExpression") && p.v.Get("children").Truthy() {
			cumulocityQueryBuilder.WriteString("\t\tflags.WithCumulocityQuery(\n")
			cumulocityQueryBuilder.WriteString("\t\t\t[]flags.GetOption{\n")

			for _, child := range p.v.Get("children").Items() {
				// Ignore special in-built values as these are handled separately
				if equalsAnyFold(child.Get("name").Str(), "queryTemplate", "orderBy") {
					continue
				}
				// Special case to handle Cumulocity query language builder
				if code := getValueFromFlag(child, "query"); code != "" {
					cumulocityQueryBuilder.WriteString(code + "\n")
				}
			}
			cumulocityQueryBuilder.WriteString("\t\t\t},\n")
			cumulocityQueryBuilder.WriteString("\t\t\t" + q(p.v.Get("property").Str()) + ",\n")
			cumulocityQueryBuilder.WriteString("\t\t),\n")
		} else {
			if code := getValueFromFlag(p.v, "query"); code != "" {
				restQueryBuilder.WriteString(code + "\n")
			}
		}
	}
	if matchesMethod(restMethod, "GET") {
		restQueryBuilderPost.WriteString("        commonOptions, err := cfg.GetOutputCommonOptions(cmd)\n" +
			"        if err != nil {\n" +
			"            return cmderrors.NewUserError(fmt.Sprintf(\"Failed to get common options. err=%s\", err))\n" +
			"        }\n")
		// Support mapping common flags to custom query parameters (e.g. if a service uses limit instead of pageSize)
		// but for consistency the flag --pageSize will still be used and mapped to 'limit'.
		if flagMapping := spec.Get("flagMapping"); flagMapping.Truthy() {
			var flagAliases []string
			for _, key := range flagMapping.Keys() {
				flagAliases = append(flagAliases, q(key)+":"+q(flagMapping.Get(key).Str()))
			}
			restQueryBuilderPost.WriteString("commonOptions.AddQueryParametersWithMapping(query, map[string]string{" + strings.Join(flagAliases, ",") + "})\n")
		} else {
			restQueryBuilderPost.WriteString("commonOptions.AddQueryParameters(query)\n")
		}
	}

	//
	// Headers
	//
	restHeaderBuilder := &strings.Builder{}
	for _, iArg := range removeSkipped(wrapArgs(spec.Get("headerParameters").Items())) {
		if code := getValueFromFlag(iArg.v, "header"); code != "" {
			restHeaderBuilder.WriteString(code + "\n")
		}
	}

	if spec.Get("contentType").Truthy() {
		restHeaderBuilder.WriteString("flags.WithStaticStringValue(\"Content-Type\", " + q(spec.Get("contentType").Str()) + "),\n")
	}

	if spec.Get("addAccept").Truthy() && spec.Get("accept").Truthy() {
		restHeaderBuilder.WriteString("flags.WithStaticStringValue(\"Accept\", " + q(spec.Get("accept").Str()) + "),\n")
	}

	// Processing Mode
	if matchesMethod(restMethod, "DELETE", "PUT", "POST") {
		restHeaderBuilder.WriteString("flags.WithProcessingModeValue(),\n")
	}

	// Add common parameters
	flagBuilderOptions := &strings.Builder{}
	if matchesMethod(restMethod, "DELETE", "PUT", "POST") {
		flagBuilderOptions.WriteString("flags.WithProcessingMode(),\n")
	}

	//
	// Pre run validation (disable some commands without switch flags)
	//
	functionalMethod := restMethod
	if spec.Get("semanticMethod").Truthy() {
		functionalMethod = spec.Get("semanticMethod").Str()
	}

	preRunFunction := "nil"
	switch {
	case strings.EqualFold(functionalMethod, "POST"):
		preRunFunction = "f.CreateModeEnabled(cmd)"
	case strings.EqualFold(functionalMethod, "PUT"):
		preRunFunction = "f.UpdateModeEnabled(cmd)"
	case strings.EqualFold(functionalMethod, "DELETE"):
		preRunFunction = "f.DeleteModeEnabled(cmd)"
	}

	// Additional options
	requestOptionsBuilder := ""
	if spec.Get("responseType").StrEquals("array") {
		requestOptionsBuilder = "ResponseData: make([]map[string]interface{}, 0),\n"
	}

	//
	// Template
	//
	b := &strings.Builder{}

	b.WriteString("// Code generated from specification version 1.0.0: DO NOT EDIT\n")
	b.WriteString("package " + packageName + "\n")
	b.WriteString("\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString("\t\"io\"\n")
	b.WriteString("\t\"net/http\"\n")
	b.WriteString("\t\"net/url\"\n")
	b.WriteString("\n")
	b.WriteString("    \"github.com/MakeNowJust/heredoc/v2\"\n")
	b.WriteString("    \"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary\"\n")
	b.WriteString("    \"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher\"\n")
	b.WriteString("    \"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand\"\n")
	b.WriteString("    \"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors\"\n")
	b.WriteString("    \"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil\"\n")
	b.WriteString("    \"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion\"\n")
	b.WriteString("    \"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags\"\n")
	b.WriteString("\t\"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder\"\n")
	b.WriteString("\t\"github.com/reubenmiller/go-c8y/pkg/c8y\"\n")
	b.WriteString("\t\"github.com/spf13/cobra\"\n")
	b.WriteString(")\n")
	b.WriteString("\n")
	b.WriteString("// " + nameCamel + "Cmd command\n")
	b.WriteString("type " + nameCamel + "Cmd struct {\n")
	b.WriteString("    *subcommand.SubCommand\n")
	b.WriteString("\n")
	b.WriteString("    factory *cmdutil.Factory\n")
	b.WriteString("}\n")
	b.WriteString("\n")
	b.WriteString("// New" + nameCamel + "Cmd creates a command to " + description + "\n")
	b.WriteString("func New" + nameCamel + "Cmd(f *cmdutil.Factory) *" + nameCamel + "Cmd {\n")
	b.WriteString("\tccmd := &" + nameCamel + "Cmd{\n")
	b.WriteString("        factory: f,\n")
	b.WriteString("    }\n")
	b.WriteString("\tcmd := &cobra.Command{\n")
	b.WriteString("\t\tUse:   \"" + use + "\",\n")
	b.WriteString("\t\tShort: \"" + description + "\",\n")
	b.WriteString("\t\tLong:  `" + descriptionLong + "`," + commandOptions + "\n")
	b.WriteString("        Example: heredoc.Doc(`\n")
	b.WriteString(strings.Join(examples, "\n\n") + "\n")
	b.WriteString("        `),\n")
	b.WriteString("        PreRunE: func(cmd *cobra.Command, args []string) error {\n")
	b.WriteString("            return " + preRunFunction + "\n")
	b.WriteString("        },\n")
	b.WriteString("\t\tRunE: ccmd.RunE,\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    cmd.SilenceUsage = true\n")
	b.WriteString("\n")

	var setFlags []string
	for _, item := range commandArgs {
		if item != nil && item.hasSetFlag {
			setFlags = append(setFlags, item.setFlag)
		}
	}
	b.WriteString("    " + strings.Join(setFlags, "\n\t") + "\n")
	b.WriteString("\n")
	b.WriteString("    completion.WithOptions(\n")
	b.WriteString("\t\tcmd,\n")
	b.WriteString("\t\t" + completionBuilder.String() + "\n")
	b.WriteString("\t)\n")
	b.WriteString("\n")
	b.WriteString("    flags.WithOptions(\n")
	b.WriteString("\t\tcmd,\n")
	b.WriteString("        " + strings.TrimRightFunc(flagBuilderOptions.String(), unicode.IsSpace) + "\n")

	var setFlagOptions []string
	for _, item := range commandArgs {
		if item != nil {
			setFlagOptions = append(setFlagOptions, item.setFlagOptions...)
		}
	}
	setFlagOptionsBlock := ""
	if len(setFlagOptions) > 0 {
		setFlagOptionsBlock = strings.Join(setFlagOptions, ",\n") + ","
	}
	b.WriteString("        " + setFlagOptionsBlock + "\n")

	if len(pipelineVariableAliases) > 0 {
		aliases := strings.Join(quoteAll(pipelineVariableAliases), ", ")
		b.WriteString("        flags.WithExtendedPipelineSupport(" + q(pipelineVariableName) + ", " + q(pipelineVariableProperty) + ", " + pipelineVariableRequired + ", " + aliases + "),\n")
	} else {
		b.WriteString("        flags.WithExtendedPipelineSupport(" + q(pipelineVariableName) + ", " + q(pipelineVariableProperty) + ", " + pipelineVariableRequired + "),\n")
	}

	var pipelineAliasLines []string
	for _, item := range commandArgs {
		if item == nil || len(item.pipelineAliases) == 0 {
			continue
		}
		used := map[string]bool{}
		var sourceAliases []string
		for _, alias := range item.pipelineAliases {
			if !used[alias] {
				sourceAliases = append(sourceAliases, q(alias))
			}
			used[alias] = true
		}
		if len(sourceAliases) > 0 {
			pipelineAliasLines = append(pipelineAliasLines, "flags.WithPipelineAliases("+q(item.name)+", "+strings.Join(sourceAliases, ", ")+"),\n")
		}
	}
	b.WriteString("        " + strings.Join(pipelineAliasLines, " ") + "\n")

	var trailerLines []string
	if collectionProperty != "" {
		trailerLines = append(trailerLines, "flags.WithCollectionProperty("+q(collectionProperty)+"),\n")
	}
	if deprecationNotice != "" {
		trailerLines = append(trailerLines, "flags.WithDeprecationNotice("+q(deprecationNotice)+"),\n")
	}
	if spec.Get("semanticMethod").Truthy() {
		trailerLines = append(trailerLines, "flags.WithSemanticMethod("+q(spec.Get("semanticMethod").Str())+"),\n")
	}
	b.WriteString("        " + strings.Join(trailerLines, " ") + "\n")
	b.WriteString("\t)\n")
	b.WriteString("\n")
	b.WriteString("    // Required flags\n")

	var requiredLines, hiddenLines, deprecatedLines []string
	for _, item := range commandArgs {
		if item == nil {
			continue
		}
		if item.required != "" {
			requiredLines = append(requiredLines, item.required)
		}
		if item.hidden != "" {
			hiddenLines = append(hiddenLines, item.hidden)
		}
		if item.deprecated != "" {
			deprecatedLines = append(deprecatedLines, item.deprecated)
		}
	}
	b.WriteString("    " + strings.Join(requiredLines, "\n\t") + "\n")
	b.WriteString("    " + strings.Join(hiddenLines, "\n\t") + "\n")
	b.WriteString("    " + strings.Join(deprecatedLines, "\n\t") + "\n")
	b.WriteString("\n")
	b.WriteString("    ccmd.SubCommand = subcommand.NewSubCommand(cmd)\n")
	b.WriteString("\n")
	b.WriteString("\treturn ccmd\n")
	b.WriteString("}\n")
	b.WriteString("\n")
	b.WriteString("// RunE executes the command\n")
	b.WriteString("func (n *" + nameCamel + "Cmd) RunE(cmd *cobra.Command, args []string) error {\n")
	b.WriteString("    cfg, err := n.factory.Config()\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString("\t\treturn err\n")
	b.WriteString("\t}\n")
	b.WriteString("    // Runtime flag options\n")
	b.WriteString("    flags.WithOptions(\n")
	b.WriteString("\t\tcmd,\n")
	b.WriteString("\t\tflags.WithRuntimePipelineProperty(),\n")
	b.WriteString("\t)\n")
	b.WriteString("    client, err := n.factory.Client()\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString("\t\treturn err\n")
	b.WriteString("\t}\n")
	b.WriteString("    inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)\n")
	b.WriteString("    if err != nil {\n")
	b.WriteString("        return err\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    // query parameters\n")
	b.WriteString("    query := flags.NewQueryTemplate()\n")
	b.WriteString("    err = flags.WithQueryParameters(\n")
	b.WriteString("\t\tcmd,\n")
	b.WriteString("        query,\n")
	b.WriteString("        inputIterators,\n")
	b.WriteString("        flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, \"custom\"),\n")
	b.WriteString("        " + restQueryBuilder.String() + "\n")
	b.WriteString("        " + cumulocityQueryBuilder.String() + "\n")
	b.WriteString("    )\n")
	b.WriteString("    if err != nil {\n")
	b.WriteString("\t\treturn cmderrors.NewUserError(err)\n")
	b.WriteString("    }\n")
	b.WriteString("    " + restQueryBuilderPost.String() + "\n")
	b.WriteString("\tqueryValue, err := query.GetQueryUnescape(true)\n")
	b.WriteString("\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString("\t\treturn cmderrors.NewSystemError(\"Invalid query parameter\")\n")
	b.WriteString("\t}\n")
	b.WriteString("\n")
	b.WriteString("    // headers\n")
	b.WriteString("    headers := http.Header{}\n")
	b.WriteString("    err = flags.WithHeaders(\n")
	b.WriteString("\t\tcmd,\n")
	b.WriteString("        headers,\n")
	b.WriteString("        inputIterators,\n")
	b.WriteString("        flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, \"header\"),\n")
	b.WriteString("        " + restHeaderBuilder.String() + "\n")
	b.WriteString("    )\n")
	b.WriteString("    if err != nil {\n")
	b.WriteString("\t\treturn cmderrors.NewUserError(err)\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    // form data\n")
	b.WriteString("    formData := make(map[string]io.Reader)\n")
	b.WriteString("    err = flags.WithFormDataOptions(\n")
	b.WriteString("\t\tcmd,\n")
	b.WriteString("        formData,\n")
	b.WriteString("        inputIterators,")
	if restFormDataBuilder.Len() > 0 {
		b.WriteString("                flags.WithOptionBuilder().\n")
		b.WriteString("                    " + restFormDataBuilder.String() + "\n")
		b.WriteString("                Build()...")
	}
	b.WriteString("\n")
	b.WriteString("    )\n")
	b.WriteString("    if err != nil {\n")
	b.WriteString("\t\treturn cmderrors.NewUserError(err)\n")
	b.WriteString("    }\n")
	b.WriteString("    \n")
	b.WriteString("\n")
	b.WriteString("    // body\n")
	hasBody := "false"
	if matchesMethod(restMethod, "PUT", "POST") && spec.Get("body").Truthy() {
		hasBody = "true"
	}
	b.WriteString("    body := mapbuilder.NewInitializedMapBuilder(" + hasBody + ")\n")
	b.WriteString("    err = flags.WithBody(\n")
	b.WriteString("        cmd,\n")
	b.WriteString("        body,\n")
	b.WriteString("        inputIterators,\n")
	b.WriteString("        " + restBodyBuilder.String() + "\n")
	b.WriteString("    )\n")
	b.WriteString("    if err != nil {\n")
	b.WriteString("\t\treturn cmderrors.NewUserError(err)\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    // path parameters\n")
	b.WriteString("    path := flags.NewStringTemplate(\"" + restPath + "\")\n")
	b.WriteString("    err = flags.WithPathParameters(\n")
	b.WriteString("        cmd,\n")
	b.WriteString("        path,\n")
	b.WriteString("        inputIterators,\n")
	b.WriteString("        " + restPathBuilder.String() + "\n")
	b.WriteString("    )\n")
	b.WriteString("    if err != nil {\n")
	b.WriteString("        return err\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    req := c8y.RequestOptions{" + restHost + "\n")
	b.WriteString("        Method:       \"" + restMethod + "\",\n")
	b.WriteString("        Path:         path.GetTemplate(),\n")
	b.WriteString("        Query:        queryValue,\n")
	b.WriteString("        Body:         " + getBodyContents + ",\n")
	b.WriteString("        FormData:     formData,\n")
	b.WriteString("        Header:       headers,\n")
	b.WriteString("        IgnoreAccept: cfg.IgnoreAcceptHeader(),\n")
	b.WriteString("        DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),\n")
	b.WriteString("        " + prepareRequest + "\n")
	b.WriteString("        " + requestOptionsBuilder + "\n")
	b.WriteString("    }\n")
	b.WriteString("    " + postActionOptions + "\n")
	b.WriteString("    " + flagConstraints(spec) + "\n")
	b.WriteString("\n")
	b.WriteString("    return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)\n")
	b.WriteString("}\n")
	b.WriteString("\n")

	return fileName, []byte(b.String())
}

// bodyTemplatesAllowUserTemplates evaluates the PowerShell expression
// ($Specification.bodyTemplates.type -ne "none"): when bodyTemplates is
// missing the comparison is $null -ne "none" which is true; when it is a
// collection the comparison filters it and the result is truthy when any
// element type differs from "none".
func bodyTemplatesAllowUserTemplates(bodyTemplates *Value) bool {
	if bodyTemplates.IsNull() {
		return true
	}
	if bodyTemplates.kind == KindArray {
		for _, item := range bodyTemplates.arr {
			if !item.Get("type").StrEquals("none") {
				return true
			}
		}
		return false
	}
	return !bodyTemplates.Get("type").StrEquals("none")
}

// flagConstraints ports Get-FlagConstraints.
func flagConstraints(spec *Value) string {
	constraints := spec.Get("constraints")
	if constraints.IsNull() {
		return ""
	}

	joinGroup := func(group *Value) string {
		var parts []string
		prefix := ""
		for _, item := range group.Items() {
			parts = append(parts, prefix+q(item.Str()))
			prefix = ", "
		}
		return strings.Join(parts, " ")
	}

	var lines []string
	for _, flagGroup := range constraints.Items() {
		if !flagGroup.Get("type").StrEquals("queryParameters") {
			continue
		}
		if oneRequired := flagGroup.Get("oneRequired"); !oneRequired.IsNull() {
			lines = append(lines, fmt.Sprintf("flags.WithQueryParameterOneRequired(%s),\n", joinGroup(oneRequired)))
		}
		if mutuallyExclusive := flagGroup.Get("mutuallyExclusive"); !mutuallyExclusive.IsNull() {
			lines = append(lines, fmt.Sprintf("flags.WithQueryParameterMutuallyExclusive(%s),\n", joinGroup(mutuallyExclusive)))
		}
		if requiredTogether := flagGroup.Get("requiredTogether"); !requiredTogether.IsNull() {
			lines = append(lines, fmt.Sprintf("flags.WithQueryParameterRequiredTogether(%s),\n", joinGroup(requiredTogether)))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	parts := append([]string{"req.WithValidateFunc(\n"}, lines...)
	parts = append(parts, ")\n")
	return strings.Join(parts, " ")
}
