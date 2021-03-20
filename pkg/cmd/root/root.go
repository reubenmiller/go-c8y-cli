package root

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/gorilla/websocket"
	"github.com/reubenmiller/go-c8y-cli/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ydefaults"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ysession"
	agentsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/agents"
	agentsListCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/agents/list"
	alarmsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alarms"
	alarmsSubscribeCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alarms/subscribe"
	aliasCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/alias"
	apiCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/api"
	applicationsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/applications"
	applicationsCreateHostedCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/applications/createhostedapplication"
	auditrecordsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/auditrecords"
	binariesCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/binaries"
	bulkoperationsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/bulkoperations"
	completionCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/completion"
	currentapplicationCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/currentapplication"
	databrokerCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/databroker"
	devicecredentialsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicecredentials"
	devicegroupsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicegroups"
	devicegroupsListCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicegroups/list"
	devicesCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices"
	devicesListCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/list"
	eventsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/events"
	eventsSubscribeCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/events/subscribe"
	identityCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/identity"
	inventoryCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory"
	inventoryFindCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/find"
	inventorySubscribeCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/subscribe"
	inventoryreferencesCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventoryreferences"
	measurementsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/measurements"
	measurementsSubscribeCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/measurements/subscribe"
	microservicesCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices"
	microservicesCreateCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/create"
	microservicesServiceUserCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/serviceuser"
	operationsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/operations"
	operationsSubscribeCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/operations/subscribe"
	realtimeCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/realtime"
	retentionrulesCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/retentionrules"
	sessionsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/sessions"
	settingsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/settings"
	systemoptionsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/systemoptions"
	templateCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/template"
	tenantoptionsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantoptions"
	tenantsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenants"
	tenantstatisticsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/tenantstatistics"
	usergroupsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/usergroups"
	userreferencesCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userreferences"
	userrolesCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/userroles"
	usersCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/users"
	versionCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/version"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/pkg/request"
	"github.com/reubenmiller/go-c8y-cli/pkg/utilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

type CmdRoot struct {
	*cobra.Command

	Verbose            bool
	Debug              bool
	ProgressBar        bool
	NoColor            bool
	SessionFile        string
	UseEnv             bool
	NoLog              bool
	ActivityLogMessage string

	Factory *cmdutil.Factory
	Config  func() (*config.Config, error)
	Logger  func() (*logger.Logger, error)
}

