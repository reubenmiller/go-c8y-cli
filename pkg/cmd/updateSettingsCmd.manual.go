package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// UpdateSettingsCmd updates settings in a session
type UpdateSettingsCmd struct {
	file string

	*baseCmd
}

type argumentHandler struct {
	Name               string
	ValueType          string
	SettingName        string
	Options            []string
	Transform          func(v string) string
	ShellCompDirective cobra.ShellCompDirective
}

// TransformReplaceAll returns a func which handles replacement of text (i.e. strings.ReplaceAll)
func TransformReplaceAll(old string, new string) func(string) string {
	return func(s string) string {
		return strings.ReplaceAll(s, old, new)
	}
}

// GetValue converts a raw string value into its typed value (based on the .ValueType property)
func (h argumentHandler) GetValue(rawValue string) interface{} {
	switch h.ValueType {
	case "bool":
		if v, err := strconv.ParseBool(rawValue); err == nil {
			return v
		}
		return nil
	case "int":
		if v, err := strconv.ParseInt(rawValue, 10, 64); err == nil {
			return v
		}
		return nil
	default:
		return rawValue
	}
}

var updateSettingsOptions = map[string]argumentHandler{
	// mode
	// mode (shortcut)
	"mode": {"mode", "custom", "", []string{
		"prod\tProduction mode (read only)",
		"qual\tQA mode (delete disabled)",
		"dev\tDevelopment mode (no restrictions)",
		"ci\tCI mode (no restrictions)",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	"mode.enableCreate": {"mode.enableCreate", "bool", SettingsModeEnableCreate, []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	"mode.enableUpdate": {"mode.enableUpdate", "bool", SettingsModeEnableUpdate, []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	"mode.enableDelete": {"mode.enableDelete", "bool", SettingsModeEnableDelete, []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	"mode.confirmation": {"mode.confirmation", "string", SettingsModeConfirmation, []string{
		"GET,PUT,POST,DELETE\tAll methods",
		"PUT,POST,DELETE\tAll methods but GET",
	}, TransformReplaceAll(",", " "), cobra.ShellCompDirectiveNoFileComp},

	// encryption
	"encryption.enabled": {"encryption.enabled", "bool", SettingsEncryptionEnabled, []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},
	"encryption.cachePassphrase": {"encryption.cachePassphrase", "bool", config.SettingEncryptionCachePassphrase, []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	// template
	"template.path": {"template.path", "string", SettingsTemplatePath, []string{}, nil, cobra.ShellCompDirectiveFilterDirs},

	// settings path
	"settings.path": {"settings.path", "string", SettingsConfigPath, []string{"json"}, nil, cobra.ShellCompDirectiveFilterFileExt},

	// activity log
	"activityLog.path": {"activityLog.path", "string", SettingsActivityLogPath, []string{}, nil, cobra.ShellCompDirectiveFilterDirs},
	"activityLog.enabled": {"activityLog.enabled", "bool", SettingsActivityLogEnabled, []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},
	"activityLog.methodFilter": {"activityLog.methodFilter", "string", SettingsActivityLogMethodFilter, []string{
		"GET,PUT,POST,DELETE\tAll methods",
		"PUT,POST,DELETE\tAll method except GET",
	}, TransformReplaceAll(",", " "), cobra.ShellCompDirectiveNoFileComp},

	// Storage
	"storage.storePassword": {"storage.storePassword", "bool", SettingsStorageStorePassword, []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},
	"storage.storeCookies": {"storage.storeCookies", "bool", SettingsStorageStoreCookies, []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	// Include All
	"includeAll.pageSize": {"includeAll.pageSize", "int", SettingsIncludeAllPageSize, []string{
		"2000",
		"1000",
		"500",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	"includeAll.delayMS": {"includeAll.delayMS", "int", SettingsIncludeAllDelayMS, []string{
		"0",
		"50",
		"100",
		"250",
		"500",
		"1000",
		"2000",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	// defaults
	// dry format
	"defaults.dryformat": {"defaultDryFormat", "string", "settings.defaults.dryFormat", []string{
		"markdown\tMarkdown (default)",
		"json\tJSON representation of full request",
		"dump\tRaw HTTP Dump",
		"curl\tEquavalent curl command (does not support multi-part/formdata)",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	// max jobs
	"defaults.maxJobs": {"defaults.maxJobs", "int", "settings.defaults.maxJobs", []string{
		"10",
		"100",
		"1000",
		"10000",
		"0",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	"defaults.pageSize": {"defaults.pageSize", "int", "settings.defaults.pageSize", []string{
		"default",
		"10",
		"20",
		"100",
		"1000",
		"2000",
	}, TransformReplaceAll("default", fmt.Sprintf("%d", CumulocityDefaultPageSize)), cobra.ShellCompDirectiveNoFileComp},

	// no accept
	"defaults.noAccept": {"defaults.noAccept", "bool", "settings.defaults.noAccept", []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	// withError
	"defaults.withError": {"defaults.withError", "bool", "settings.defaults.withError", []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	// csvHeader
	"defaults.csvHeader": {"defaults.csvHeader", "bool", "settings.defaults.csvHeader", []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	// noColor
	"defaults.noColor": {"defaults.noColor", "bool", "settings.defaults.noColor", []string{
		"true",
		"false",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	// silentStatusCodes
	"defaults.silentStatusCodes": {"defaults.silentStatusCodes", "string", "settings.defaults.silentStatusCodes", []string{
		"404,409",
		"404",
		"409",
		"none",
	}, TransformReplaceAll(",", " "), cobra.ShellCompDirectiveNoFileComp},

	// timeout (in seconds)
	"defaults.timeout": {"defaults.timeout", "int", "settings.defaults.timeout", []string{
		"0",
		"10",
		"30",
		"60",
		"600",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	// workers
	"defaults.workers": {"defaults.workers", "int", "settings.defaults.workers", []string{
		"1",
		"2",
		"3",
		"4",
		"5",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	// abortOnErrors
	"defaults.abortOnErrors": {"defaults.abortOnErrors", "int", "settings.defaults.abortOnErrors", []string{
		"1",
		"10",
		"20",
		"50",
	}, nil, cobra.ShellCompDirectiveNoFileComp},

	// outputFile
	"defaults.outputFile": {"defaults.outputFile", "int", "settings.defaults.outputFile", []string{}, nil, cobra.ShellCompDirectiveDefault},
}

// UpdateSettingsCmd returns a new command used to update session settings
func NewUpdateSettingsCmd() *UpdateSettingsCmd {
	ccmd := &UpdateSettingsCmd{}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update session settings",
		Long:  `Update settings in the current session or a given session file`,
		Example: `
$ c8y settings update mode dev

Show the active cli settings
		`,
		RunE: ccmd.RunE,
		Args: cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 2 {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 0 {
				// keys
				keys := make([]string, 0)
				for key := range updateSettingsOptions {
					keys = append(keys, key)
				}
				return keys, cobra.ShellCompDirectiveNoFileComp
			} else {
				// values
				name := args[len(args)-1]

				if handler, ok := updateSettingsOptions[name]; ok {
					return handler.Options, cobra.ShellCompDirectiveNoSpace | handler.ShellCompDirective
				}
			}

			return []string{}, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringVar(&ccmd.file, "file", "", "Session or settings file to be modified")

	cmd.SilenceUsage = true

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateSettingsCmd) RunE(cmd *cobra.Command, args []string) error {

	var v *viper.Viper
	if n.file != "" {
		v = viper.New()
		v.SetConfigFile(n.file)
		if err := v.ReadInConfig(); err != nil {
			return err
		}
	} else {
		v = cliConfig.Persistent
	}

	name := ""
	value := ""

	for i := 0; i < len(args)-1; i += 2 {
		name = args[i]
		value = args[i+1]

		switch name {
		case "mode":
			if value == "prod" {
				v.Set(SettingsModeEnableCreate, false)
				v.Set(SettingsModeEnableUpdate, false)
				v.Set(SettingsModeEnableDelete, false)
				v.Set(SettingsModeCI, false)

			} else if value == "qual" {
				v.Set(SettingsModeEnableCreate, true)
				v.Set(SettingsModeEnableUpdate, true)
				v.Set(SettingsModeEnableDelete, false)
				v.Set(SettingsModeCI, false)

			} else if value == "dev" {
				v.Set(SettingsModeEnableCreate, true)
				v.Set(SettingsModeEnableUpdate, true)
				v.Set(SettingsModeEnableDelete, true)
				v.Set(SettingsModeCI, false)
			} else if value == "ci" {
				v.Set(SettingsModeCI, true)
			}
		default:
			if handler, ok := updateSettingsOptions[name]; ok {
				if handler.SettingName != "" {
					if handler.Transform != nil {
						value = handler.Transform(value)
					}
					v.Set(handler.SettingName, handler.GetValue(value))
				}
			}
		}
	}

	err := WriteAuth(v, globalStorageStorePassword, globalStorageStoreCookies)

	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.ErrOrStderr(), color.GreenString("Updated session file %s", v.ConfigFileUsed()))
	return nil
}