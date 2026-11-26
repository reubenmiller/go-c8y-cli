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
// flags option code that reads a parameter value and applies it to
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

	// binaryUploadURL: the upload of the binary has to be added by hand (e.g. with the
	// c8ybinary.UploadBinary helper), so no value is read from the flag
	case "binaryUploadURL":
		return ""

	// json - don't do anything because it should be manually set
	case "json":
		return ""

	case "tenant":
		return fmt.Sprintf("flags.WithStringDefaultValue(n.factory.GetTenant(), %s, %s%s),", q(prop), q(queryParam), formatValue)

	case "certificatefile":
		return fmt.Sprintf("flags.WithCertificateFile(%s, %s),", q(prop), q(queryParam))

	// Lookups of a value by name or id are not resolved by the generated command (the go-c8y v2
	// client resolves them where it is needed), so the flag value is used as is.
	case "application", "application_with_versions", "hostedapplication", "microservice", "uiplugin",
		"remoteaccessconfiguration", "firmwarepatchName", "certificate[]", "configuration[]", "deviceprofile[]":
		return fmt.Sprintf("flags.WithStringValue(%s, %s%s),", q(prop), q(queryParam), formatValue)

	// Reference lists (ids or names) of a flag (which can be piped)
	case "devicerequest[]", "id[]", "software[]", "softwareversion[]", "deviceservice[]", "firmware[]",
		"firmwareversion[]", "firmwarepatch[]", "device[]", "agent[]", "devicegroup[]", "smartgroup[]",
		"user[]", "userself[]", "roleself[]", "role[]", "usergroup[]":
		return fmt.Sprintf("flags.WithSliceRefValue(%s, %s),", q(prop), q(queryParam))

	// Details of a software/firmware version or a configuration file have to be looked up by hand
	case "softwareDetails", "configurationDetails", "firmwareDetails":
		return ""
	}

	// Unknown types produce no setter (the PowerShell hashtable lookup with a
	// non-matching key returns an empty result).
	return ""
}

// parameterKind is the way the value of a flag is read and written to the request
func parameterKind(p *Value) string {
	switch strings.ToLower(p.Get("type").Str()) {
	case "stringstatic":
		return "static"
	case "boolean", "booleandefault", "optional_fragment":
		return "bool"
	case "datetime", "date":
		return "time"
	case "integer":
		return "int"
	case "float":
		return "float"
	case "string[]":
		return "slice"
	case "stringcsv[]":
		return "csv"
	case "tenant":
		return "tenant"
	case "json", "json_custom", "binaryuploadurl", "file", "filecontents", "attachment", "formdatafile",
		"certificatefile", "softwaredetails", "configurationdetails", "firmwaredetails":
		// not a plain request value
		return ""
	}
	for _, key := range getValueFromFlagKeys {
		if strings.EqualFold(key, p.Get("type").Str()) {
			return "string"
		}
	}
	// Unknown types produce no value
	return ""
}

// valueSetterCode returns the code which reads the value of a parameter from the flags (of the current
// input item, see c8ystream.Resolver), and sets it to a query parameter or a header. The target is
// either "query" (url.Values) or "header" (map[string]string).
func valueSetterCode(p *Value, target string) string {
	prop := p.Get("name").Str()
	key := p.Get("property").Str()
	if key == "" {
		key = prop
	}
	set := func(value string) string {
		if target == "header" {
			return fmt.Sprintf("headers[%s] = %s\n", q(key), value)
		}
		return fmt.Sprintf("query.Set(%s, %s)\n", q(key), value)
	}
	switch parameterKind(p) {
	case "static":
		return set(q(p.Get("value").Str()))
	case "bool":
		return fmt.Sprintf("if v := in.BoolPtrIfChanged(%s); v != nil {\n%s}\n", q(prop), set("strconv.FormatBool(*v)"))
	case "time":
		return fmt.Sprintf("if v := in.Time(%s); v != \"\" {\n%s}\n", q(prop), set("v"))
	case "int":
		return fmt.Sprintf("if cmd.Flags().Changed(%s) {\nv, _ := cmd.Flags().GetInt(%s)\n%s}\n", q(prop), q(prop), set("strconv.Itoa(v)"))
	case "float":
		return fmt.Sprintf("if cmd.Flags().Changed(%s) {\nv, _ := cmd.Flags().GetFloat32(%s)\n%s}\n", q(prop), q(prop), set("strconv.FormatFloat(float64(v), 'f', -1, 32)"))
	case "slice":
		if target == "header" {
			return fmt.Sprintf("if v := strings.Join(in.StringSlice(%s), \",\"); v != \"\" {\n%s}\n", q(prop), set("v"))
		}
		return fmt.Sprintf("for _, v := range in.StringSlice(%s) {\nif v != \"\" {\nquery.Add(%s, v)\n}\n}\n", q(prop), q(key))
	case "csv":
		return fmt.Sprintf("if v := strings.Join(in.StringSlice(%s), \",\"); v != \"\" {\n%s}\n", q(prop), set("v"))
	case "tenant":
		return fmt.Sprintf("{\nv := in.FirstString(%s)\nif v == \"\" {\nv = n.factory.GetTenant()\n}\nif v != \"\" {\n%s}\n}\n", q(prop), set("v"))
	case "string":
		return fmt.Sprintf("if v := in.FirstString(%s); v != \"\" {\n%s}\n", q(prop), set("v"))
	}
	return ""
}

// pathParameterCode returns the code which replaces a path parameter (and the same variable in
// the host) with the value of the flag (of the current input item, see c8ystream.Resolver)
func pathParameterCode(p *Value, hasHost bool) string {
	prop := p.Get("name").Str()
	key := p.Get("property").Str()
	if key == "" {
		key = prop
	}
	if parameterKind(p) == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("{\n")
	b.WriteString(fmt.Sprintf("v := in.FirstString(%s)\n", q(prop)))
	if parameterKind(p) == "tenant" {
		b.WriteString("if v == \"\" {\nv = n.factory.GetTenant()\n}\n")
	}
	b.WriteString("if v != \"\" {\n")
	b.WriteString(fmt.Sprintf("path = strings.ReplaceAll(path, %s, url.PathEscape(v))\n", q("{"+key+"}")))
	if hasHost {
		b.WriteString(fmt.Sprintf("host = strings.ReplaceAll(host, %s, v)\n", q("{"+key+"}")))
	}
	b.WriteString("}\n}\n")
	return b.String()
}