func NewCmdRoot(f *cmdutil.Factory, version, buildDate string) *CmdRoot {
	ccmd := &CmdRoot{
		Factory: f,
		Config:  f.Config,
		Logger:  f.Logger,
	}
	cmd := &cobra.Command{
		Use:   "c8y",
		Short: "Cumulocity command line interface",
		Long:  `A command line interface to interact with Cumulocity REST API. Ideal for quick prototyping, exploring the REST API and for Platform maintainers/power users`,

		SilenceErrors: true,
		SilenceUsage:  true,
		Example: heredoc.Doc(`
			$ c8y devices list
			$ c8y devices list --type "myDevice" | c8y devices update --data "myValue=1"
			$ c8y operations list --device myDeviceName
		`),
		Annotations: map[string]string{
			"help:feedback": heredoc.Doc(`
				Open an issue using 'c8y issue create -R github.com/cli/cli'
			`),
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := ccmd.Configure(); err != nil {
				return err
			}
			return ccmd.checkSessionExists(cmd, args)
		},
	}

	cmd.SetOut(f.IOStreams.Out)
	cmd.SetErr(f.IOStreams.ErrOut)

	isTerm := f.IOStreams.IsStdoutTTY()
	defaultOutputFormat := "json"
	defaultView := config.ViewsNone
	if isTerm {
		defaultOutputFormat = "table"
		defaultView = config.ViewsAll
	}

	// Global flags
	cmd.PersistentFlags().StringVar(&ccmd.SessionFile, "session", "", "Session configuration")
	cmd.PersistentFlags().BoolVarP(&ccmd.Verbose, "verbose", "v", false, "Verbose logging")
	cmd.PersistentFlags().IntP(flags.FlagPageSize, "p", c8ydefaults.PageSize, "Maximum results per page")
	cmd.PersistentFlags().Int64(flags.FlagCurrentPage, 0, "Current page size which should be returned")
	cmd.PersistentFlags().Int64("totalPages", 0, "Total number of pages to get")
	cmd.PersistentFlags().Bool("includeAll", false, "Include all results by iterating through each page")
	cmd.PersistentFlags().BoolP(flags.FlagWithTotalPages, "t", false, "Include all results")
	cmd.PersistentFlags().BoolP("compact", "c", !isTerm, "Compact instead of pretty-printed output. Pretty print is the default if output is the terminal")
	cmd.PersistentFlags().Bool("noAccept", false, "Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect")
	cmd.PersistentFlags().Bool("dry", false, "Dry run. Don't send any data to the server")
	cmd.PersistentFlags().String("dryFormat", "markdown", "Dry run output format. i.e. json, dump, markdown or curl")
	cmd.PersistentFlags().BoolVar(&ccmd.ProgressBar, "progress", false, "Show progress bar. This will also disable any other verbose output")
	cmd.PersistentFlags().BoolVarP(&ccmd.NoColor, "noColor", "M", !isTerm, "Don't use colors when displaying log entries on the console")
	cmd.PersistentFlags().BoolVar(&ccmd.UseEnv, "useEnv", false, "Allow loading Cumulocity session setting from environment variables")
	cmd.PersistentFlags().BoolP("raw", "r", false, "Raw values")
	cmd.PersistentFlags().String("proxy", "", "Proxy setting, i.e. http://10.0.0.1:8080")
	cmd.PersistentFlags().Bool("noProxy", false, "Ignore the proxy settings")
	cmd.PersistentFlags().Bool("withError", false, "Errors will be printed on stdout instead of stderr")

	// Activity log
	cmd.PersistentFlags().BoolVar(&ccmd.NoLog, "noLog", false, "Disables the activity log for the current command")
	cmd.PersistentFlags().StringVarP(&ccmd.ActivityLogMessage, "logMessage", "l", "", "Add custom message to the activity log")
	cmd.PersistentFlags().BoolVar(&ccmd.Debug, "debug", false, "Set very verbose log messages")

	// Concurrency
	cmd.PersistentFlags().Int("workers", 1, "Number of workers")
	cmd.PersistentFlags().Int64("maxJobs", 100, "Maximum number of jobs. 0 = unlimited (use with caution!)")

	// viper.BindPFlag("settings.defaults.maxJobs", cmd.PersistentFlags().Lookup("maxJobs"))
	cmd.PersistentFlags().Int("delay", 1000, "delay in milliseconds after each request")
	cmd.PersistentFlags().Int("abortOnErrors", 10, "Abort batch when reaching specified number of errors")

	// Error handling
	cmd.PersistentFlags().String("silentStatusCodes", "", "Status codes which will not print out an error message")

	cmd.PersistentFlags().Bool("flatten", false, "flatten")
	cmd.PersistentFlags().StringSlice("filter", nil, "filter")
	cmd.PersistentFlags().StringArray("select", nil, "select")
	cmd.PersistentFlags().String("view", defaultView, "View option")
	cmd.PersistentFlags().Float64("timeout", float64(10*60), "Timeout in seconds")

	// output
	cmd.PersistentFlags().StringP("output", "o", defaultOutputFormat, "Output format i.e. table, json, csv, csvheader")
	cmd.PersistentFlags().String("outputFile", "", "Save JSON output to file (after select)")
	cmd.PersistentFlags().String("outputFileRaw", "", "Save raw response to file")

	// confirmation
	cmd.PersistentFlags().BoolP("force", "f", false, "Do not prompt for confirmation")
	cmd.PersistentFlags().Bool("prompt", false, "Prompt for confirmation")
	cmd.PersistentFlags().String("confirmText", "", "Custom confirmation text")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("dryFormat", "json", "dump", "curl", "markdown"),
		completion.WithValidateSet("output", "json", "table", "csv", "csvheader"),
		completion.WithValidateSet("view", config.ViewsNone, config.ViewsAll),
		completion.WithSessionFile("session", func() string {
			cfg, err := ccmd.Config()
			if err != nil {
				return ""
			}
			return cfg.GetSessionHomeDir()
		}),
	)

	// Child commands
	commands := []*cobra.Command{
		auditrecordsCmd.NewSubCommand(f).GetCommand(),
		binariesCmd.NewSubCommand(f).GetCommand(),
		bulkoperationsCmd.NewSubCommand(f).GetCommand(),
		currentapplicationCmd.NewSubCommand(f).GetCommand(),
		databrokerCmd.NewSubCommand(f).GetCommand(),
		devicecredentialsCmd.NewSubCommand(f).GetCommand(),
		identityCmd.NewSubCommand(f).GetCommand(),
		inventoryreferencesCmd.NewSubCommand(f).GetCommand(),

		retentionrulesCmd.NewSubCommand(f).GetCommand(),
		sessionsCmd.NewSubCommand(f).GetCommand(),
		systemoptionsCmd.NewSubCommand(f).GetCommand(),
		tenantoptionsCmd.NewSubCommand(f).GetCommand(),
		tenantsCmd.NewSubCommand(f).GetCommand(),
		tenantstatisticsCmd.NewSubCommand(f).GetCommand(),
		usergroupsCmd.NewSubCommand(f).GetCommand(),
		userreferencesCmd.NewSubCommand(f).GetCommand(),
		userrolesCmd.NewSubCommand(f).GetCommand(),
		usersCmd.NewSubCommand(f).GetCommand(),
		versionCmd.NewCmdVersion(f).GetCommand(),
		completionCmd.NewCmdCompletion().GetCommand(),
		templateCmd.NewSubCommand(f).GetCommand(),
		settingsCmd.NewSubCommand(f).GetCommand(),
		realtimeCmd.NewSubCommand(f).GetCommand(),
	}

	cmd.AddCommand(commands...)

	// todo: merge custom commands
	//
	alarms := alarmsCmd.NewSubCommand(f).GetCommand()
	alarms.AddCommand(alarmsSubscribeCmd.NewCmdSubscribe(f).GetCommand())
	cmd.AddCommand(alarms)

	events := eventsCmd.NewSubCommand(f).GetCommand()
	events.AddCommand(eventsSubscribeCmd.NewCmdSubscribe(f).GetCommand())
	cmd.AddCommand(events)

	operations := operationsCmd.NewSubCommand(f).GetCommand()
	operations.AddCommand(operationsSubscribeCmd.NewCmdSubscribe(f).GetCommand())
	cmd.AddCommand(operations)

	measurements := measurementsCmd.NewSubCommand(f).GetCommand()
	measurements.AddCommand(measurementsSubscribeCmd.NewCmdSubscribe(f).GetCommand())
	cmd.AddCommand(measurements)

	// devices
	devices := devicesCmd.NewSubCommand(f).GetCommand()
	devices.AddCommand(devicesListCmd.NewCmdDevicesList(f).GetCommand())
	// devices.AddCommand(NewGetDeviceGroupCollectionCmd(f).GetCommand())
	cmd.AddCommand(devices)

	// devicegroups
	devicegroups := devicegroupsCmd.NewSubCommand(f).GetCommand()
	devicegroups.AddCommand(devicegroupsListCmd.NewCmdList(f).GetCommand())
	cmd.AddCommand(devicegroups)

	agents := agentsCmd.NewSubCommand(f).GetCommand()
	agents.AddCommand(agentsListCmd.NewCmdAgentList(f).GetCommand())
	cmd.AddCommand(agents)

	// microservices
	microservices := microservicesCmd.NewSubCommand(f).GetCommand()
	microservices.AddCommand(microservicesCreateCmd.NewCmdCreate(f).GetCommand())
	microservices.AddCommand(microservicesServiceUserCmd.NewSubCommand(f).GetCommand())
	cmd.AddCommand(microservices)

	// inventory
	inventory := inventoryCmd.NewSubCommand(f).GetCommand()
	inventory.AddCommand(inventoryFindCmd.NewCmdFind(f).GetCommand())
	inventory.AddCommand(inventorySubscribeCmd.NewCmdSubscribe(f).GetCommand())
	cmd.AddCommand(inventory)

	// applications
	applications := applicationsCmd.NewSubCommand(f).GetCommand()
	applications.AddCommand(applicationsCreateHostedCmd.NewCmdCreateHostedApplication(f).GetCommand())
	cmd.AddCommand(applications)

	// Manual commands
	cmd.AddCommand(aliasCmd.NewCmdAlias(f))
	cmd.AddCommand(apiCmd.NewSubCommand(f).GetCommand())

	// Handle errors (not in cobra libary)
	cmd.SilenceErrors = true

	ccmd.Command = cmd
	return ccmd
}

