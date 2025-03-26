package golang

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/integration/models"
)

type CommandFlag struct {
	Code string
}

func GetFlagGetterCode(param models.Parameter) *CommandFlag {
	cmd := &CommandFlag{}

	// TODO: scripts/build-cli/New-C8yApiGoGetValueFromFlag.ps1
	switch param.Type {
	case "string":
		cmd.Code = fmt.Sprintf(`flags.WithStringValue("%s", "%s, "%s"),`, param.Name, param.GetTargetProperty(), param.Format)
	}
	return cmd
}

func GetFlagCode(param models.Parameter) *CommandFlag {
	cmd := &CommandFlag{}

	pipelineAliases := make([]string, 0)
	pipelineAliases = append(pipelineAliases, param.PipelineAliases...)

	useQuotes := true
	var fType string
	switch param.Type {
	case "json_custom", "directory", "softwareName", "softwareversionName", "firmwareName", "firmwareversionName", "firmwarepatchName", "binaryUploadURL", "feature", "inventoryChildType", "string", "stringAny", "subscriptionName", "subscriptionId", "stringStatic", "file", "formDataFile", "attachment", "fileContents", "certificatefile":
		fType = "String"
	case "datefrom", "dateto", "datetime", "date":
		fType = "String"
		pipelineAliases = append(pipelineAliases, "time", "creationTime", "creationTime", "lastUpdated")
	case "source":
		fType = "String"
		pipelineAliases = append(pipelineAliases, "id", "source.id", "managedObject.id", "deviceId")

	case "string[]", "stringcsv[]", "software[]", "softwareversion[]", "firmware[]", "firmwareversion[]", "firmwarepatch[]", "configuration[]", "deviceprofile[]", "deviceservice[]", "id[]", "devicerequest[]", "user[]", "userself[]", "certificate[]", "remoteaccessconfiguration":
		fType = "StringSlice"
	case "device[]", "agent[]":
		fType = "StringSlice"
		pipelineAliases = append(pipelineAliases, "deviceId", "source.id", "managedObject.id", "id")
	case "devicegroup[]":
		fType = "StringSlice"
		pipelineAliases = append(pipelineAliases, "source.id", "managedObject.id", "id")
	case "smartgroup[]":
		fType = "StringSlice"
		pipelineAliases = append(pipelineAliases, "managedObject.id")
	case "roleself[]":
		fType = "StringSlice"
		pipelineAliases = append(pipelineAliases, "self", "id")
	case "role[]", "usergroup[]":
		fType = "StringSlice"
		pipelineAliases = append(pipelineAliases, "id")
	case "application", "applicationname", "hostedapplication", "application_with_versions", "uiplugin", "uipluginversion", "microservice", "microserviceinstance":
		fType = "String"
		pipelineAliases = append(pipelineAliases, "id")
	case "microservicename":
		fType = "String"
		pipelineAliases = append(pipelineAliases, "name")
	case "integer":
		fType = "Int"
		useQuotes = false
		if param.Default == "" {
			param.Default = "0"
		}
	case "float":
		fType = "Float32"
		useQuotes = false
		if param.Default == "" {
			param.Default = "0"
		}
	case "tenant", "tenantname":
		fType = "String"
		pipelineAliases = append(pipelineAliases, "tenant", "owner.tenant.id")
	case "boolean", "booleanDefault", "optional_fragment":
		fType = "Bool"
		if param.Default == "" {
			param.Default = "false"
		}
	default:
		panic("Unknown flag type. " + param.Type)
	}

	defaultValue := param.Default
	if useQuotes {
		defaultValue = "\"" + param.Default + "\""
	}

	if param.ShortName != "" {
		cmd.Code = fmt.Sprintf(`cmd.Flags().%sP("%s", "%s", %s, "%s"),`, fType, param.Name, param.ShortName, defaultValue, param.GetDescription())
	} else {
		cmd.Code = fmt.Sprintf(`cmd.Flags().%s("%s", %s, "%s"),`, fType, param.Name, defaultValue, param.GetDescription())
	}

	return cmd
}
