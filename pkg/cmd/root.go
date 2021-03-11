package cmd

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/reubenmiller/go-c8y-cli/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

// Logger is used to record the log messages which should be visible to the user when using the verbose flag
var Logger *logger.Logger
var activityLogger *activitylogger.ActivityLogger

// Console provides a thread safe way to write to the console output
var Console *console.Console

// SecureDataAccessor reads and writes encrypted data
var SecureDataAccessor *encrypt.SecureData

// Build data
// These variables should be set using the -ldflags "-X github.com/reubenmiller/go-c8y-cli/pkg/cmd.version=1.0.0" when running go build
var buildVersion string
var buildBranch string

const (
	module = "c8yapi"
)

func init() {
	Logger = logger.NewLogger(module, logger.Options{})
	SecureDataAccessor = encrypt.NewSecureData("{encrypted}")
	rootCmd = newC8yCmd()
	// set seed for random generation
	rand.Seed(time.Now().UTC().UnixNano())
}

type baseCmd struct {
	cmd *cobra.Command
}

func (c *baseCmd) getCommand() *cobra.Command {
	return c.cmd
}

func newBaseCmd(cmd *cobra.Command) *baseCmd {
	return &baseCmd{cmd: cmd}
}

type c8yCmd struct {
	cobra.Command
	Logger *logger.Logger
	useEnv bool
}

func (c *c8yCmd) DryRunHandler(options *c8y.RequestOptions, req *http.Request) {

	if !globalFlagDryRun {
		return
	}
	if req == nil {
		Logger.Warn("Response is nil")
		return
	}
	w := c.Command.ErrOrStderr()
	if globalFlagPrintErrorsOnStdout {
		w = c.Command.OutOrStdout()
	}

	PrintRequestDetails(w, nil, req)
}
func (c *c8yCmd) createCumulocityClient() {
	c.Logger.Debug("Creating c8y client")
	httpClient := newHTTPClient(globalFlagNoProxy)

	// Only bind when not setting the session
	if c.useEnv {
		c.Logger.Info("Binding authorization environment variables")
		if err := cliConfig.BindAuthorization(); err != nil {
			c.Logger.Warnf("Failed to bind to authorization variables. %s", err)
		}
	}

	// Try reading session from file
	if sessionFilePath, readErr := ReadConfigFiles(viper.GetViper()); readErr != nil {
		Logger.Printf("Error reading config file. %s", readErr)
		// Fallback to reading session from environment variables
		// client = c8y.NewClientFromEnvironment(httpClient, true)
	} else {
		cliConfig.SetSessionFilePath(sessionFilePath)
	}

	// bind config to flags
	bindFlags(&rootCmd.Command, viper.GetViper())

	client = c8y.NewClient(
		httpClient,
		formatHost(cliConfig.GetHost()),
		cliConfig.GetTenant(),
		cliConfig.GetUsername(),
		cliConfig.MustGetPassword(),
		true,
	)

	client.SetRequestOptions(c8y.DefaultRequestOptions{
		DryRun:        globalFlagDryRun,
		DryRunHandler: c.DryRunHandler,
	})

	// load authentication
	if err := loadAuthentication(cliConfig, client); err != nil {
		Logger.Warnf("Could not load authentication. %s", err)
	}

	//
	// Timeout setting preference
	// 1. User provideds --timeout argument
	// 2. "timeout" is set in the session.json file (and value is greater than 0)
	// 3. C8Y_TIMEOUT is set and greater than 0
	if !rootCmd.Flags().Changed("timeout") {
		timeout := viper.GetFloat64("timeout")
		if timeout > 0 {
			globalFlagTimeout = timeout
			Logger.Debugf("timeout: %0.3f", timeout)
		}
	}

	// Should we use the tenant in the name or not
	if viper.IsSet("useTenantPrefix") {
		client.UseTenantInUsername = viper.GetBool("useTenantPrefix")
	}

	// Logger.Printf("Use tenant prefix: %v", client.UseTenantInUsername)

	// read additional configuration
	if err := readConfiguration(&rootCmd.Command); err != nil {
		Logger.Errorf("Could not read configuration. %s", err)
	}
	Console.Colorized = !globalFlagNoColor && !globalCSVOutput
	Console.Compact = globalFlagCompact || globalCSVOutput
	Console.IsJSON = !globalCSVOutput
	Console.Disabled = globalFlagProgressBar && isTerminal()

	// Add the realtime client
	client.Realtime = c8y.NewRealtimeClient(
		client.BaseURL.String(),
		newWebsocketDialer(globalFlagNoProxy),
		client.TenantName,
		client.Username,
		client.Password,
	)

	// Set realtime authorization
	if client.AuthorizationMethod == c8y.AuthMethodOAuth2Internal {
		client.Realtime.SetXSRFToken(client.GetXSRFToken())

		if len(client.Cookies) > 0 {
			if err := client.Realtime.SetCookies(client.Cookies); err != nil {
				Logger.Errorf("Failed to set websocket cookie jar. %s", err)
			}
		}
	}
}