func (c *CmdRoot) Configure() error {
	cfg, err := c.Config()
	if err != nil {
		return err
	}
	log, err := c.Logger()
	if err != nil {
		return err
	}
	log.Debugf("Configuring core modules")
	consoleHandler, err := c.Factory.Console()
	if err != nil {
		return err
	}

	// config/env binding
	if err := cfg.BindPFlag(c.Command.PersistentFlags()); err != nil {
		log.Warningf("Some configuration binding failed. %s", err)
	}

	//
	// Update cmd factory before passing it along
	//

	// Update logger
	c.Factory.Logger = func() (*logger.Logger, error) {
		logOptions := logger.Options{
			Level: zapcore.WarnLevel,
			Color: !cfg.DisableColor(),
			Debug: cfg.Debug(),
		}
		if cfg.ShowProgress() {
			logOptions.Silent = true
		} else {
			if cfg.Verbose() {
				logOptions.Level = zapcore.InfoLevel
			}
			if cfg.Debug() {
				logOptions.Level = zapcore.DebugLevel
			}
		}

		customLogger := logger.NewLogger("c8y", logOptions)
		c8y.Logger = customLogger
		cfg.SetLogger(customLogger)
		return customLogger, nil
	}

	// Update activity logger
	c.Factory.ActivityLogger = func() (*activitylogger.ActivityLogger, error) {
		return c.configureActivityLog(cfg)
	}

	// Update data views
	c.Factory.DataView = func() (*dataview.DataView, error) {
		return dataview.NewDataView(".*", ".json", log, cfg.GetViewPaths()...)
	}

	consoleHandler.Format = cfg.GetOutputFormat()
	consoleHandler.Colorized = !cfg.DisableColor()
	consoleHandler.Compact = cfg.CompactJSON()
	consoleHandler.Disabled = cfg.ShowProgress() && c.Factory.IOStreams.IsStdoutTTY()

	// Update client
	c.Factory.Client = createCumulocityClient(c.Factory)
	return nil
}

