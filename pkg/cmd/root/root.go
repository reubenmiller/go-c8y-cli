package root

import (
	"os"
	"strings"
	"sync"

	"github.com/MakeNowJust/heredoc"
	"github.com/reubenmiller/go-c8y-cli/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ydefaults"
	activityLogCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/activitylog"
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
	currenttenantCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/currenttenant"
	currentuserCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/currentuser"
	databrokerCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/databroker"
	devicegroupsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicegroups"
	devicegroupsListCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devicegroups/list"
	deviceregistrationCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/deviceregistration"
	devicesCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices"
	devicesListCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/devices/list"
	eventsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/events"
	eventsSubscribeCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/events/subscribe"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/factory"
	identityCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/identity"
	inventoryCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory"
	inventoryAdditionsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/additions"
	inventoryAssetsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/assets"
	inventoryFindCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/find"
	inventorySubscribeCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/inventory/subscribe"
	measurementsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/measurements"
	measurementsSubscribeCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/measurements/subscribe"
	microservicesCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices"
	microservicesCreateCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/create"
	microservicesServiceUserCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/microservices/serviceuser"
	operationsCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/operations"
	operationsSubscribeCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/operations/subscribe"
	operationsWaitCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/operations/wait"
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
	"github.com/reubenmiller/go-c8y-cli/pkg/utilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

type CmdRoot struct {
	*cobra.Command

	Verbose            bool
	Debug              bool
	ProgressBar        bool
	NoColor            bool
	SessionFile        string
	SessionUsername    string
	SessionPassword    string
	NoLog              bool
	ActivityLogMessage string

	Factory *cmdutil.Factory

	client      *c8y.Client
	log         *logger.Logger
	activitylog *activitylogger.ActivityLogger
	dataview    *dataview.DataView
	mu          sync.RWMutex
	muLog       sync.RWMutex
	muDataView  sync.RWMutex
}

func NewCmdRoot(f *cmdutil.Factory, version, buildDate string) *CmdRoot {
	ccmd := &CmdRoot{
		Factory: f,
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
	defaultView := config.ViewsOff
	if isTerm {
		defaultOutputFormat = "table"
		defaultView = config.ViewsAuto
	}

	// Global flags
	cmd.PersistentFlags().StringVar(&ccmd.SessionFile, "session", "", "Session configuration")
	cmd.PersistentFlags().StringVarP(&ccmd.SessionUsername, "sessionUsername", "U", "", "Override session username. i.e. peter or t1234/peter (with tenant)")
	cmd.PersistentFlags().StringVarP(&ccmd.SessionPassword, "sessionPassword", "P", "", "Override session password")
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
	cmd.PersistentFlags().BoolP("raw", "r", false, "Raw values")
	cmd.PersistentFlags().String("proxy", "", "Proxy setting, i.e. http://10.0.0.1:8080")
	cmd.PersistentFlags().Bool("noProxy", false, "Ignore the proxy settings")
	cmd.PersistentFlags().Bool("withError", false, "Errors will be printed on stdout instead of stderr")
	cmd.PersistentFlags().StringSliceP("header", "H", nil, `custom headers. i.e. --header "Accept: value, AnotherHeader: myvalue"`)
	cmd.PersistentFlags().StringSlice("queryParam", nil, `custom query parameters. i.e. --queryParam "withCustomOption=true myOtherOption=myvalue"`)

	// Activity log
	cmd.PersistentFlags().BoolVar(&ccmd.NoLog, "noLog", false, "Disables the activity log for the current command")
	cmd.PersistentFlags().StringVarP(&ccmd.ActivityLogMessage, "logMessage", "l", "", "Add custom message to the activity log")
	cmd.PersistentFlags().BoolVar(&ccmd.Debug, "debug", false, "Set very verbose log messages")

	// Concurrency
	cmd.PersistentFlags().Int("workers", 1, "Number of workers")
	cmd.PersistentFlags().Int64("maxJobs", 100, "Maximum number of jobs. 0 = unlimited (use with caution!)")

	cmd.PersistentFlags().Int("delay", 1000, "delay in milliseconds after each request")
	cmd.PersistentFlags().Int("delayBefore", 0, "delay in milliseconds before each request")
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
	cmd.PersistentFlags().Bool("confirm", false, "Prompt for confirmation")
	cmd.PersistentFlags().String("confirmText", "", "Custom confirmation text")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("dryFormat", "json", "dump", "curl", "markdown"),
		completion.WithValidateSet("output", "json", "table", "csv", "csvheader"),
		completion.WithSessionFile("session", config.ConfigExtensions, func() string {
			cfg, err := ccmd.Factory.Config()
			if err != nil {
				return ""
			}
			return cfg.GetSessionHomeDir()
		}),
		cmdutil.WithViewCompletion("view", func() (*dataview.DataView, error) { return ccmd.Factory.DataView() }),
	)

	// Child commands
	commands := []*cobra.Command{
		auditrecordsCmd.NewSubCommand(f).GetCommand(),
		binariesCmd.NewSubCommand(f).GetCommand(),
		bulkoperationsCmd.NewSubCommand(f).GetCommand(),
		currentapplicationCmd.NewSubCommand(f).GetCommand(),
		databrokerCmd.NewSubCommand(f).GetCommand(),
		deviceregistrationCmd.NewSubCommand(f).GetCommand(),
		identityCmd.NewSubCommand(f).GetCommand(),
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
		currenttenantCmd.NewSubCommand(f).GetCommand(),
		currentuserCmd.NewSubCommand(f).GetCommand(),
		activityLogCmd.NewSubCommand(f).GetCommand(),
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
	operations.AddCommand(operationsWaitCmd.NewCmdWait(f).GetCommand())
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
	inventory.AddCommand(inventoryAdditionsCmd.NewSubCommand(f).GetCommand())
	inventory.AddCommand(inventoryAssetsCmd.NewSubCommand(f).GetCommand())
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
	cfg, err := c.Factory.Config()
	if err != nil {
		return err
	}
	log, err := c.Factory.Logger()
	if err != nil {
		return err
	}
	log.Debugf("Configuring core modules")
	consoleHandler, err := c.Factory.Console()
	if err != nil {
		return err
	}

	// config/env binding
	previousSession := cfg.GetSessionFile()
	if err := cfg.BindPFlag(c.Command.PersistentFlags()); err != nil {
		log.Warningf("Some configuration binding failed. %s", err)
	}

	if c.SessionFile != "" {
		cfg.SetSessionFile(c.SessionFile)
	}

	// re-load config if they are using the session argument
	currentSession := cfg.GetSessionFile()
	if previousSession != currentSession {
		log.Infof("Session file has changed from %s to %s. Reading new session", previousSession, currentSession)
		if _, err := cfg.ReadConfigFiles(nil); err != nil {
			log.Infof("Failed to read configuration. Trying to proceed anyway. %s", err)
		}
	}

	//
	// Update cmd factory before passing it along
	//

	// Update logger
	c.Factory.Logger = func() (*logger.Logger, error) {
		c.muLog.Lock()
		defer c.muLog.Unlock()
		if c.log != nil {
			return c.log, nil
		}
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
		c.log = customLogger
		return customLogger, nil
	}

	// Update activity logger
	c.Factory.ActivityLogger = func() (*activitylogger.ActivityLogger, error) {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.activitylog != nil {
			return c.activitylog, nil
		}
		al, err := c.configureActivityLog(cfg)
		c.activitylog = al
		return al, err
	}

	// Update data views
	c.Factory.DataView = func() (*dataview.DataView, error) {
		c.muDataView.Lock()
		defer c.muDataView.Unlock()
		if c.dataview != nil {
			return c.dataview, nil
		}

		l, _ := c.Factory.Logger()
		dv, err := dataview.NewDataView(".*", ".json", l, cfg.GetViewPaths()...)
		c.dataview = dv
		return dv, err
	}

	consoleHandler.Format = cfg.GetOutputFormat()
	consoleHandler.Colorized = !cfg.DisableColor()
	consoleHandler.Compact = cfg.CompactJSON()
	consoleHandler.Disabled = cfg.ShowProgress() && c.Factory.IOStreams.IsStdoutTTY()

	// Update client

	c.Factory.Client = func() (*c8y.Client, error) {
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.client != nil {
			return c.client, nil
		}
		client, err := factory.CreateCumulocityClient(c.Factory, c.SessionFile, c.SessionUsername, c.SessionPassword)()
		if c.log != nil {
			c8y.Logger = c.log
		} else {
			c8y.Logger = logger.NewDummyLogger("c8y")
		}
		c.client = client
		return client, err
	}
	return nil
}

func (c *CmdRoot) checkSessionExists(cmd *cobra.Command, args []string) error {
	log, err := c.Factory.Logger()
	if err != nil {
		return err
	}
	cfg, err := c.Factory.Config()
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

	// print log information
	sessionFile := cfg.GetSessionFile()
	if sessionFile != "" {
		log.Infof("Loaded session: %s", config.HideSensitiveInformationIfActive(client, sessionFile))
		if _, err := os.Stat(sessionFile); err != nil {
			log.Warnf("Failed to verify session file. %s", err)
		}
	}

	log.Debugf("command str: %s", cmdStr)
	log.Infof("command: c8y %s", utilities.GetCommandLineArgs())
	log.Debugf("output format: %s", cfg.GetOutputFormat().String())

	if cmd.Name() != cobra.ShellCompRequestCmd && cmd.CalledAs() != cobra.ShellCompNoDescRequestCmd && !strings.HasPrefix(cmdStr, "activitylog") {
		activityHandler.LogCommand(cmd, args, cmdStr, c.ActivityLogMessage)
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
