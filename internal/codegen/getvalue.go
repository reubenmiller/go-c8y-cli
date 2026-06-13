package codegen

import (
	"fmt"
	"strings"
)

// getValueFromFlagKeys lists the parameter types supported by
// getValueFromFlag. Lookup is case-insensitive, mirroring the PowerShell
// hashtable key comparison ($Definitions.Keys -eq $Type).
var getValueFromFlagKeys = []string{
	"file",
	"fileContents",
	"attachment",
	"formDataFile",
	"boolean",
	"booleanDefault",
	"optional_fragment",
	"datetime",
	"date",
	"string[]",
	"stringcsv[]",
	"inventoryChildType",
	"string",
	"stringAny",
	"stringStatic",
	"source",
	"integer",
	"float",
	"json_custom",
	"binaryUploadURL",
	"json",
	"tenant",
	"tenantname",
	"subscriptionName",
	"subscriptionId",
	"application",
	"application_with_versions",
	"hostedapplication",
	"applicationname",
	"microservice",
	"microserviceinstance",
	"microservicename",
	"feature",
	"uiplugin",
	"uipluginversion",
	"devicerequest[]",
	"id[]",
	"software[]",
	"softwareDetails",
	"configurationDetails",
	"softwareName",
	"softwareversion[]",
	"deviceservice[]",
	"softwareversionName",
	"certificatefile",
	"certificate[]",
	"firmware[]",
	"firmwareName",
	"firmwareversion[]",
	"firmwareversionName",
	"firmwareDetails",
	"firmwarepatch[]",
	"firmwarepatchName",
	"configuration[]",
	"deviceprofile[]",
	"device[]",
	"agent[]",
	"devicegroup[]",
	"smartgroup[]",
	"user[]",
	"userself[]",
	"roleself[]",
	"role[]",
	"usergroup[]",
	"remoteaccessconfiguration",
}