func (c *CmdRoot) checkSessionExists(cmd *cobra.Command, args []string) error {
	log, err := c.Logger()
	if err != nil {
		return err
	}
	client, err := c.Factory.Client()
	if err != nil {
		return err
	}
	activityHandler, err := c.Factory.ActivityLogger()
	if err != nil {
		return err
	}
	cmdStr := cmd.Use
	if cmd.HasParent() && cmd.Parent().Use != "c8y" {
		cmdStr = cmd.Parent().Use + " " + cmdStr
	}
	log.Debugf("command str: %s", cmdStr)

	log.Infof("command: c8y %s", utilities.GetCommandLineArgs())

	activityHandler.LogCommand(cmd, args, cmdStr, c.ActivityLogMessage)

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
		return cmderrors.NewUserErrorWithExitCode(cmderrors.ExitNoSession, "A c8y session has not been loaded. Please create or activate a session and try again")
	}

	return nil
}

func (c *CmdRoot) configureActivityLog(cfg *config.Config) (*activitylogger.ActivityLogger, error) {
	disabled := !cfg.ActivityLogEnabled()
	if c.NoLog {
		disabled = true
	}
	options := activitylogger.Options{
		Disabled:     disabled,
		OutputFolder: cfg.GetActivityLogPath(),
		Methods:      strings.ToUpper(cfg.GetActivityLogMethodFilter()),
	}

	activitylog, err := activitylogger.NewActivityLogger(options)
	if err != nil {
		cfg.Logger.Errorf("Failed to load activity logger. %s", err)
		return nil, err
	}

	if disabled {
		cfg.Logger.Info("activityLog is disabled")
	} else {
		cfg.Logger.Infof("activityLog path: %s", activitylog.GetPath())
	}
	return activitylog, nil
}