func newC8yCmd() *c8yCmd {
	command := &c8yCmd{
		Logger: Logger,
	}
	command.Command = cobra.Command{

		Use:               "c8y",
		Short:             "Cumulocity command line interface",
		Long:              `A command line interface to interact with Cumulocity REST API. Ideal for quick prototyping, exploring the REST API and for Platform maintainers/power users`,
		PersistentPreRunE: command.checkSessionExists,
	}
	return command
}

var rootCmd *c8yCmd

var (
	client                           *c8y.Client
	cliConfig                        *config.CliConfiguration
	globalFlagStream                 bool
	globalFlagPageSize               int
	globalFlagIncludeAllPageSize     int
	globalFlagBatchMaxWorkers        int
	globalFlagBatchMaxJobs           int64
	globalFlagCurrentPage            int64
	globalFlagTotalPages             int64
	globalFlagIncludeAll             bool
	globalFlagIncludeAllDelayMS      int64
	globalFlagVerbose                bool
	globalFlagDebug                  bool
	globalFlagWithTotalPages         bool
	globalFlagCompact                bool
	globalFlagDryRun                 bool
	globalFlagDryRunFormat           string
	globalFlagProgressBar            bool
	globalFlagIgnoreAccept           bool
	globalFlagNoColor                bool
	globalFlagSessionFile            string
	globalFlagOutputFile             string
	globalFlagUseEnv                 bool
	globalFlagRaw                    bool
	globalFlagProxy                  string
	globalFlagNoProxy                bool
	globalFlagNoLog                  bool
	globalFlagActivityLogMessage     string
	globalFlagTimeout                float64
	globalCSVOutput                  bool
	globalCSVOutputHeaders           bool
	globalFlagTemplatePath           string
	globalFlagBatchWorkers           int
	globalFlagBatchDelayMS           int
	globalFlagBatchAbortOnErrorCount int
	globalFlagFlatten                bool
	globalFlagPrintErrorsOnStdout    bool
	globalFlagForce                  bool
	globalFlagConfirm                bool
	globalFlagConfirmText            string
	globalFlagSelect                 []string
	globalFlagSilentStatusCodes      string

	globalModeEnableCreate bool
	globalModeEnableUpdate bool
	globalModeEnableDelete bool
	globalCIMode           bool

	globalStorageStorePassword bool
	globalStorageStoreCookies  bool
)

// CumulocityDefaultPageSize is the default page size used by Cumulocity
const CumulocityDefaultPageSize int = 5

const (
	// SettingsIncludeAllPageSize property name used to control the default page size when using includeAll parameter
	SettingsIncludeAllPageSize string = "settings.includeAll.pageSize"

	// SettingsIncludeAllDelayMS property name used to control the delay between fetching the next page
	SettingsIncludeAllDelayMS string = "settings.includeAll.delayMS"

	// SettingsDefaultsPageSize default pageSize
	SettingsDefaultsPageSize string = "settings.defaults.pageSize"

	// SettingsDefaultBatchMaxWorkers property name used to control the hard limit on the maximum workers used in batch operations
	SettingsDefaultBatchMaxWorkers string = "settings.default.batchMaximumWorkers"

	// SettingsConfigPath configuration path
	SettingsConfigPath string = "settings.path"

	// SettingsTemplatePath template path where the template files are located
	SettingsTemplatePath string = "settings.template.path"

	// SettingsModeEnableCreate enables create (post) commands
	SettingsModeEnableCreate string = "settings.mode.enableCreate"

	// SettingsModeEnableUpdate enables update commands
	SettingsModeEnableUpdate string = "settings.mode.enableUpdate"

	// SettingsModeEnableDelete enables delete commands
	SettingsModeEnableDelete string = "settings.mode.enableDelete"

	// SettingsEncryptionEnabled enables encryption when storing sensitive session data
	SettingsEncryptionEnabled string = "settings.encryption.enabled"

	// SettingsStorageStoreCookies controls if the cookies are saved to the session file or not
	SettingsStorageStoreCookies string = "settings.storage.storeCookies"

	// SettingsStorageStorePassword controls if the password is saved to the session file or not
	SettingsStorageStorePassword string = "settings.storage.storePassword"

	// SettingsModeCI enable continuous integration mode (this will enable all commands)
	SettingsModeCI string = "settings.ci"

	// SettingsModeConfirmation sets the confirm mode
	SettingsModeConfirmation string = "settings.mode.confirmation"

	// SettingsActivityLogPath path where the activity log will be stored
	SettingsActivityLogPath string = "settings.activityLog.path"

	// SettingsActivityLogEnabled enables/disables the activity log
	SettingsActivityLogEnabled string = "settings.activityLog.enabled"

	// SettingsActivityLogMethodFilter filters the activity log entries by a space delimited methods, i.e. GET POST PUT
	SettingsActivityLogMethodFilter string = "settings.activityLog.methodFilter"
)

