package codegen

import (
	"fmt"
	"strconv"
	"strings"
)

// cmdArg is one entry of the $CommandArgs list built by Get-C8yGoArgs.ps1.
// Entries created for unknown flag types are nil.
type cmdArg struct {
	name            string
	setFlag         string
	hasSetFlag      bool
	setFlagOptions  []string
	pipelineAliases []string
	required        string
	hidden          string
	deprecated      string
}

// argInput carries the (stringified) parameter fields passed to Get-C8yGoArgs.
type argInput struct {
	Name              string
	Type              string
	OptionName        string
	Description       string
	Default           string
	Required          string
	Hidden            string
	Deprecated        string
	DeprecationNotice string
	Pipeline          string
}

func equalsAnyFold(v string, options ...string) bool {
	for _, o := range options {
		if strings.EqualFold(v, o) {
			return true
		}
	}
	return false
}

// goArgs ports Get-C8yGoArgs.ps1. It returns nil for unknown flag types (the
// PowerShell default switch branch only writes a warning and the entry stays
// null).
func goArgs(in argInput) *cmdArg {
	description := in.Description
	if matchesTrueYes(in.Required) {
		description += " (required)"
	}
	if matchesTrueYes(in.Pipeline) {
		description += " (accepts pipeline)"
	}

	stringFlag := func() string {
		return fmt.Sprintf("cmd.Flags().String(%s, %s, %s)", q(in.Name), q(in.Default), q(description))
	}
	stringSliceFlag := func() string {
		return fmt.Sprintf("cmd.Flags().StringSlice(%s, []string{%s}, %s)", q(in.Name), q(in.Default), q(description))
	}

	var entry *cmdArg

	switch {
	case strings.EqualFold(in.Type, "json"):
		entry = &cmdArg{
			setFlagOptions: []string{
				"flags.WithData()",
				"f.WithTemplateFlag(cmd)",
			},
		}

	// Usage: Accept json, but assign it to a nested property
	case strings.EqualFold(in.Type, "json_custom"):
		entry = &cmdArg{setFlag: stringFlag(), hasSetFlag: true}

	case equalsAnyFold(in.Type, "datefrom", "dateto", "datetime", "date"):
		entry = &cmdArg{
			setFlag:         stringFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"time", "creationTime", "creationTime", "lastUpdated"},
		}

	case strings.EqualFold(in.Type, "source"):
		entry = &cmdArg{
			setFlag:         stringFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"id", "source.id", "managedObject.id", "deviceId"},
		}

	case strings.EqualFold(in.Type, "directory"):
		entry = &cmdArg{setFlag: stringFlag(), hasSetFlag: true}

	case equalsAnyFold(in.Type, "string[]", "stringcsv[]"):
		entry = &cmdArg{setFlag: stringSliceFlag(), hasSetFlag: true}

	case equalsAnyFold(in.Type, "device[]", "agent[]"):
		entry = &cmdArg{
			setFlag:         stringSliceFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"deviceId", "source.id", "managedObject.id", "id"},
		}

	// Management repository types and device extensions
	case equalsAnyFold(in.Type, "software[]", "softwareversion[]", "firmware[]", "firmwareversion[]", "firmwarepatch[]", "configuration[]", "deviceprofile[]", "deviceservice[]"):
		entry = &cmdArg{setFlag: stringSliceFlag(), hasSetFlag: true}

	// Management name lookup
	case equalsAnyFold(in.Type, "softwareName", "softwareversionName", "firmwareName", "firmwareversionName", "firmwarepatchName"):
		entry = &cmdArg{setFlag: stringFlag(), hasSetFlag: true}

	case strings.EqualFold(in.Type, "binaryUploadURL"):
		entry = &cmdArg{setFlag: stringFlag(), hasSetFlag: true}

	case strings.EqualFold(in.Type, "devicegroup[]"):
		entry = &cmdArg{
			setFlag:         stringSliceFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"source.id", "managedObject.id", "id"},
		}

	case strings.EqualFold(in.Type, "id[]"):
		entry = &cmdArg{setFlag: stringSliceFlag(), hasSetFlag: true}

	case strings.EqualFold(in.Type, "smartgroup[]"):
		entry = &cmdArg{
			setFlag:         stringSliceFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"managedObject.id"},
		}

	case strings.EqualFold(in.Type, "roleself[]"):
		entry = &cmdArg{
			setFlag:         stringSliceFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"self", "id"},
		}

	case strings.EqualFold(in.Type, "role[]"):
		entry = &cmdArg{
			setFlag:         stringSliceFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"id"},
		}

	case strings.EqualFold(in.Type, "usergroup[]"):
		entry = &cmdArg{
			setFlag:         stringSliceFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"id"},
		}

	case equalsAnyFold(in.Type, "application", "applicationname", "hostedapplication", "application_with_versions", "uiplugin", "uipluginversion"):
		entry = &cmdArg{
			setFlag:         stringFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"id"},
		}

	case equalsAnyFold(in.Type, "microservice", "microserviceinstance"):
		entry = &cmdArg{
			setFlag:         stringFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"id"},
		}

	case strings.EqualFold(in.Type, "microservicename"):
		entry = &cmdArg{
			setFlag:         stringFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"name"},
		}

	case equalsAnyFold(in.Type, "feature", "inventoryChildType", "string", "stringAny", "subscriptionName", "subscriptionId", "stringStatic"):
		entry = &cmdArg{setFlag: stringFlag(), hasSetFlag: true}

	case strings.EqualFold(in.Type, "devicerequest[]"):
		entry = &cmdArg{setFlag: stringSliceFlag(), hasSetFlag: true}

	case strings.EqualFold(in.Type, "integer"):
		defaultInt, err := strconv.ParseInt(strings.TrimSpace(in.Default), 10, 64)
		if err != nil {
			defaultInt = 0
		}
		entry = &cmdArg{
			setFlag:    fmt.Sprintf("cmd.Flags().Int(%s, %d, %s)", q(in.Name), defaultInt, q(description)),
			hasSetFlag: true,
		}

	case strings.EqualFold(in.Type, "float"):
		defaultFloat, err := strconv.ParseFloat(strings.TrimSpace(in.Default), 64)
		if err != nil {
			defaultFloat = 0
		}
		entry = &cmdArg{
			setFlag:    fmt.Sprintf("cmd.Flags().Float32(%s, %s, %s)", q(in.Name), strconv.FormatFloat(defaultFloat, 'f', -1, 64), q(description)),
			hasSetFlag: true,
		}

	case equalsAnyFold(in.Type, "tenant", "tenantname"):
		entry = &cmdArg{
			setFlag:         stringFlag(),
			hasSetFlag:      true,
			pipelineAliases: []string{"tenant", "owner.tenant.id"},
		}

	case equalsAnyFold(in.Type, "file", "formDataFile", "attachment", "fileContents"):
		entry = &cmdArg{setFlag: stringFlag(), hasSetFlag: true}

	case equalsAnyFold(in.Type, "boolean", "booleanDefault", "optional_fragment"):
		defaultValue := in.Default
		if defaultValue == "" {
			defaultValue = "false"
		}
		entry = &cmdArg{
			setFlag:    fmt.Sprintf("cmd.Flags().Bool(%s, %s, %s)", q(in.Name), defaultValue, q(description)),
			hasSetFlag: true,
		}

	case equalsAnyFold(in.Type, "user[]", "userself[]", "certificate[]", "remoteaccessconfiguration"):
		entry = &cmdArg{setFlag: stringSliceFlag(), hasSetFlag: true}

	case strings.EqualFold(in.Type, "certificatefile"):
		entry = &cmdArg{setFlag: stringFlag(), hasSetFlag: true}

	default:
		// Unknown flag type: PowerShell only warns and produces a null entry.
		return nil
	}

	if matchesTrueYes(in.Required) && !matchesTrue(in.Pipeline) {
		entry.required = fmt.Sprintf("_ = cmd.MarkFlagRequired(%s)", q(in.Name))
	}

	if matchesTrueYes(in.Hidden) && !matchesTrue(in.Pipeline) {
		entry.hidden = fmt.Sprintf("_ = cmd.Flags().MarkHidden(%s)", q(in.Name))
	}

	if matchesTrueYes(in.Deprecated) {
		entry.deprecated = fmt.Sprintf("flags.MarkDeprecated(cmd, %s, %s)", q(in.Name), q(in.DeprecationNotice))
	}

	entry.name = in.Name

	return entry
}
