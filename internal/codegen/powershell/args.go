package powershell

import (
	"fmt"
	"strings"
)

// cmdletArgument is the result of New-C8yPowershellArguments.ps1: the
// PowerShell parameter name, data type and [Parameter()] definition entries
// of one specification argument.
type cmdletArgument struct {
	Name        string
	Type        string
	Definition  []string
	Description string
	Ignore      bool
}

// dataTypes maps the specification argument types to PowerShell data types.
// It ports the switch statement of New-C8yPowershellArguments.ps1, which
// matches case-insensitively, so the keys are lower case.
var dataTypes = map[string]string{
	"agent[]":                   "object[]",
	"certificate[]":             "object[]",
	"configuration[]":           "object[]",
	"device[]":                  "object[]",
	"devicegroup[]":             "object[]",
	"deviceprofile[]":           "object[]",
	"devicerequest[]":           "object[]",
	"firmware[]":                "object[]",
	"firmwarepatch[]":           "object[]",
	"firmwareversion[]":         "object[]",
	"id[]":                      "object[]",
	"role[]":                    "object[]",
	"roleself[]":                "object[]",
	"smartgroup[]":              "object[]",
	"software[]":                "object[]",
	"softwareversion[]":         "object[]",
	"deviceservice[]":           "object[]",
	"string[]":                  "string[]",
	"stringcsv[]":               "string[]",
	"[]tenant":                  "object[]",
	"user[]":                    "object[]",
	"usergroup[]":               "object[]",
	"userself[]":                "object[]",
	"application":               "object[]",
	"application_with_versions": "object[]",
	"applicationname":           "string",
	"uiplugin":                  "object[]",
	"uipluginversion":           "object[]",
	"attachment":                "string",
	"binaryuploadurl":           "string",
	"boolean":                   "switch",
	"booleandefault":            "switch",
	"certificatefile":           "string",
	"date":                      "string",
	"datefrom":                  "string",
	"datetime":                  "string",
	"dateto":                    "string",
	"directory":                 "string",
	"file":                      "string",
	"formdatafile":              "string",
	"filecontents":              "string",
	"firmwarename":              "object[]",
	"firmwarepatchname":         "object[]",
	"firmwareversionname":       "object[]",
	"float":                     "float",
	"hostedapplication":         "object[]",
	"id":                        "object[]",
	"integer":                   "long",
	"inventorychildtype":        "string",
	"json_custom":               "object",
	"json":                      "object",
	"microservice":              "object[]",
	"microserviceinstance":      "string",
	"microservicename":          "object[]",
	"feature":                   "string",
	"optional_fragment":         "switch",
	"remoteaccessconfiguration": "object[]",
	"set":                       "object[]",
	"softwarename":              "object[]",
	"softwareversionname":       "object[]",
	"source":                    "object",
	"string":                    "string",
	"stringany":                 "string",
	"strings":                   "string",
	"subscriptionid":            "string",
	"subscriptionname":          "string",
	"tenant":                    "object",
	"tenantname":                "string",
}

// ignoredTypes are the complex lookup types that must not be visible in
// PowerShell.
var ignoredTypes = map[string]bool{
	"stringstatic":         true,
	"queryexpression":      true,
	"softwaredetails":      true,
	"firmwaredetails":      true,
	"configurationdetails": true,
}

// newCmdletArgument ports New-C8yPowershellArguments.ps1.
func newCmdletArgument(name, argType, description, required string, readFromPipeline bool) (*cmdletArgument, error) {
	arg := &cmdletArgument{
		Name:        upperFirst(name),
		Description: description,
	}

	if matchesTrueYes(required) {
		arg.Definition = append(arg.Definition, "Mandatory = $true")
		arg.Description = description + " (required)"
	}

	// Piped argument ($Type -match "(source|id)" is a case-insensitive
	// substring match, so e.g. subscriptionId also qualifies)
	lowerType := strings.ToLower(argType)
	if strings.Contains(lowerType, "source") || strings.Contains(lowerType, "id") || readFromPipeline {
		arg.Definition = append(arg.Definition, "ValueFromPipeline=$true", "ValueFromPipelineByPropertyName=$true")
	}

	if ignoredTypes[lowerType] {
		arg.Ignore = true
		return arg, nil
	}

	dataType, ok := dataTypes[lowerType]
	if !ok {
		return nil, fmt.Errorf("unsupported type. %s", argType)
	}
	arg.Type = dataType
	return arg, nil
}