// SettingsGlobalName name of the settings file (without extension)
const SettingsGlobalName = "settings"

func (c *c8yCmd) checkCommandError(err error) {

	w := ioutil.Discard
	if globalFlagPrintErrorsOnStdout {
		w = rootCmd.OutOrStdout()
	}

	if errors.Is(err, ErrNoMatchesFound) {
		// Simulate a 404 error
		customErr := cmderrors.CommandError{}
		customErr.StatusCode = 404
		customErr.ExitCode = 4
		customErr.Message = err.Error()
		err = customErr
	}

	if cErr, ok := err.(cmderrors.CommandError); ok {
		if cErr.StatusCode == 403 || cErr.StatusCode == 401 {
			c.Logger.Error(fmt.Sprintf("Authentication failed (statusCode=%d). Try to run set-session again, or check the password", cErr.StatusCode))
		}

		// format errors as json messages
		// only log users errors
		if !cErr.IsSilent() && !strings.Contains(globalFlagSilentStatusCodes, fmt.Sprintf("%d", cErr.StatusCode)) {
			c.Logger.Errorf("%s", cErr)
			fmt.Fprintf(w, "%s\n", cErr.JSONString())
		}
	} else {
		// unexpected error
		cErr := cmderrors.NewSystemErrorF("%s", err)
		c.Logger.Errorf("%s", cErr)
		fmt.Fprintf(w, "%s\n", cErr.JSONString())
	}
}

// LogErrorF dynamically changes where the error is logged based on the users Silent Status Codes preferences
// Silent errors are only logged on the INFO level, where as non-silent errors are logged on the ERROR level
func LogErrorF(err error, format string, args ...interface{}) {
	errorLogger := Logger.Errorf
	if errors.Is(err, ErrNoMatchesFound) {
		if strings.Contains(globalFlagSilentStatusCodes, "404") {
			errorLogger = Logger.Infof
		}
	} else if cErr, ok := err.(cmderrors.CommandError); ok {

		// format errors as json messages
		// only log users errors
		if strings.Contains(globalFlagSilentStatusCodes, fmt.Sprintf("%d", cErr.StatusCode)) {
			errorLogger = Logger.Infof
		}
	}
	errorLogger(format, args...)
}

func (c *c8yCmd) checkSessionExists(cmd *cobra.Command, args []string) error {

	cmdStr := cmd.Use
	if cmd.HasParent() && cmd.Parent().Use != "c8y" {
		cmdStr = cmd.Parent().Use + " " + cmdStr
	}
	c.Logger.Debugf("command str: %s", cmdStr)

	if globalFlagSessionFile == "" || !(strings.HasPrefix(cmdStr, "sessions list") || c.Flags().Changed("session")) {
		c.useEnv = true
	}

	c.createCumulocityClient()

	// only setup activity log after the global config
	configureActivityLog()
	activityLogger.LogCommand(cmd, args, cmdStr, globalFlagActivityLogMessage)

	localCmds := []string{
		"completion",
		"sessions",
		"template",
		"version",
		"tenants getID",
		"tenants getId",
		"settings list",

		// allow hidden completion commands
		"__complete",
		"__completeNoDesc",
	}

	for i := range localCmds {
		if strings.HasPrefix(cmdStr, localCmds[i]) {
			return nil
		}
	}

	if client == nil {
		return cmderrors.NewSystemError("Client failed to load")
	}
	if client.BaseURL == nil || client.BaseURL.Host == "" {
		return cmderrors.NewUserErrorWithExitCode(102, "A c8y session has not been loaded. Please create or activate a session and try again")
	}

	return nil
}

func isTerminal() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}
	return false
}

func getOutputHeaders(input []string) (headers []byte) {
	if !globalCSVOutput || !globalCSVOutputHeaders || len(input) == 0 {
		return
	}
	if len(input) > 0 {
		return []byte(input[0] + "\n")
	}

	// TODO: improve detection by parsing more lines to find column names (if more lines are available)
	columns := make([][]byte, 0)
	for _, v := range globalFlagSelect {
		for _, column := range strings.Split(v, ",") {

			if i := strings.Index(column, ":"); i > -1 {
				columns = append(columns, []byte(column[0:i]))
			} else {
				columns = append(columns, []byte(column))
			}
		}
	}
	return append(bytes.Join(columns, []byte(",")), []byte("\n")...)
}

