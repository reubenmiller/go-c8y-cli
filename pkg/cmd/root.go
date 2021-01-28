package cmd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Logger is used to record the log messages which should be visible to the user when using the verbose flag
var Logger *logger.Logger

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
	Logger = logger.NewDummyLogger(module)
	SecureDataAccessor = encrypt.NewSecureData("{encrypted}")
	rootCmd = newC8yCmd()
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
	client *c8y.Client
	Logger *logger.Logger
	useEnv bool
}

func (c *c8yCmd) createCumulocityClient() {
	c.Logger.Info("Creating c8y client")
	httpClient := newHTTPClient(globalFlagNoProxy)

	// Only bind when not setting the session
	if c.useEnv {
		c.Logger.Info("Binding authorization environment variables")
		cliConfig.BindAuthorization()
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

	// load authentication
	loadAuthentication(cliConfig, client)

	//
	// Timeout setting preference
	// 1. User provideds --timeout argument
	// 2. "timeout" is set in the session.json file (and value is greater than 0)
	// 3. C8Y_TIMEOUT is set and greater than 0
	if !rootCmd.Flags().Changed("timeout") {
		timeout := viper.GetUint("timeout")
		if timeout > 0 {
			globalFlagTimeout = timeout
			Logger.Debugf("timeout: %v", timeout)
		}
	}

	// Should we use the tenant in the name or not
	if viper.IsSet("useTenantPrefix") {
		client.UseTenantInUsername = viper.GetBool("useTenantPrefix")
	}

	// Logger.Printf("Use tenant prefix: %v", client.UseTenantInUsername)

	// read additional configuration
	readConfiguration()

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
	client                       *c8y.Client
	cliConfig                    *config.CliConfiguration
	globalFlagPageSize           int
	globalFlagIncludeAllPageSize int
	globalFlagBatchMaxWorkers    int
	globalFlagCurrentPage        int64
	globalFlagTotalPages         int64
	globalFlagIncludeAll         bool
	globalFlagIncludeAllDelayMS  int64
	globalFlagVerbose            bool
	globalFlagWithTotalPages     bool
	globalFlagPrettyPrint        bool
	globalFlagDryRun             bool
	globalFlagNoColor            bool
	globalFlagSessionFile        string
	globalFlagConfigFile         string
	globalFlagOutputFile         string
	globalFlagUseEnv             bool
	globalFlagRaw                bool
	globalFlagProxy              string
	globalFlagNoProxy            bool
	globalFlagTimeout            uint
	globalFlagUseTenantPrefix    bool
	globalUseNonDefaultPageSize  bool
	globalFlagTemplatePath       string

	globalModeEnableCreate bool
	globalModeEnableUpdate bool
	globalModeEnableDelete bool
	globalModeEnableBatch  bool
	globalCIMode           bool
)

// CumulocityDefaultPageSize is the default page size used by Cumulocity
const CumulocityDefaultPageSize int = 5

const (
	// SettingsIncludeAllPageSize property name used to control the default page size when using includeAll parameter
	SettingsIncludeAllPageSize string = "settings.includeAll.pageSize"

	// SettingsIncludeAllDelayMS property name used to control the delay between fetching the next page
	SettingsIncludeAllDelayMS string = "settings.includeAll.delayMS"

	// SettingsDefaultPageSize property name used to control the default page size
	SettingsDefaultPageSize string = "settings.default.pageSize"

	// SettingsDefaultBatchMaxWorkers property name used to control the hard limit on the maximum workers used in batch operations
	SettingsDefaultBatchMaxWorkers string = "settings.default.batchMaximumWorkers"

	// SettingsConfigPath configuration path
	SettingsConfigPath string = "settings.path"

	// SettingsTemplatePath template path where the template files are located
	SettingsTemplatePath string = "settings.template.path"

	// SettingsModeEnableCreate enables create (post) commands
	SettingsModeEnableCreate string = "settings.mode.enableCreate"

	// SettingsModeEnableBatch enables batch commands
	SettingsModeEnableBatch string = "settings.mode.enableBatch"

	// SettingsModeEnableUpdate enables update commands
	SettingsModeEnableUpdate string = "settings.mode.enableUpdate"

	// SettingsModeEnableDelete enables delete commands
	SettingsModeEnableDelete string = "settings.mode.enableDelete"

	// SettingsEncryptionEnabled enables encryption when storing sensitive session data
	SettingsEncryptionEnabled string = "settings.encryption.enabled"

	// SettingsModeCI enable continuous integration mode (this will enable all commands)
	SettingsModeCI string = "settings.ci"
)

// SettingsGlobalName name of the settings file (without extension)
const SettingsGlobalName = "settings"

func (c *c8yCmd) checkCommandError(err error) {

	if cErr, ok := err.(commandError); ok {
		if cErr.statusCode == 403 || cErr.statusCode == 401 {
			c.Logger.Error(fmt.Sprintf("Authentication failed (statusCode=%d). Try to run set-session again, or check the password", cErr.statusCode))
		}

		// format errors as json messages
		// only log users errors
		if !cErr.isSilent() {
			message := ""
			if cErr.statusCode != 0 {
				message = fmt.Sprintf(`{"error":"commandError","message":"%s","statusCode":%d}`, err, cErr.statusCode)
			} else {
				message = fmt.Sprintf(`{"error":"commandError","message":"%s"}`, err)
			}
			rootCmd.PrintErrln(strings.ReplaceAll(message, "\n", ""))
		}
	} else {
		// unexpected error
		c.Logger.Errorf("%s", err)
		message := fmt.Sprintf(`{"error":"commandError","message":"%s"}`, err)
		rootCmd.PrintErrln(strings.ReplaceAll(message, "\n", ""))
	}
}

func (c c8yCmd) showLoginUsage() {
	if globalFlagSessionFile != "" {
		fmt.Printf("c8y sessions login")
	} else {
		fmt.Printf("export C8Y_SESSION=$(c8y sessions list)")
	}
}

func (c *c8yCmd) checkSessionExists(cmd *cobra.Command, args []string) error {

	cmdStr := cmd.Use
	if cmd.HasParent() && cmd.Parent().Use != "c8y" {
		cmdStr = cmd.Parent().Use + " " + cmdStr
	}
	c.Logger.Infof("command str: %s", cmdStr)

	if globalFlagSessionFile == "" || !(strings.HasPrefix(cmdStr, "sessions list") || strings.HasPrefix(cmdStr, "sessions checkPassphrase") || c.Flags().Changed("session")) {
		c.useEnv = true
	}
	c.createCumulocityClient()

	localCmds := []string{
		"completion",
		"sessions",
		"template",
		"version",
		"tenants getID",
		"tenants getId",
	}

	for i := range localCmds {
		if strings.HasPrefix(cmdStr, localCmds[i]) {
			return nil
		}
	}

	if client == nil {
		return newSystemError("Client failed to load")
	}
	if client.BaseURL == nil || client.BaseURL.Host == "" {
		return newUserErrorWithExitCode(102, "A c8y session has not been loaded. Please create or activate a session and try again")
	}

	return nil
}

// Execute runs the root command and initializes the configuration manager and c8y client
func Execute() {
	// config file
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&globalFlagSessionFile, "session", "", "Session configuration")

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&globalFlagVerbose, "verbose", "v", false, "Verbose logging")
	rootCmd.PersistentFlags().IntVar(&globalFlagPageSize, "pageSize", 5, "Maximum results per page")
	rootCmd.PersistentFlags().Int64Var(&globalFlagCurrentPage, "currentPage", 0, "Current page size which should be returned")
	rootCmd.PersistentFlags().Int64Var(&globalFlagTotalPages, "totalPages", 0, "Total number of pages to get")
	rootCmd.PersistentFlags().BoolVar(&globalFlagIncludeAll, "includeAll", false, "Include all results by iterating through each page")
	rootCmd.PersistentFlags().BoolVar(&globalFlagWithTotalPages, "withTotalPages", false, "Include all results")
	rootCmd.PersistentFlags().BoolVar(&globalFlagPrettyPrint, "pretty", true, "Pretty print the json responses")
	rootCmd.PersistentFlags().BoolVar(&globalFlagDryRun, "dry", false, "Dry run. Don't send any data to the server")
	rootCmd.PersistentFlags().BoolVar(&globalFlagNoColor, "noColor", false, "Don't use colors when displaying log entries on the console")
	rootCmd.PersistentFlags().BoolVar(&globalFlagUseEnv, "useEnv", false, "Allow loading Cumulocity session setting from environment variables")
	rootCmd.PersistentFlags().BoolVar(&globalFlagRaw, "raw", false, "Raw values")
	rootCmd.PersistentFlags().StringVar(&globalFlagProxy, "proxy", "", "Proxy setting, i.e. http://10.0.0.1:8080")
	rootCmd.PersistentFlags().BoolVar(&globalFlagNoProxy, "noProxy", false, "Ignore the proxy settings")

	rootCmd.PersistentFlags().StringVar(&globalFlagOutputFile, "outputFile", "", "Output file")

	rootCmd.PersistentFlags().StringSlice("filter", nil, "filter")
	rootCmd.PersistentFlags().StringSlice("select", nil, "select")
	rootCmd.PersistentFlags().String("format", "", "format")
	rootCmd.PersistentFlags().UintVarP(&globalFlagTimeout, "timeout", "t", 10*60*1000, "Timeout in milliseconds")

	// Map settings to flags, allowing the user to set the own default settings
	viper.BindPFlag(SettingsDefaultPageSize, rootCmd.PersistentFlags().Lookup("pageSize"))
	// viper.BindPFlag(SettingsConfigPath, rootCmd.PersistentFlags().Lookup("config"))

	// TODO: Make flags case-insensitive
	// rootCmd.PersistentFlags().SetNormalizeFunc(flagNormalizeFunc)

	rootCmd.AddCommand(newCompletionsCmd().getCommand())
	rootCmd.AddCommand(newVersionCmd().getCommand())

	rootCmd.AddCommand(newRealtimeCmd().getCommand())
	rootCmd.AddCommand(newSessionsRootCmd().getCommand())

	// generic commands
	rootCmd.AddCommand(newGetGenericRestCmd().getCommand())

	// template commands
	rootCmd.AddCommand(newTemplateRootCmd().getCommand())

	// settings commands
	rootCmd.AddCommand(newSettingsRootCmd().getCommand())

	// Auto generated commands

	// agents commands
	agents := newAgentsRootCmd().getCommand()
	agents.AddCommand(newGetAgentCollectionCmd().getCommand())
	rootCmd.AddCommand(agents)

	// alarms commands
	alarms := newAlarmsRootCmd().getCommand()
	alarms.AddCommand(newSubscribeAlarmCmd().getCommand())
	rootCmd.AddCommand(alarms)

	// applications commands
	applications := newApplicationsRootCmd().getCommand()
	applications.AddCommand(newNewHostedApplicationCmd().getCommand())
	rootCmd.AddCommand(applications)

	// auditRecords commands
	rootCmd.AddCommand(newAuditRecordsRootCmd().getCommand())

	// batch commands
	rootCmd.AddCommand(newBatchRootCmd().getCommand())

	// binaries commands
	rootCmd.AddCommand(newBinariesRootCmd().getCommand())

	// bulkOperations commands
	rootCmd.AddCommand(newBulkOperationsRootCmd().getCommand())

	// currentApplication commands
	rootCmd.AddCommand(newCurrentApplicationRootCmd().getCommand())

	// databroker commands
	rootCmd.AddCommand(newDatabrokerRootCmd().getCommand())

	// deviceCredentials commands
	rootCmd.AddCommand(newDeviceCredentialsRootCmd().getCommand())

	// devices commands
	devices := newDevicesRootCmd().getCommand()
	devices.AddCommand(newGetDeviceCollectionCmd().getCommand())
	devices.AddCommand(newGetDeviceGroupCollectionCmd().getCommand())
	rootCmd.AddCommand(devices)

	// operations commands
	operations := newOperationsRootCmd().getCommand()
	operations.AddCommand(newSubscribeOperationCmd().getCommand())
	rootCmd.AddCommand(operations)

	// events commands
	events := newEventsRootCmd().getCommand()
	events.AddCommand(newSubscribeEventCmd().getCommand())
	rootCmd.AddCommand(events)

	// identity commands
	rootCmd.AddCommand(newIdentityRootCmd().getCommand())

	// inventory commands
	inventory := newInventoryRootCmd().getCommand()
	inventory.AddCommand(newSubscribeManagedObjectCmd().getCommand())
	inventory.AddCommand(newQueryManagedObjectCollectionCmd().getCommand())
	rootCmd.AddCommand(inventory)

	// inventoryReferences commands
	rootCmd.AddCommand(newInventoryReferencesRootCmd().getCommand())

	// measurements commands
	measurements := newMeasurementsRootCmd().getCommand()
	measurements.AddCommand(newSubscribeMeasurementCmd().getCommand())
	rootCmd.AddCommand(measurements)

	// microservices commands
	microservices := newMicroservicesRootCmd().getCommand()
	microservices.AddCommand(newNewMicroserviceCmd().getCommand())
	microservices.AddCommand(newNewServiceUserCmd().getCommand())
	microservices.AddCommand(newGetServiceUserCmd().getCommand())
	rootCmd.AddCommand(microservices)

	// retentionRules commands
	rootCmd.AddCommand(newRetentionRulesRootCmd().getCommand())

	// systemOptions commands
	rootCmd.AddCommand(newSystemOptionsRootCmd().getCommand())

	// tenantOptions commands
	rootCmd.AddCommand(newTenantOptionsRootCmd().getCommand())

	// tenants commands
	rootCmd.AddCommand(newTenantsRootCmd().getCommand())

	// tenantStatistics commands
	rootCmd.AddCommand(newTenantStatisticsRootCmd().getCommand())

	// users commands
	rootCmd.AddCommand(newUsersRootCmd().getCommand())

	// userGroups commands
	rootCmd.AddCommand(newUserGroupsRootCmd().getCommand())

	// userReferences commands
	rootCmd.AddCommand(newUserReferencesRootCmd().getCommand())

	// userRoles commands
	rootCmd.AddCommand(newUserRolesRootCmd().getCommand())

	// Handle errors (not in cobra libary)
	rootCmd.SilenceErrors = true

	if err := rootCmd.Execute(); err != nil {
		rootCmd.checkCommandError(err)

		if cErr, ok := err.(commandError); ok {
			os.Exit(cErr.exitCode)
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
		cliConfig.ReadConfig(globalFlagSessionFile)
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

func initConfig() {
	// Set logging
	if globalFlagVerbose || globalFlagDryRun {
		Logger = logger.NewLogger(module, !globalFlagNoColor)
	} else {
		// Disable log messages
		Logger = logger.NewDummyLogger(module)
		c8y.SilenceLogger()
	}
	rootCmd.Logger = Logger

	if globalFlagSessionFile == "" && os.Getenv("C8Y_SESSION") != "" {
		globalFlagSessionFile = os.Getenv("C8Y_SESSION")
		if globalFlagSessionFile != "" {
			Logger.Printf("Using session environment variable: %s\n", hideSensitiveInformationIfActive(globalFlagSessionFile))
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
	loadConfiguration()

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
	Logger.Info("Loading configuration")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("c8y")
	//viper.AutomaticEnv()
	bindEnv(SettingsIncludeAllPageSize, 2000)
	bindEnv(SettingsDefaultPageSize, CumulocityDefaultPageSize)
	bindEnv(SettingsDefaultBatchMaxWorkers, 5)
	bindEnv(SettingsIncludeAllDelayMS, 0)
	bindEnv(SettingsTemplatePath, "")
	bindEnv(SettingsModeEnableCreate, false)
	bindEnv(SettingsModeEnableUpdate, false)
	bindEnv(SettingsModeEnableDelete, false)
	bindEnv(SettingsEncryptionEnabled, false)
	bindEnv(SettingsModeCI, false)

	return nil
}

func loadAuthentication(v *config.CliConfiguration, c *c8y.Client) error {
	cookies := v.GetCookies()

	if len(cookies) > 0 {
		c.SetCookies(cookies)
		c.AuthorizationMethod = c8y.AuthMethodOAuth2Internal
	}

	return nil
}

func readConfiguration() error {

	globalFlagIncludeAllPageSize = viper.GetInt(SettingsIncludeAllPageSize)
	globalFlagPageSize = viper.GetInt(SettingsDefaultPageSize)
	globalFlagBatchMaxWorkers = viper.GetInt(SettingsDefaultBatchMaxWorkers)
	globalFlagIncludeAllDelayMS = viper.GetInt64(SettingsIncludeAllDelayMS)
	globalFlagTemplatePath = viper.GetString(SettingsTemplatePath)

	globalModeEnableBatch = viper.GetBool(SettingsModeEnableBatch)
	globalModeEnableCreate = viper.GetBool(SettingsModeEnableCreate)
	globalModeEnableUpdate = viper.GetBool(SettingsModeEnableUpdate)
	globalModeEnableDelete = viper.GetBool(SettingsModeEnableDelete)

	globalCIMode = viper.GetBool(SettingsModeCI)

	Logger.Infof("%s: %d", SettingsDefaultPageSize, globalFlagPageSize)
	Logger.Infof("%s: %d", SettingsIncludeAllPageSize, globalFlagIncludeAllPageSize)
	Logger.Infof("%s: %d", SettingsIncludeAllDelayMS, globalFlagIncludeAllDelayMS)
	Logger.Infof("%s: %s", SettingsTemplatePath, globalFlagTemplatePath)
	Logger.Infof("%s: %t", SettingsModeEnableCreate, globalModeEnableCreate)
	Logger.Infof("%s: %t", SettingsModeEnableUpdate, globalModeEnableUpdate)
	Logger.Infof("%s: %t", SettingsModeEnableDelete, globalModeEnableDelete)
	Logger.Infof("%s: %t", SettingsModeEnableBatch, globalModeEnableBatch)

	return nil
}

func bindEnv(name string, defaultValue interface{}) {
	viper.BindEnv(name)
	viper.SetDefault(name, defaultValue)
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
	}

	basicAuthMatcher := regexp.MustCompile("(Basic\\s+)[A-Za-z0-9=]+")
	message = basicAuthMatcher.ReplaceAllString(message, "$1 {base64 tenant/username:password}")

	return message
}
