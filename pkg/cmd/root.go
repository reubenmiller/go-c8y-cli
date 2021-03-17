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
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/cli/safeexec"
	"github.com/gorilla/websocket"
	"github.com/reubenmiller/go-c8y-cli/internal/run"
	"github.com/reubenmiller/go-c8y-cli/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ydefaults"
	"github.com/reubenmiller/go-c8y-cli/pkg/clierrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/alias"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/alias/expand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/factory"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

// Logger is used to record the log messages which should be visible to the user when using the verbose flag
var Logger *logger.Logger
var activityLogger *activitylogger.ActivityLogger

// Console provides a thread safe way to write to the console output
var Console *console.Console

// Build data
// These variables should be set using the -ldflags "-X github.com/reubenmiller/go-c8y-cli/pkg/cmd.version=1.0.0" when running go build
var buildVersion string
var buildBranch string

const (
	module = "c8yapi"
)

func init() {
	Logger = logger.NewLogger(module, logger.Options{})
	rootCmd = NewRootCmd()
	// set seed for random generation
	rand.Seed(time.Now().UTC().UnixNano())
}

type baseCmd struct {
	cmd *cobra.Command
}

func (c *baseCmd) getCommand() *cobra.Command {
	// mark local flags so the appear in completions before global flags
	completion.WithOptions(
		c.cmd,
		completion.MarkLocalFlag(),
	)
	return c.cmd
}

func newBaseCmd(cmd *cobra.Command) *baseCmd {
	return &baseCmd{cmd: cmd}
}

type RootCmd struct {
	cobra.Command
	Logger *logger.Logger
	useEnv bool

	dataView *dataview.DataView
}