// configureRootCmd initializes the configuration manager and c8y client
func configureRootCmd() {
	// config file
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&globalFlagSessionFile, "session", "", "Session configuration")

	isTerm := isTerminal()
	Console = console.NewConsole(rootCmd.OutOrStdout(), getOutputHeaders)

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&globalFlagVerbose, "verbose", "v", false, "Verbose logging")
	rootCmd.PersistentFlags().IntVarP(&globalFlagPageSize, "pageSize", "p", CumulocityDefaultPageSize, "Maximum results per page")
	rootCmd.PersistentFlags().Int64Var(&globalFlagCurrentPage, "currentPage", 0, "Current page size which should be returned")
	rootCmd.PersistentFlags().Int64Var(&globalFlagTotalPages, "totalPages", 0, "Total number of pages to get")
	rootCmd.PersistentFlags().BoolVar(&globalFlagIncludeAll, "includeAll", false, "Include all results by iterating through each page")
	rootCmd.PersistentFlags().BoolVarP(&globalFlagWithTotalPages, "withTotalPages", "t", false, "Include all results")
	rootCmd.PersistentFlags().BoolVarP(&globalFlagCompact, "compact", "c", !isTerm, "Compact instead of pretty-printed output. Pretty print is the default if output is the terminal")
	rootCmd.PersistentFlags().BoolVar(&globalFlagCompact, "compress", !isTerm, "Alias for --compact for users coming from PowerShell")
	rootCmd.PersistentFlags().BoolVar(&globalFlagStream, "stream", true, "Stream transforms JSON arrays to single json objects to make them pipeable. Automatically activated when output is not the terminal")
	rootCmd.PersistentFlags().BoolVar(&globalFlagIgnoreAccept, "noAccept", false, "Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect")
	rootCmd.PersistentFlags().BoolVar(&globalFlagDryRun, "dry", false, "Dry run. Don't send any data to the server")
	rootCmd.PersistentFlags().StringVar(&globalFlagDryRunFormat, "dryFormat", "markdown", "Dry run output format. i.e. json, dump, markdown or curl")
	rootCmd.PersistentFlags().BoolVar(&globalFlagProgressBar, "progress", false, "Show progress bar. This will also disable any other verbose output")
	rootCmd.PersistentFlags().BoolVarP(&globalFlagNoColor, "noColor", "M", !isTerm, "Don't use colors when displaying log entries on the console")
	rootCmd.PersistentFlags().BoolVar(&globalFlagUseEnv, "useEnv", false, "Allow loading Cumulocity session setting from environment variables")
	rootCmd.PersistentFlags().BoolVarP(&globalFlagRaw, "raw", "r", false, "Raw values")
	rootCmd.PersistentFlags().StringVar(&globalFlagProxy, "proxy", "", "Proxy setting, i.e. http://10.0.0.1:8080")
	rootCmd.PersistentFlags().BoolVar(&globalFlagNoProxy, "noProxy", false, "Ignore the proxy settings")
	rootCmd.PersistentFlags().BoolVar(&globalFlagPrintErrorsOnStdout, "withError", false, "Errors will be printed on stdout instead of stderr")

	// Activity log
	rootCmd.PersistentFlags().BoolVar(&globalFlagNoLog, "noLog", false, "Disables the activity log for the current command")
	rootCmd.PersistentFlags().StringVarP(&globalFlagActivityLogMessage, "logMessage", "l", "", "Add custom message to the activity log")
	rootCmd.PersistentFlags().BoolVar(&globalFlagDebug, "debug", false, "Set very verbose log messages")

	// Concurrency
	rootCmd.PersistentFlags().IntVar(&globalFlagBatchWorkers, "workers", 1, "Number of workers")
	rootCmd.PersistentFlags().Int64Var(&globalFlagBatchMaxJobs, "maxJobs", 100, "Maximum number of jobs. 0 = unlimited (use with caution!)")
	rootCmd.PersistentFlags().IntVar(&globalFlagBatchDelayMS, "delay", 1000, "delay in milliseconds after each request")
	rootCmd.PersistentFlags().IntVar(&globalFlagBatchAbortOnErrorCount, "abortOnErrors", 10, "Abort batch when reaching specified number of errors")

	// Error handling
	rootCmd.PersistentFlags().StringVar(&globalFlagSilentStatusCodes, "silentStatusCodes", "", "Status codes which will not print out an error message")

	rootCmd.PersistentFlags().StringVar(&globalFlagOutputFile, "outputFile", "", "Output file")

	rootCmd.PersistentFlags().BoolVar(&globalFlagFlatten, "flatten", false, "flatten")
	rootCmd.PersistentFlags().StringSlice("filter", nil, "filter")
	rootCmd.PersistentFlags().StringArrayVar(&globalFlagSelect, "select", nil, "select")
	rootCmd.PersistentFlags().BoolVar(&globalCSVOutput, "csv", false, "Print output as csv format. comma (,) delimited")
	rootCmd.PersistentFlags().BoolVar(&globalCSVOutputHeaders, "csvHeader", false, "Include header when in csv output")
	rootCmd.PersistentFlags().Float64Var(&globalFlagTimeout, "timeout", float64(10*60), "Timeout in seconds")

	// confirmation
	rootCmd.PersistentFlags().BoolVarP(&globalFlagForce, "force", "f", false, "Do not prompt for confirmation")
	rootCmd.PersistentFlags().BoolVar(&globalFlagConfirm, "confirm", false, "Prompt for confirmation")
	rootCmd.PersistentFlags().StringVar(&globalFlagConfirmText, "confirmText", "", "Custom confirmation text")

	completion.WithOptions(
		&rootCmd.Command,
		completion.WithValidateSet("dryFormat", "json", "dump", "curl", "markdown"),
	)

	rootCmd.AddCommand(NewCompletionsCmd().getCommand())
	rootCmd.AddCommand(NewVersionCmd().getCommand())

	rootCmd.AddCommand(NewRealtimeCmd().getCommand())
	rootCmd.AddCommand(NewSessionsRootCmd().getCommand())

	// generic commands
	rootCmd.AddCommand(NewGetGenericRestCmd().getCommand())

	// template commands
	rootCmd.AddCommand(NewTemplateRootCmd().getCommand())

	// settings commands
	rootCmd.AddCommand(NewSettingsRootCmd().getCommand())

	// Auto generated commands

	// agents commands
	agents := NewAgentsRootCmd().getCommand()
	agents.AddCommand(NewGetAgentCollectionCmd().getCommand())
	rootCmd.AddCommand(agents)

	// alarms commands
	alarms := NewAlarmsRootCmd().getCommand()
	alarms.AddCommand(NewSubscribeAlarmCmd().getCommand())
	rootCmd.AddCommand(alarms)

	// applications commands
	applications := NewApplicationsRootCmd().getCommand()
	applications.AddCommand(NewNewHostedApplicationCmd().getCommand())
	rootCmd.AddCommand(applications)

	// auditRecords commands
	rootCmd.AddCommand(NewAuditRecordsRootCmd().getCommand())

	// binaries commands
	rootCmd.AddCommand(NewBinariesRootCmd().getCommand())

	// bulkOperations commands
	rootCmd.AddCommand(NewBulkOperationsRootCmd().getCommand())

	// currentApplication commands
	rootCmd.AddCommand(NewCurrentApplicationRootCmd().getCommand())

	// currentUser commands
	rootCmd.AddCommand(newCurrentUserRootCmd().getCommand())

	// databroker commands
	rootCmd.AddCommand(NewDatabrokerRootCmd().getCommand())

	// deviceCredentials commands
	rootCmd.AddCommand(NewDeviceCredentialsRootCmd().getCommand())

	// devices commands
	devices := NewDevicesRootCmd().getCommand()
	devices.AddCommand(NewGetDeviceCollectionCmd().getCommand())
	devices.AddCommand(NewGetDeviceGroupCollectionCmd().getCommand())
	rootCmd.AddCommand(devices)

	// operations commands
	operations := NewOperationsRootCmd().getCommand()
	operations.AddCommand(NewSubscribeOperationCmd().getCommand())
	rootCmd.AddCommand(operations)

	// events commands
	events := NewEventsRootCmd().getCommand()
	events.AddCommand(NewSubscribeEventCmd().getCommand())
	rootCmd.AddCommand(events)

	// identity commands
	rootCmd.AddCommand(NewIdentityRootCmd().getCommand())

	// inventory commands
	inventory := NewInventoryRootCmd().getCommand()
	inventory.AddCommand(NewSubscribeManagedObjectCmd().getCommand())
	inventory.AddCommand(NewQueryManagedObjectCollectionCmd().getCommand())
	rootCmd.AddCommand(inventory)

	// inventoryReferences commands
	rootCmd.AddCommand(NewInventoryReferencesRootCmd().getCommand())

	// measurements commands
	measurements := NewMeasurementsRootCmd().getCommand()
	measurements.AddCommand(NewSubscribeMeasurementCmd().getCommand())
	rootCmd.AddCommand(measurements)

	// microservices commands
	microservices := NewMicroservicesRootCmd().getCommand()
	microservices.AddCommand(NewNewMicroserviceCmd().getCommand())
	microservices.AddCommand(NewNewServiceUserCmd().getCommand())
	microservices.AddCommand(NewGetServiceUserCmd().getCommand())
	rootCmd.AddCommand(microservices)

	// retentionRules commands
	rootCmd.AddCommand(NewRetentionRulesRootCmd().getCommand())

	// systemOptions commands
	rootCmd.AddCommand(NewSystemOptionsRootCmd().getCommand())

	// tenantOptions commands
	rootCmd.AddCommand(NewTenantOptionsRootCmd().getCommand())

	// tenants commands
	rootCmd.AddCommand(NewTenantsRootCmd().getCommand())

	// tenantStatistics commands
	rootCmd.AddCommand(NewTenantStatisticsRootCmd().getCommand())

	// users commands
	rootCmd.AddCommand(NewUsersRootCmd().getCommand())

	// userGroups commands
	rootCmd.AddCommand(NewUserGroupsRootCmd().getCommand())

	// userReferences commands
	rootCmd.AddCommand(NewUserReferencesRootCmd().getCommand())

	// userRoles commands
	rootCmd.AddCommand(NewUserRolesRootCmd().getCommand())

	// Handle errors (not in cobra libary)
	rootCmd.SilenceErrors = true

}