func createCumulocityClient(f *cmdutil.Factory) func() (*c8y.Client, error) {
	return func() (*c8y.Client, error) {
		cfg, err := f.Config()
		if err != nil {
			return nil, err
		}
		log, err := f.Logger()
		if err != nil {
			return nil, err
		}
		consol, err := f.Console()
		if err != nil {
			return nil, err
		}

		log.Debug("Creating c8y client")
		configureProxySettings(cfg, log)
		httpClient := newHTTPClient(cfg.IgnoreProxy())

		// Only bind when not setting the session
		if cfg.UseEnvironment() {
			log.Info("Binding authorization environment variables")
			if err := cfg.BindAuthorization(); err != nil {
				log.Warnf("Failed to bind to authorization variables. %s", err)
			}
		}

		client := c8y.NewClient(
			httpClient,
			c8ysession.FormatHost(cfg.GetHost()),
			cfg.GetTenant(),
			cfg.GetUsername(),
			cfg.MustGetPassword(),
			true,
		)

		client.SetRequestOptions(c8y.DefaultRequestOptions{
			DryRun: cfg.DryRun(),
			DryRunHandler: func(options *c8y.RequestOptions, req *http.Request) {
				handler := &request.RequestHandler{
					IsTerminal:    f.IOStreams.IsStdoutTTY(),
					Client:        client,
					Config:        cfg,
					Logger:        log,
					Console:       consol,
					HideSensitive: config.HideSensitiveInformationIfActive,
				}
				handler.DryRunHandler(f.IOStreams, options, req)
			},
		})

		// load authentication
		if err := loadAuthentication(cfg, client); err != nil {
			log.Warnf("Could not load authentication. %s", err)
		}

		timeout := cfg.RequestTimeout()
		log.Debugf("timeout: %0.3f", timeout)

		// Should we use the tenant in the name or not
		if viper.IsSet("useTenantPrefix") {
			client.UseTenantInUsername = viper.GetBool("useTenantPrefix")
		}

		// Add the realtime client
		client.Realtime = c8y.NewRealtimeClient(
			client.BaseURL.String(),
			newWebsocketDialer(cfg.IgnoreProxy()),
			client.TenantName,
			client.Username,
			client.Password,
		)

		// Set realtime authorization
		if client.AuthorizationMethod == c8y.AuthMethodOAuth2Internal {
			client.Realtime.SetXSRFToken(client.GetXSRFToken())

			if len(client.Cookies) > 0 {
				if err := client.Realtime.SetCookies(client.Cookies); err != nil {
					log.Errorf("Failed to set websocket cookie jar. %s", err)
				}
			}
		}
		return client, nil
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

func configureProxySettings(cfg *config.Config, log *logger.Logger) {

	// only parse env variables if no explict config file is given
	// if globalFlagUseEnv {
	// 	Logger.Println("C8Y_USE_ENVIRONMENT is set. Environment variables can be used to override config settings")
	// 	viper.AutomaticEnv()
	// }

	// Proxy settings
	// Either use explicit proxy, ignore proxy, or use existing env variables
	// --proxy "http://10.0.0.1:8080"
	// --noProxy
	// HTTP_PROXY=http://10.0.0.1:8080
	// NO_PROXY=localhost,127.0.0.1
	proxy := cfg.Proxy()
	noProxy := cfg.IgnoreProxy()
	if noProxy {
		log.Debug("using explicit noProxy setting")
		os.Setenv("HTTP_PROXY", "")
		os.Setenv("HTTPS_PROXY", "")
		os.Setenv("http_proxy", "")
		os.Setenv("https_proxy", "")
	} else {
		if proxy != "" {
			log.Debugf("using explicit proxy [%s]", proxy)

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
				log.Debugf("Using existing env variables.%s", proxySettings)
			}
		}
	}
}