// getValueFromFlag ports New-C8yApiGoGetValueFromFlag.ps1: it returns the
// flags/c8yfetcher option code that reads a parameter value and applies it to
// the request part given by setterType (query, path, body or header).
func getValueFromFlag(p *Value, setterType string) string {
	prop := p.Get("name").Str()
	queryParam := p.Get("property").Str()
	if queryParam == "" {
		queryParam = prop
	}

	fixedValue := p.Get("value").Str()

	formatValue := ""
	if p.Get("format").Truthy() {
		formatValue = `, ` + q(p.Get("format").Str())
	}

	typeName := ""
	for _, key := range getValueFromFlagKeys {
		if strings.EqualFold(key, p.Get("type").Str()) {
			typeName = key
			break
		}
	}

	// Special type: encoded relative datetime when used as a query parameter
	if typeName == "datetime" && setterType == "query" {
		return fmt.Sprintf("flags.WithEncodedRelativeTimestamp(%s, %s%s),", q(prop), q(queryParam), formatValue)
	}

	switch typeName {
	// file (used in multipart/form-data uploads). It writes to the formData object instead of the body
	case "file":
		return fmt.Sprintf("flags.WithFormDataFileAndInfoWithTemplateSupport(cmdutil.NewTemplateResolver(n.factory), %s, \"data\"),", q(prop))

	// fileContents. File contents will be added to body
	case "fileContents":
		return fmt.Sprintf("flags.WithFilePath(%s, %s, %s),", q(prop), q(queryParam), q(fixedValue))

	// attachment (used in multipart/form-data uploads), without extra details
	case "attachment":
		return fmt.Sprintf("flags.WithFormDataFile(%s, \"data\")...,", q(prop))

	// multi-part file without extra details and control the form-data field name
	case "formDataFile":
		return fmt.Sprintf("flags.WithFileReader(%s, %s),", q(prop), q(queryParam))

	case "boolean":
		return fmt.Sprintf("flags.WithBoolValue(%s, %s, %s),", q(prop), q(queryParam), q(fixedValue))

	case "booleanDefault":
		return fmt.Sprintf("flags.WithDefaultBoolValue(%s, %s, %s),", q(prop), q(queryParam), q(fixedValue))

	case "optional_fragment":
		return fmt.Sprintf("flags.WithOptionalFragment(%s, %s, %s),", q(prop), q(queryParam), q(fixedValue))

	case "datetime":
		return fmt.Sprintf("flags.WithRelativeTimestamp(%s, %s%s),", q(prop), q(queryParam), formatValue)

	case "date":
		return fmt.Sprintf("flags.WithRelativeDate(false, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "string[]":
		return fmt.Sprintf("flags.WithStringSliceValues(%s, %s, %s),", q(prop), q(queryParam), q(fixedValue))

	case "stringcsv[]":
		return fmt.Sprintf("flags.WithStringSliceCSV(%s, %s, %s),", q(prop), q(queryParam), q(fixedValue))

	case "inventoryChildType":
		return fmt.Sprintf("flags.WithInventoryChildType(%s, %s%s),", q(prop), q(queryParam), formatValue)

	case "string", "source", "tenantname", "subscriptionName", "subscriptionId", "applicationname",
		"microserviceinstance", "microservicename", "feature", "uipluginversion", "softwareName",
		"softwareversionName", "firmwareName", "firmwareversionName":
		return fmt.Sprintf("flags.WithStringValue(%s, %s%s),", q(prop), q(queryParam), formatValue)

	case "stringAny":
		return fmt.Sprintf("flags.WithAnyStringValue(%s, %s%s),", q(prop), q(queryParam), formatValue)

	case "stringStatic":
		return fmt.Sprintf("flags.WithStaticStringValue(%s, %s),", q(prop), q(fixedValue))

	case "integer":
		return fmt.Sprintf("flags.WithIntValue(%s, %s%s),", q(prop), q(queryParam), formatValue)

	case "float":
		return fmt.Sprintf("flags.WithFloatValue(%s, %s%s),", q(prop), q(queryParam), formatValue)

	// json_custom: Only supported for use with the body
	case "json_custom":
		return fmt.Sprintf("flags.WithDataValue(%s, %s%s),", q(prop), q(queryParam), formatValue)

	// binaryUploadURL: uploads a binary and returns the URL
	case "binaryUploadURL":
		return fmt.Sprintf("c8ybinary.WithBinaryUploadURL(n.factory.Client, n.factory.IOStreams.ProgressIndicator(), %s, %s%s),", q(prop), q(queryParam), formatValue)

	// json - don't do anything because it should be manually set
	case "json":
		return ""

	case "tenant":
		return fmt.Sprintf("flags.WithStringDefaultValue(n.factory.GetTenant(), %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "application", "application_with_versions":
		return fmt.Sprintf("c8yfetcher.WithApplicationByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "hostedapplication":
		return fmt.Sprintf("c8yfetcher.WithHostedApplicationByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "microservice":
		return fmt.Sprintf("c8yfetcher.WithMicroserviceByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "uiplugin":
		return fmt.Sprintf("c8yfetcher.WithUIPluginByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "devicerequest[]", "id[]":
		return fmt.Sprintf("c8yfetcher.WithIDSlice(args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "software[]":
		return fmt.Sprintf("c8yfetcher.WithSoftwareByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "softwareDetails":
		return fmt.Sprintf("c8yfetcher.WithSoftwareVersionData(n.factory, \"software\", \"version\", \"url\", \"softwareType\", args, \"\", %s%s),", q(queryParam), formatValue)

	case "configurationDetails":
		return fmt.Sprintf("c8yfetcher.WithConfigurationFileData(n.factory, \"configuration\", \"configurationType\", \"url\", args, \"\", %s%s),", q(queryParam), formatValue)

	case "softwareversion[]":
		return fmt.Sprintf("c8yfetcher.WithSoftwareVersionByNameFirstMatch(n.factory, \"software\", args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "deviceservice[]":
		return fmt.Sprintf("c8yfetcher.WithDeviceServiceByNameFirstMatch(n.factory, \"device\", args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "certificatefile":
		return fmt.Sprintf("flags.WithCertificateFile(%s, %s),", q(prop), q(queryParam))

	case "certificate[]":
		return fmt.Sprintf("c8yfetcher.WithCertificateByNameFirstMatch(n.factory, args, %s, %s),", q(prop), q(queryParam))

	case "firmware[]":
		return fmt.Sprintf("c8yfetcher.WithFirmwareByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "firmwareversion[]":
		return fmt.Sprintf("c8yfetcher.WithFirmwareVersionByNameFirstMatch(n.factory, \"firmware\", args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "firmwareDetails":
		return fmt.Sprintf("c8yfetcher.WithFirmwareVersionData(n.factory, \"firmware\", \"version\", \"url\", args, \"\", %s),", q(queryParam))

	case "firmwarepatch[]":
		return fmt.Sprintf("c8yfetcher.WithFirmwarePatchByNameFirstMatch(n.factory, \"firmware\", args, %s, %s),", q(prop), q(queryParam))

	case "firmwarepatchName":
		return fmt.Sprintf("flags.WithStringValue(%s, %s),", q(prop), q(queryParam))

	case "configuration[]":
		return fmt.Sprintf("c8yfetcher.WithConfigurationByNameFirstMatch(n.factory, args, %s, %s),", q(prop), q(queryParam))

	case "deviceprofile[]":
		return fmt.Sprintf("c8yfetcher.WithDeviceProfileByNameFirstMatch(n.factory, args, %s, %s),", q(prop), q(queryParam))

	case "device[]":
		return fmt.Sprintf("c8yfetcher.WithDeviceByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "agent[]":
		return fmt.Sprintf("c8yfetcher.WithAgentByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "devicegroup[]":
		return fmt.Sprintf("c8yfetcher.WithDeviceGroupByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "smartgroup[]":
		return fmt.Sprintf("c8yfetcher.WithSmartGroupByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "user[]":
		return fmt.Sprintf("c8yfetcher.WithUserByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "userself[]":
		return fmt.Sprintf("c8yfetcher.WithUserSelfByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "roleself[]":
		return fmt.Sprintf("c8yfetcher.WithRoleSelfByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "role[]":
		return fmt.Sprintf("c8yfetcher.WithRoleByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "usergroup[]":
		return fmt.Sprintf("c8yfetcher.WithUserGroupByNameFirstMatch(n.factory, args, %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "remoteaccessconfiguration":
		return fmt.Sprintf("c8yfetcher.WithRemoteAccessConfigurationFirstMatch(n.factory, \"device\", args, %s, %s),", q(prop), q(queryParam))
	}

	// Unknown types produce no setter (the PowerShell hashtable lookup with a
	// non-matching key returns an empty result).
	return ""
}