// Execute runs the root command
func Execute() {
	configureRootCmd()
	executeRootCmd()
}

func executeRootCmd() {
	if err := rootCmd.Execute(); err != nil {

		rootCmd.checkCommandError(err)

		if cErr, ok := err.(cmderrors.CommandError); ok {
			os.Exit(cErr.ExitCode)
		}
		if errors.Is(err, ErrNoMatchesFound) {
			// 404
			os.Exit(4)
		}
		os.Exit(100)
	}
}

// ReadConfigFiles reads multiple configuration files to load the c8y session and other settings
//
// The session files are
// 1. load settings (from C8Y_SESSION_HOME path)
// 2. load session file (by path)
// 3. load session file (by name)
func ReadConfigFiles(v *viper.Viper) (path string, err error) {
	home := getSessionHomeDir()
	v.AddConfigPath(".")
	v.AddConfigPath(home)

	// Load (non-session) preferences
	v.SetConfigName(SettingsGlobalName)

	if err := v.ReadInConfig(); err == nil {
		path = v.ConfigFileUsed()
		Logger.Infof("Loaded settings: %s", hideSensitiveInformationIfActive(path))
	}

	// Load session
	if _, err := os.Stat(globalFlagSessionFile); err == nil {
		// Load config by file path
		v.SetConfigFile(globalFlagSessionFile)
		if err := cliConfig.ReadConfig(globalFlagSessionFile); err != nil {
			Logger.Warnf("Could not read global settings file. file=%s, err=%s", globalFlagSessionFile, err)
		}
	} else {
		// Load config by name
		sessionName := "session"
		if globalFlagSessionFile != "" {
			sessionName = globalFlagSessionFile
		}

		if sessionName != "" {
			v.SetConfigName(sessionName)
		}
	}

	err = v.MergeInConfig()
	path = v.ConfigFileUsed()

	if err != nil {
		Logger.Debugf("Failed to merge config. %s", err)
	}

	Logger.Infof("Loaded session: %s", hideSensitiveInformationIfActive(path))

	return path, err
}