func (c *RootCmd) DryRunHandler(options *c8y.RequestOptions, req *http.Request) {

	if !cliConfig.DryRun() {
		return
	}
	if req == nil {
		Logger.Warn("Response is nil")
		return
	}
	w := c.Command.ErrOrStderr()
	if cliConfig.WithError() {
		w = c.Command.OutOrStdout()
	}

	PrintRequestDetails(w, nil, req)
}
func (c *RootCmd) createCumulocityClient() {
	c.Logger.Debug("Creating c8y client")
	httpClient := newHTTPClient(cliConfig.IgnoreProxy())

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

	client = c8y.NewClient(
		httpClient,
		formatHost(cliConfig.GetHost()),
		cliConfig.GetTenant(),
		cliConfig.GetUsername(),
		cliConfig.MustGetPassword(),
		true,
	)

	client.SetRequestOptions(c8y.DefaultRequestOptions{
		DryRun:        cliConfig.DryRun(),
		DryRunHandler: c.DryRunHandler,
	})

	// load authentication
	if err := loadAuthentication(cliConfig, client); err != nil {
		Logger.Warnf("Could not load authentication. %s", err)
	}

	timeout := cliConfig.RequestTimeout()
	Logger.Debugf("timeout: %0.3f", timeout)

	// Should we use the tenant in the name or not
	if viper.IsSet("useTenantPrefix") {
		client.UseTenantInUsername = viper.GetBool("useTenantPrefix")
	}

	// set output format
	Console.Format = cliConfig.GetOutputFormat()
	Console.Colorized = !globalFlagNoColor
	Console.Compact = cliConfig.CompactJSON()
	Console.Disabled = globalFlagProgressBar && isTerminal()

	// Add the realtime client
	client.Realtime = c8y.NewRealtimeClient(
		client.BaseURL.String(),
		newWebsocketDialer(cliConfig.IgnoreProxy()),
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

func NewRootCmd() *RootCmd {
	command := &RootCmd{
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

var rootCmd *RootCmd

var (
	client                       *c8y.Client
	cliConfig                    *config.Config
	globalFlagVerbose            bool
	globalFlagDebug              bool
	globalFlagProgressBar        bool
	globalFlagNoColor            bool
	globalFlagSessionFile        string
	globalFlagUseEnv             bool
	globalFlagNoLog              bool
	globalFlagActivityLogMessage string
)

// SettingsGlobalName name of the settings file (without extension)
const SettingsGlobalName = "settings"

func (c *RootCmd) checkCommandError(err error) {

	w := ioutil.Discard
	if cliConfig != nil && cliConfig.WithError() {
		w = rootCmd.OutOrStdout()
	}

	if errors.Is(err, clierrors.ErrNoMatchesFound) {
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
		silentStatusCodes := ""
		if cliConfig != nil {
			silentStatusCodes = cliConfig.GetSilentStatusCodes()
		}
		if !cErr.IsSilent() && !strings.Contains(silentStatusCodes, fmt.Sprintf("%d", cErr.StatusCode)) {
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

func (c *RootCmd) checkSessionExists(cmd *cobra.Command, args []string) error {

	cmdStr := cmd.Use
	if cmd.HasParent() && cmd.Parent().Use != "c8y" {
		cmdStr = cmd.Parent().Use + " " + cmdStr
	}
	c.Logger.Debugf("command str: %s", cmdStr)

	if globalFlagSessionFile == "" || !(strings.HasPrefix(cmdStr, "sessions list") || c.Flags().Changed("session")) {
		c.useEnv = true
	}

	// config/env binding
	if err := cliConfig.BindPFlag(rootCmd.PersistentFlags()); err != nil {
		Logger.Warningf("Some configuration binding failed. %s", err)
	}

	c.createCumulocityClient()

	// only setup activity log after the global config
	configureActivityLog()
	activityLogger.LogCommand(cmd, args, cmdStr, globalFlagActivityLogMessage)

	// load views
	if views, err := dataview.NewDataView(".*", ".json", Logger, cliConfig.GetViewPaths()...); err == nil {
		c.dataView = views
	}

	if !cmdutil.IsAuthCheckEnabled(cmd) {
		return nil
	}

	localCmds := []string{
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

func setArgs() ([]string, error) {
	expandedArgs := []string{}
	if len(os.Args) > 0 {
		expandedArgs = os.Args[1:]
	}
	cmd, _, err := rootCmd.Traverse(expandedArgs)
	if err != nil || cmd == rootCmd.Root() {
		originalArgs := expandedArgs
		isShell := false

		v := viper.GetViper()
		aliases := v.GetStringMapString("settings.aliases")
		expandedArgs, isShell, err = expand.ExpandAlias(aliases, os.Args, nil)
		if err != nil {
			Logger.Errorf("failed to process aliases:  %s", err)
			return nil, err
		}

		Logger.Debugf("%v -> %v", originalArgs, expandedArgs)

		if isShell {
			exe, err := safeexec.LookPath(expandedArgs[0])
			if err != nil {
				Logger.Errorf("failed to run external command: %s", err)
				return nil, err
			}

			externalCmd := exec.Command(exe, expandedArgs[1:]...)
			externalCmd.Stderr = os.Stderr
			externalCmd.Stdout = os.Stdout
			externalCmd.Stdin = os.Stdin
			preparedCmd := run.PrepareCmd(externalCmd)

			err = preparedCmd.Run()
			if err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					return nil, cmderrors.NewUserErrorWithExitCode(ee.ExitCode(), ee)
				}

				Logger.Errorf("failed to run external command: %s", err)
				return nil, err
			}
		}
	}
	return expandedArgs, nil
}

func isTerminal() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}
	return false
}

func getOutputHeaders(input []string) (headers []byte) {
	if !Console.IsCSV() || !Console.WithCSVHeader() || len(input) == 0 {
		Logger.Debugf("Ignoring csv headers: isCSV=%v, WithHeader=%v", Console.IsCSV(), Console.WithCSVHeader())
		return
	}
	if len(input) > 0 {
		return []byte(input[0] + "\n")
	}

	// TODO: improve detection by parsing more lines to find column names (if more lines are available)
	columns := make([][]byte, 0)
	for _, v := range cliConfig.GetJSONSelect() {
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

// ConfigureRootCmd initializes the configuration manager and c8y client
func (c *RootCmd) ConfigureRootCmd() {
	// config file
	cobra.OnInitialize(initConfig)

	c.PersistentFlags().StringVar(&globalFlagSessionFile, "session", "", "Session configuration")

	isTerm := isTerminal()
	Console = console.NewConsole(rootCmd.OutOrStdout(), getOutputHeaders)
	defaultOutputFormat := "json"
	defaultView := config.ViewsNone
	if isTerm {
		defaultOutputFormat = "table"
		defaultView = config.ViewsAll
	}

	// TODO: Migrate to new root command
	// cmd factory
	configFunc := func() (*config.Config, error) {
		if cliConfig == nil {
			return nil, fmt.Errorf("config is missing")
		}
		return cliConfig, nil
	}
	clientFunc := func() (*c8y.Client, error) {
		if client == nil {
			return nil, fmt.Errorf("client is missing")
		}
		return client, nil
	}
	loggerFunc := func() (*logger.Logger, error) {
		if Logger == nil {
			return nil, fmt.Errorf("logger is missing")
		}
		return Logger, nil
	}
	cmdFactory := factory.New(buildVersion, configFunc, clientFunc, loggerFunc)

	// customRootCmd := root.NewCmdRoot(cmdFactory, buildVersion, "")
	// rootCmd.AddCommand(customRootCmd)

	// Global flags
	c.PersistentFlags().BoolVarP(&globalFlagVerbose, "verbose", "v", false, "Verbose logging")
	c.PersistentFlags().IntP("pageSize", "p", c8ydefaults.PageSize, "Maximum results per page")
	c.PersistentFlags().Int64("currentPage", 0, "Current page size which should be returned")
	c.PersistentFlags().Int64("totalPages", 0, "Total number of pages to get")
	c.PersistentFlags().Bool("includeAll", false, "Include all results by iterating through each page")
	c.PersistentFlags().BoolP("withTotalPages", "t", false, "Include all results")
	c.PersistentFlags().BoolP("compact", "c", !isTerm, "Compact instead of pretty-printed output. Pretty print is the default if output is the terminal")
	c.PersistentFlags().Bool("noAccept", false, "Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect")
	c.PersistentFlags().Bool("dry", false, "Dry run. Don't send any data to the server")
	c.PersistentFlags().String("dryFormat", "markdown", "Dry run output format. i.e. json, dump, markdown or curl")
	c.PersistentFlags().BoolVar(&globalFlagProgressBar, "progress", false, "Show progress bar. This will also disable any other verbose output")
	c.PersistentFlags().BoolVarP(&globalFlagNoColor, "noColor", "M", !isTerm, "Don't use colors when displaying log entries on the console")
	c.PersistentFlags().BoolVar(&globalFlagUseEnv, "useEnv", false, "Allow loading Cumulocity session setting from environment variables")
	c.PersistentFlags().BoolP("raw", "r", false, "Raw values")
	c.PersistentFlags().String("proxy", "", "Proxy setting, i.e. http://10.0.0.1:8080")
	c.PersistentFlags().Bool("noProxy", false, "Ignore the proxy settings")
	c.PersistentFlags().Bool("withError", false, "Errors will be printed on stdout instead of stderr")

	// Activity log
	c.PersistentFlags().BoolVar(&globalFlagNoLog, "noLog", false, "Disables the activity log for the current command")
	c.PersistentFlags().StringVarP(&globalFlagActivityLogMessage, "logMessage", "l", "", "Add custom message to the activity log")
	c.PersistentFlags().BoolVar(&globalFlagDebug, "debug", false, "Set very verbose log messages")

	// Concurrency
	c.PersistentFlags().Int("workers", 1, "Number of workers")
	c.PersistentFlags().Int64("maxJobs", 100, "Maximum number of jobs. 0 = unlimited (use with caution!)")

	// viper.BindPFlag("settings.defaults.maxJobs", c.PersistentFlags().Lookup("maxJobs"))
	c.PersistentFlags().Int("delay", 1000, "delay in milliseconds after each request")
	c.PersistentFlags().Int("abortOnErrors", 10, "Abort batch when reaching specified number of errors")

	// Error handling
	c.PersistentFlags().String("silentStatusCodes", "", "Status codes which will not print out an error message")

	c.PersistentFlags().Bool("flatten", false, "flatten")
	c.PersistentFlags().StringSlice("filter", nil, "filter")
	c.PersistentFlags().StringArray("select", nil, "select")
	c.PersistentFlags().String("view", defaultView, "View option")
	c.PersistentFlags().Float64("timeout", float64(10*60), "Timeout in seconds")

	// output
	c.PersistentFlags().StringP("output", "o", defaultOutputFormat, "Output format i.e. table, json, csv, csvheader")
	c.PersistentFlags().String("outputFile", "", "Save JSON output to file (after select)")
	c.PersistentFlags().String("outputFileRaw", "", "Save raw response to file")

	// confirmation
	c.PersistentFlags().BoolP("force", "f", false, "Do not prompt for confirmation")
	c.PersistentFlags().Bool("prompt", false, "Prompt for confirmation")
	c.PersistentFlags().String("confirmText", "", "Custom confirmation text")

	completion.WithOptions(
		&c.Command,
		completion.WithValidateSet("dryFormat", "json", "dump", "curl", "markdown"),
		completion.WithValidateSet("output", "json", "table", "csv", "csvheader"),
		completion.WithValidateSet("view", config.ViewsNone, config.ViewsAll),
	)

	c.AddCommand(NewCompletionsCmd().getCommand())
	c.AddCommand(NewVersionCmd().getCommand())

	c.AddCommand(NewRealtimeCmd().getCommand())
	c.AddCommand(NewSessionsRootCmd().getCommand())

	// generic commands
	c.AddCommand(NewGetGenericRestCmd().getCommand())

	// template commands
	c.AddCommand(NewTemplateRootCmd().getCommand())

	// settings commands
	c.AddCommand(NewSettingsRootCmd().getCommand())

	// alias commands
	c.AddCommand(alias.NewCmdAlias(cmdFactory))

	// Auto generated commands

	// agents commands
	agents := NewAgentsRootCmd().getCommand()
	agents.AddCommand(NewGetAgentCollectionCmd().getCommand())
	c.AddCommand(agents)

	// alarms commands
	alarms := NewAlarmsRootCmd().getCommand()
	alarms.AddCommand(NewSubscribeAlarmCmd().getCommand())
	c.AddCommand(alarms)

	// applications commands
	applications := NewApplicationsRootCmd().getCommand()
	applications.AddCommand(NewNewHostedApplicationCmd().getCommand())
	c.AddCommand(applications)

	// auditRecords commands
	c.AddCommand(NewAuditRecordsRootCmd().getCommand())

	// binaries commands
	c.AddCommand(NewBinariesRootCmd().getCommand())

	// bulkOperations commands
	c.AddCommand(NewBulkOperationsRootCmd().getCommand())

	// currentApplication commands
	c.AddCommand(NewCurrentApplicationRootCmd().getCommand())

	// currentUser commands
	c.AddCommand(newCurrentUserRootCmd().getCommand())

	// databroker commands
	c.AddCommand(NewDatabrokerRootCmd().getCommand())

	// deviceCredentials commands
	c.AddCommand(NewDeviceCredentialsRootCmd().getCommand())

	// devices commands
	devices := NewDevicesRootCmd().getCommand()
	devices.AddCommand(NewGetDeviceCollectionCmd(cmdFactory).getCommand())
	devices.AddCommand(NewGetDeviceGroupCollectionCmd().getCommand())
	c.AddCommand(devices)

	// operations commands
	operations := NewOperationsRootCmd().getCommand()
	operations.AddCommand(NewSubscribeOperationCmd().getCommand())
	c.AddCommand(operations)

	// events commands
	events := NewEventsRootCmd().getCommand()
	events.AddCommand(NewSubscribeEventCmd().getCommand())
	c.AddCommand(events)

	// identity commands
	c.AddCommand(NewIdentityRootCmd().getCommand())

	// inventory commands
	inventory := NewInventoryRootCmd().getCommand()
	inventory.AddCommand(NewSubscribeManagedObjectCmd().getCommand())
	inventory.AddCommand(NewQueryManagedObjectCollectionCmd().getCommand())
	c.AddCommand(inventory)

	// inventoryReferences commands
	c.AddCommand(NewInventoryReferencesRootCmd().getCommand())

	// measurements commands
	measurements := NewMeasurementsRootCmd().getCommand()
	measurements.AddCommand(NewSubscribeMeasurementCmd().getCommand())
	c.AddCommand(measurements)

	// microservices commands
	microservices := NewMicroservicesRootCmd().getCommand()
	microservices.AddCommand(NewNewMicroserviceCmd().getCommand())
	microservices.AddCommand(NewNewServiceUserCmd().getCommand())
	microservices.AddCommand(NewGetServiceUserCmd().getCommand())
	c.AddCommand(microservices)

	// retentionRules commands
	c.AddCommand(NewRetentionRulesRootCmd().getCommand())

	// systemOptions commands
	c.AddCommand(NewSystemOptionsRootCmd().getCommand())

	// tenantOptions commands
	c.AddCommand(NewTenantOptionsRootCmd().getCommand())

	// tenants commands
	c.AddCommand(NewTenantsRootCmd().getCommand())

	// tenantStatistics commands
	c.AddCommand(NewTenantStatisticsRootCmd().getCommand())

	// users commands
	c.AddCommand(NewUsersRootCmd().getCommand())

	// userGroups commands
	c.AddCommand(NewUserGroupsRootCmd().getCommand())

	// userReferences commands
	c.AddCommand(NewUserReferencesRootCmd().getCommand())

	// userRoles commands
	c.AddCommand(NewUserRolesRootCmd().getCommand())

	// Handle errors (not in cobra libary)
	c.SilenceErrors = true

}

// Execute runs the root command
func Execute() {
	rootCmd.ConfigureRootCmd()
	executeRootCmd()
}

func executeRootCmd() {
	// Expand any aliases
	if globalFlagSessionFile == "" && os.Getenv("C8Y_SESSION") != "" {
		globalFlagSessionFile = os.Getenv("C8Y_SESSION")
	}
	_, _ = ReadConfigFiles(viper.GetViper())
	expandedArgs, err := setArgs()
	if err != nil {
		Logger.Errorf("Could not expand aliases. %s", err)
	}
	Logger.Debugf("Expanded args: %v", expandedArgs)
	rootCmd.SetArgs(expandedArgs)

	if err := rootCmd.Execute(); err != nil {

		rootCmd.checkCommandError(err)

		if cErr, ok := err.(cmderrors.CommandError); ok {
			os.Exit(cErr.ExitCode)
		}
		if errors.Is(err, clierrors.ErrNoMatchesFound) {
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
		if cliConfig != nil {
			if err := cliConfig.ReadConfig(globalFlagSessionFile); err != nil {
				Logger.Warnf("Could not read global settings file. file=%s, err=%s", globalFlagSessionFile, err)
			}
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
	cliConfig = config.NewConfig(viper.GetViper(), getSessionHomeDir(), os.Getenv("C8Y_PASSPHRASE"))
	cliConfig.SetLogger(Logger)

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
	proxy := cliConfig.Proxy()
	noProxy := cliConfig.IgnoreProxy()
	if noProxy {
		Logger.Debug("using explicit noProxy setting")
		os.Setenv("HTTP_PROXY", "")
		os.Setenv("HTTPS_PROXY", "")
		os.Setenv("http_proxy", "")
		os.Setenv("https_proxy", "")
	} else {
		if proxy != "" {
			Logger.Debugf("using explicit proxy [%s]", proxy)

			os.Setenv("HTTP_PROXY", proxy)
			os.Setenv("HTTPS_PROXY", proxy)
			os.Setenv("http_proxy", proxy)
			os.Setenv("https_proxy", proxy)

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

func configureActivityLog() {
	disabled := !cliConfig.ActivityLogEnabled()
	if globalFlagNoLog {
		disabled = true
	}
	options := activitylogger.Options{
		Disabled:     disabled,
		OutputFolder: cliConfig.GetActivityLogPath(),
		Methods:      strings.ToUpper(cliConfig.GetActivityLogMethodFilter()),
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

func loadAuthentication(v *config.Config, c *c8y.Client) error {
	cookies := v.GetCookies()

	if len(cookies) > 0 {
		c.SetCookies(cookies)
		c.AuthorizationMethod = c8y.AuthMethodOAuth2Internal
	}

	return nil
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