// bindFlags binds cobra flag values to cobra. This allows the user to set the default values via configuration or environment variables
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		settingsName := "settings.defaults." + f.Name
		if err := v.BindEnv(settingsName); err != nil {
			Logger.Warnf("Could not bind to environment variable. name=%s, err=%s", settingsName, err)
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(settingsName) {
			val := v.Get(settingsName)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				Logger.Warnf("Could not set flag. name=%s, err=%s", settingsName, err)
			}
		}
	})
}

func printCommand() {
	buf := bytes.Buffer{}

	escapeQuotesWindows := func(value string) string {
		return "'" + strings.ReplaceAll(value, "\"", `\"`) + "'"
	}
	escapeQuotesLinux := func(value string) string {
		return "\"" + strings.ReplaceAll(value, "\"", `\"`) + "\""
	}

	quoter := escapeQuotesLinux
	if runtime.GOOS == "windows" {
		quoter = escapeQuotesWindows
	}

	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, " ") {
			if strings.HasPrefix(arg, "-") {
				if index := strings.Index(arg, "="); index > -1 {
					buf.WriteString(arg[0:index] + "=")
					buf.WriteString(quoter(arg[index+1:]))
				} else {
					buf.WriteString(quoter(arg))
				}
			} else {
				buf.WriteString(quoter(arg))
			}
		} else {
			buf.WriteString(arg)
		}
		buf.WriteByte(' ')
	}
	Logger.Infof("command: c8y %s", strings.TrimSpace(buf.String()))
}

func initConfig() {
	logOptions := logger.Options{
		Level: zapcore.WarnLevel,
		Color: !globalFlagNoColor,
		Debug: globalFlagDebug,
	}
	if globalFlagProgressBar {
		logOptions.Silent = true
	} else {
		if globalFlagVerbose {
			logOptions.Level = zapcore.InfoLevel
		}
		if globalFlagDebug {
			logOptions.Level = zapcore.DebugLevel
		}
	}

	Logger = logger.NewLogger(module, logOptions)
	rootCmd.Logger = Logger

	c8y.Logger = Logger

	if logOptions.Level.Enabled(zapcore.InfoLevel) {
		printCommand()
	}

	if globalFlagSessionFile == "" && os.Getenv("C8Y_SESSION") != "" {
		globalFlagSessionFile = os.Getenv("C8Y_SESSION")
		if globalFlagSessionFile != "" {
			Logger.Printf("Using session environment variable: %s", hideSensitiveInformationIfActive(globalFlagSessionFile))
		}
	}

	// global session flag has precendence over use environment
	if !rootCmd.Flags().Changed("useEnv") {
		if globalFlagSessionFile == "" && os.Getenv("C8Y_USE_ENVIRONMENT") != "" {
			globalFlagUseEnv = true
		}
	}

	cliConfig = config.NewCliConfiguration(viper.GetViper(), SecureDataAccessor, getSessionHomeDir(), os.Getenv("C8Y_PASSPHRASE"))
	cliConfig.SetLogger(Logger)
	if err := loadConfiguration(); err != nil {
		Logger.Warnf("Loading configuration had errors. %s", err)
	}

	// only parse env variables if no explict config file is given
	if globalFlagUseEnv {
		Logger.Println("C8Y_USE_ENVIRONMENT is set. Environment variables can be used to override config settings")
		viper.AutomaticEnv()
	}

	// Proxy settings
	// Either use explicit proxy, ignore proxy, or use existing env variables
	// --proxy "http://10.0.0.1:8080"
	// --noProxy
	// HTTP_PROXY=http://10.0.0.1:8080
	// NO_PROXY=localhost,127.0.0.1
	explictProxy := rootCmd.Flags().Changed("proxy")
	explictNoProxy := rootCmd.Flags().Changed("noProxy")
	if explictNoProxy {
		Logger.Debug("using explicit noProxy setting")
		os.Setenv("HTTP_PROXY", "")
		os.Setenv("HTTPS_PROXY", "")
		os.Setenv("http_proxy", "")
		os.Setenv("https_proxy", "")
	} else {
		if explictProxy {
			Logger.Debugf("using explicit proxy [%s]", globalFlagProxy)

			globalFlagProxy = strings.TrimSpace(globalFlagProxy)

			os.Setenv("HTTP_PROXY", globalFlagProxy)
			os.Setenv("HTTPS_PROXY", globalFlagProxy)
			os.Setenv("http_proxy", globalFlagProxy)
			os.Setenv("https_proxy", globalFlagProxy)

		} else {
			proxyVars := []string{"HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "NO_PROXY", "no_proxy"}

			var proxySettings strings.Builder

			for _, name := range proxyVars {
				if v := os.Getenv(name); v != "" {
					proxySettings.WriteString(fmt.Sprintf(" %s [%s]", name, v))
				}
			}
			if proxySettings.Len() > 0 {
				Logger.Debugf("Using existing env variables.%s", proxySettings)
			}

		}
	}
}

func loadConfiguration() error {
	Logger.Debug("Loading configuration")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("c8y")

	//viper.AutomaticEnv()
	bindEnv(SettingsIncludeAllPageSize, 2000)
	bindEnv(SettingsDefaultBatchMaxWorkers, 5)
	bindEnv(SettingsIncludeAllDelayMS, 0)
	bindEnv(SettingsTemplatePath, "")

	// Mode
	bindEnv(SettingsModeEnableCreate, false)
	bindEnv(SettingsModeEnableUpdate, false)
	bindEnv(SettingsModeEnableDelete, false)
	bindEnv(SettingsEncryptionEnabled, false)
	bindEnv(SettingsModeCI, false)
	bindEnv(SettingsModeConfirmation, "PUT POST DELETE")

	// Storage options
	bindEnv(SettingsStorageStorePassword, true)
	bindEnv(SettingsStorageStoreCookies, true)

	// Activity log settings
	bindEnv(SettingsActivityLogEnabled, true)
	bindEnv(SettingsActivityLogPath, "")
	bindEnv(SettingsActivityLogMethodFilter, "GET PUT POST DELETE")

	return nil
}

func shouldConfirm(methods ...string) bool {
	if globalFlagConfirm {
		return true
	}

	if cliConfig.IsCIMode() || globalFlagForce || globalFlagDryRun {
		Logger.Debugf("no confirmation required. ci_mode=%v, force=%v, dry=%v", cliConfig.IsCIMode(), globalFlagForce, globalFlagDryRun)
		return false
	}

	if len(methods) == 0 {
		return true
	}

	confirmMethods := strings.ToUpper(viper.GetString(SettingsModeConfirmation))
	for _, method := range methods {
		if strings.Contains(confirmMethods, strings.ToUpper(method)) {
			Logger.Debugf("confirmation required due to method=%s", method)
			return true
		}
	}
	return false
}

func configureActivityLog() {
	disabled := !viper.GetBool(SettingsActivityLogEnabled)
	if globalFlagNoLog {
		disabled = true
	}
	options := activitylogger.Options{
		Disabled:     disabled,
		OutputFolder: viper.GetString(SettingsActivityLogPath),
		Methods:      strings.ToUpper(viper.GetString(SettingsActivityLogMethodFilter)),
	}

	if l, err := activitylogger.NewActivityLogger(options); err == nil {
		activityLogger = l
		if disabled {
			Logger.Info("activityLog is disabled")
		} else {
			Logger.Infof("activityLog path: %s", activityLogger.GetPath())
		}
	} else {
		Logger.Errorf("Failed to load activity logger. %s", err)
	}
}

func loadAuthentication(v *config.CliConfiguration, c *c8y.Client) error {
	cookies := v.GetCookies()

	if len(cookies) > 0 {
		c.SetCookies(cookies)
		c.AuthorizationMethod = c8y.AuthMethodOAuth2Internal
	}

	return nil
}

func readConfiguration(cmd *cobra.Command) error {

	globalFlagIncludeAllPageSize = viper.GetInt(SettingsIncludeAllPageSize)
	globalFlagBatchMaxWorkers = viper.GetInt(SettingsDefaultBatchMaxWorkers)
	globalFlagIncludeAllDelayMS = viper.GetInt64(SettingsIncludeAllDelayMS)
	globalFlagTemplatePath = viper.GetString(SettingsTemplatePath)

	globalModeEnableCreate = viper.GetBool(SettingsModeEnableCreate)
	globalModeEnableUpdate = viper.GetBool(SettingsModeEnableUpdate)
	globalModeEnableDelete = viper.GetBool(SettingsModeEnableDelete)

	globalStorageStoreCookies = viper.GetBool(SettingsStorageStoreCookies)
	globalStorageStorePassword = viper.GetBool(SettingsStorageStorePassword)

	globalCIMode = viper.GetBool(SettingsModeCI)

	Logger.Debugf("%s: %d", SettingsDefaultsPageSize, globalFlagPageSize)
	Logger.Debugf("%s: %d", SettingsIncludeAllPageSize, globalFlagIncludeAllPageSize)
	Logger.Debugf("%s: %d", SettingsIncludeAllDelayMS, globalFlagIncludeAllDelayMS)
	Logger.Debugf("%s: %s", SettingsTemplatePath, globalFlagTemplatePath)
	Logger.Debugf("%s: %t", SettingsModeEnableCreate, globalModeEnableCreate)
	Logger.Debugf("%s: %t", SettingsModeEnableUpdate, globalModeEnableUpdate)
	Logger.Debugf("%s: %t", SettingsModeEnableDelete, globalModeEnableDelete)
	Logger.Debugf("%s: %s", SettingsModeConfirmation, viper.GetString(SettingsModeConfirmation))

	return nil
}

func bindEnv(name string, defaultValue interface{}) {
	err := viper.BindEnv(name)
	viper.SetDefault(name, defaultValue)

	if err != nil {
		Logger.Warnf("Could not bind env variable. name=%s, err=%s", name, err)
	}
}

func newWebsocketDialer(ignoreProxySettings bool) *websocket.Dialer {
	dialer := &websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: false,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if ignoreProxySettings {
		dialer.Proxy = nil
	}

	return dialer
}

func newHTTPClient(ignoreProxySettings bool) *http.Client {
	// Default client ignores self signed certificates (to enable compatibility to the edge which uses self signed certs)
	defaultTransport := http.DefaultTransport.(*http.Transport)
	tr := &http.Transport{
		Proxy:                 defaultTransport.Proxy,
		DialContext:           defaultTransport.DialContext,
		MaxIdleConns:          defaultTransport.MaxIdleConns,
		IdleConnTimeout:       defaultTransport.IdleConnTimeout,
		ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
		TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	if ignoreProxySettings {
		tr.Proxy = nil
	}

	return &http.Client{
		Transport: tr,
	}
}

func hideSensitiveInformationIfActive(message string) string {

	if !strings.EqualFold(os.Getenv(c8y.EnvVarLoggerHideSensitive), "true") {
		return message
	}

	if os.Getenv("USERNAME") != "" {
		message = strings.ReplaceAll(message, os.Getenv("USERNAME"), "******")
	}

	if client != nil {
		message = strings.ReplaceAll(message, client.TenantName, "{tenant}")
		message = strings.ReplaceAll(message, client.Username, "{username}")
		message = strings.ReplaceAll(message, client.Password, "{password}")
		if client.BaseURL != nil {
			message = strings.ReplaceAll(message, strings.TrimRight(client.BaseURL.Host, "/"), "{host}")
		}
	}

	basicAuthMatcher := regexp.MustCompile(`(Basic\s+)[A-Za-z0-9=]+`)
	message = basicAuthMatcher.ReplaceAllString(message, "$1 {base64 tenant/username:password}")

	return message
}
