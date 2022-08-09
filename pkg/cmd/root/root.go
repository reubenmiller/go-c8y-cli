package root

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydefaults"
	activityLogCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/activitylog"
	agentsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/agents"
	agentsListCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/agents/list"
	alarmsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alarms"
	alarmsAssertCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alarms/assert"
	alarmsSubscribeCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alarms/subscribe"
	aliasCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/alias"
	apiCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/api"
	applicationsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/applications"
	applicationsCreateHostedCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/applications/createhostedapplication"
	applicationsOpenCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/applications/open"
	assertCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/assert"
	auditrecordsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/auditrecords"
	binariesCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/binaries"
	bulkoperationsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/bulkoperations"
	cacheCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/cache"
	completionCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/completion"
	configurationCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/configuration"
	configurationListCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/configuration/list"
	currentapplicationCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/currentapplication"
	currenttenantCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/currenttenant"
	currentuserCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/currentuser"
	databrokerCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/databroker"
	devicegroupsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups"
	devicegroupsListCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicegroups/list"
	deviceManagementCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devicemanagement"
	deviceprofilesCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceprofiles"
	deviceregistrationCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/deviceregistration"
	devicesCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices"
	devicesAssertCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/assert"
	devicesAvailabilityCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/availability"
	devicesListCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/list"
	deviceStatisticsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/statistics"
	deviceUserCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/devices/user"
	eventsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events"
	eventsAssertCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/assert"
	eventsSubscribeCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/events/subscribe"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/factory"
	firmwareCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/firmware"
	firmwareListCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/firmware/list"
	firmwareVersionsPatchesCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/firmware/patches"
	firmwarePatchesCreateCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/firmware/patches/create"
	firmwareVersionsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/firmware/versions"
	firmwareVersionsCreateCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/firmware/versions/create"
	identityCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/identity"
	inventoryCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory"
	inventoryAdditionsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/additions"
	inventoryAssertCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/assert"
	inventoryAssetsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/assets"
	inventoryFindCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/find"
	inventorySubscribeCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/subscribe"
	inventoryWaitCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/inventory/wait"
	measurementsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/measurements"
	measurementsAssertCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/measurements/assert"
	measurementsCreateBulkCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/measurements/createBulk"
	measurementsSubscribeCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/measurements/subscribe"
	microservicesCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/microservices"
	microservicesCreateCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/microservices/create"
	microservicesLogLevelsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/microservices/loglevels"
	microservicesServiceUserCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/microservices/serviceuser"
	operationsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/operations"
	operationsAssertCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/operations/assert"
	operationsSubscribeCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/operations/subscribe"
	operationsWaitCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/operations/wait"
	realtimeCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/realtime"
	retentionrulesCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/retentionrules"
	sessionsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/sessions"
	settingsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/settings"
	smartgroupsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/smartgroups"
	smartgroupsListCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/smartgroups/list"
	softwareCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/software"
	softwareListCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/software/list"
	softwareVersionsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/software/versions"
	softwareVersionsCreateCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/software/versions/create"
	systemoptionsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/systemoptions"
	templateCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/template"
	tenantoptionsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenantoptions"
	tenantsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenants"
	tenantstatisticsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/tenantstatistics"
	usergroupsCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/usergroups"
	userreferencesCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/userreferences"
	userrolesCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/userroles"
	usersCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/users"
	utilCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/util"
	versionCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/version"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/utilities"
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			disableEncryptionCheck := !cmdutil.IsConfigEncryptionCheckEnabled(cmd)
			if err := ccmd.Configure(disableEncryptionCheck); err != nil {
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
	cmd.PersistentFlags().Int64(flags.FlagCurrentPage, 0, "Current page which should be returned")
	cmd.PersistentFlags().Int64("totalPages", 0, "Total number of pages to get")
	cmd.PersistentFlags().Bool("includeAll", false, "Include all results by iterating through each page")
	cmd.PersistentFlags().BoolP(flags.FlagWithTotalPages, "t", false, "Request Cumulocity to include the total pages in the response statistics under .statistics.totalPages")
	cmd.PersistentFlags().BoolP("compact", "c", !isTerm, "Compact instead of pretty-printed output when using json output. Pretty print is the default if output is the terminal")
	cmd.PersistentFlags().Bool("noAccept", false, "Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect")
	cmd.PersistentFlags().Bool("dry", false, "Dry run. Don't send any data to the server")
	cmd.PersistentFlags().String("dryFormat", "markdown", "Dry run output format. i.e. json, dump, markdown or curl")
	cmd.PersistentFlags().BoolVar(&ccmd.ProgressBar, "progress", false, "Show progress bar. This will also disable any other verbose output")
	cmd.PersistentFlags().BoolVarP(&ccmd.NoColor, "noColor", "M", !isTerm, "Don't use colors when displaying log entries on the console")
	cmd.PersistentFlags().BoolP("raw", "r", false, "Show raw response. This mode will force output=json and view=off")
	cmd.PersistentFlags().String("proxy", "", "Proxy setting, i.e. http://10.0.0.1:8080")
	cmd.PersistentFlags().Bool("noProxy", false, "Ignore the proxy settings")
	cmd.PersistentFlags().Bool("withError", false, "Errors will be printed on stdout instead of stderr")
	cmd.PersistentFlags().StringSliceP("header", "H", nil, `custom headers. i.e. --header "Accept: value, AnotherHeader: myvalue"`)
	cmd.PersistentFlags().StringSlice("customQueryParam", nil, `add custom URL query parameters. i.e. --customQueryParam 'withCustomOption=true,myOtherOption=myvalue'`)

	// help
	cmd.PersistentFlags().Bool("examples", false, "Show examples for the current command")

	// Activity log
	cmd.PersistentFlags().BoolVar(&ccmd.NoLog, "noLog", false, "Disables the activity log for the current command")
	cmd.PersistentFlags().StringVarP(&ccmd.ActivityLogMessage, "logMessage", "l", "", "Add custom message to the activity log")
	cmd.PersistentFlags().BoolVar(&ccmd.Debug, "debug", false, "Set very verbose log messages")

	// Concurrency
	cmd.PersistentFlags().Int("workers", 1, "Number of workers")
	cmd.PersistentFlags().Int64("maxJobs", 0, "Maximum number of jobs. 0 = unlimited (use with caution!)")

	cmd.PersistentFlags().String("delay", "0ms", "delay after each request, i.e. 5ms, 1.2s")
	cmd.PersistentFlags().String("delayBefore", "0ms", "delay before each request, i.e. 5ms, 1.2s")
	cmd.PersistentFlags().Int("abortOnErrors", 10, "Abort batch when reaching specified number of errors")

	// Error handling
	cmd.PersistentFlags().String("silentStatusCodes", "", "Status codes which will not print out an error message")
	cmd.PersistentFlags().Bool("silentExit", false, "Silent status codes do not affect the exit code")

	cmd.PersistentFlags().Bool("flatten", false, "flatten json output by replacing nested json properties with properties where their names are represented by dot notation")
	cmd.PersistentFlags().StringArray("filter", nil, "Apply a client side filter to response before returning it to the user")
	cmd.PersistentFlags().StringArray("select", nil, "Comma separated list of properties to return. wildcards and globstar accepted, i.e. --select 'id,name,type,**.serialNumber'")
	cmd.PersistentFlags().String("view", defaultView, "Use views when displaying data on the terminal. Disable using --view off")
	cmd.PersistentFlags().String("timeout", "60s", "Request timeout duration, i.e. 60s, 2m")

	// output
	cmd.PersistentFlags().StringP("output", "o", defaultOutputFormat, "Output format i.e. table, json, csv, csvheader")
	cmd.PersistentFlags().String("outputFile", "", "Save JSON output to file (after select/view)")
	cmd.PersistentFlags().String("outputFileRaw", "", "Save raw response to file (before select/view)")

	// input parsing
	cmd.PersistentFlags().BoolP(flags.FlagNullInput, "n", false, "Don't read the input (stdin). Useful if using in shell for/while loops")
	cmd.PersistentFlags().Bool(flags.FlagAllowEmptyPipe, false, "Don't fail when piped input is empty (stdin)")

	// confirmation
	cmd.PersistentFlags().BoolP("force", "f", false, "Do not prompt for confirmation. Ignored when using --confirm")
	cmd.PersistentFlags().Bool("confirm", false, "Prompt for confirmation")
	cmd.PersistentFlags().String("confirmText", "", "Custom confirmation text")

	// caching
	cmd.PersistentFlags().Bool("cache", false, "Enable cached responses")
	cmd.PersistentFlags().Bool("noCache", false, "Force disabling of cached responses (overwrites cache setting)")
	cmd.PersistentFlags().String("cacheTTL", "60s", "Cache time-to-live (TTL) as a duration, i.e. 60s, 2m")
	cmd.PersistentFlags().StringSlice("cacheBodyPaths", []string{}, "Cache should limit hashing of selected paths in the json body. Empty indicates all values")

	// ssl settings
	cmd.PersistentFlags().BoolP("insecure", "k", false, "Allow insecure server connections when using SSL")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("dryFormat", "json", "dump", "curl", "markdown"),
		completion.WithValidateSet(
			"output",
			config.OutputJSON.String()+"\tjson",
			config.OutputTable.String()+"\ttable format",
			config.OutputCSV.String()+"\tcsv format without headers",
			config.OutputCSVWithHeader.String()+"\tcsv format with headers",
			config.OutputServerResponse.String()+"\tUnparsed server response",
		),
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
		assertCmd.NewSubCommand(f).GetCommand(),
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
		utilCmd.NewSubCommand(f).GetCommand(),
		cacheCmd.NewSubCommand(f).GetCommand(),
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
	alarms.AddCommand(alarmsAssertCmd.NewSubCommand(f).GetCommand())
	cmd.AddCommand(alarms)

	events := eventsCmd.NewSubCommand(f).GetCommand()
	events.AddCommand(eventsSubscribeCmd.NewCmdSubscribe(f).GetCommand())
	events.AddCommand(eventsAssertCmd.NewSubCommand(f).GetCommand())
	cmd.AddCommand(events)

	operations := operationsCmd.NewSubCommand(f).GetCommand()
	operations.AddCommand(operationsSubscribeCmd.NewCmdSubscribe(f).GetCommand())
	operations.AddCommand(operationsWaitCmd.NewCmdWait(f).GetCommand())
	operations.AddCommand(operationsAssertCmd.NewSubCommand(f).GetCommand())
	cmd.AddCommand(operations)

	measurements := measurementsCmd.NewSubCommand(f).GetCommand()
	measurements.AddCommand(measurementsSubscribeCmd.NewCmdSubscribe(f).GetCommand())
	measurements.AddCommand(measurementsAssertCmd.NewSubCommand(f).GetCommand())
	measurements.AddCommand(measurementsCreateBulkCmd.NewCreateBulkCmd(f).GetCommand())
	cmd.AddCommand(measurements)

	// device management
	devicemanagement := deviceManagementCmd.NewSubCommand(f).GetCommand()
	cmd.AddCommand(devicemanagement)

	// devices
	devices := devicesCmd.NewSubCommand(f).GetCommand()
	devices.AddCommand(devicesListCmd.NewCmdDevicesList(f).GetCommand())
	devices.AddCommand(devicesAssertCmd.NewSubCommand(f).GetCommand())
	devices.AddCommand(devicesAvailabilityCmd.NewSubCommand(f).GetCommand())
	devices.AddCommand(deviceStatisticsCmd.NewSubCommand(f).GetCommand())
	devices.AddCommand(deviceUserCmd.NewSubCommand(f).GetCommand())
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
	microservices.AddCommand(microservicesLogLevelsCmd.NewSubCommand(f).GetCommand())
	cmd.AddCommand(microservices)

	// inventory
	inventory := inventoryCmd.NewSubCommand(f).GetCommand()
	inventory.AddCommand(inventoryFindCmd.NewCmdFind(f).GetCommand())
	inventory.AddCommand(inventorySubscribeCmd.NewCmdSubscribe(f).GetCommand())
	inventory.AddCommand(inventoryAdditionsCmd.NewSubCommand(f).GetCommand())
	inventory.AddCommand(inventoryAssetsCmd.NewSubCommand(f).GetCommand())
	inventory.AddCommand(inventoryWaitCmd.NewCmdWait(f).GetCommand())
	inventory.AddCommand(inventoryAssertCmd.NewSubCommand(f).GetCommand())
	cmd.AddCommand(inventory)

	// configuration
	configuration := configurationCmd.NewSubCommand(f).GetCommand()
	configuration.AddCommand(configurationListCmd.NewListCmd(f).GetCommand())
	cmd.AddCommand(configuration)

	// software
	softwareVersions := softwareVersionsCmd.NewSubCommand(f).GetCommand()
	softwareVersions.AddCommand(softwareVersionsCreateCmd.NewCreateCmd(f).GetCommand())
	software := softwareCmd.NewSubCommand(f).GetCommand()
	software.AddCommand(softwareListCmd.NewListCmd(f).GetCommand())
	software.AddCommand(softwareVersions)
	cmd.AddCommand(software)

	// firmware
	firmwarePatches := firmwareVersionsPatchesCmd.NewSubCommand(f).GetCommand()
	firmwarePatches.AddCommand(firmwarePatchesCreateCmd.NewCreatePatchCmd(f).GetCommand())
	firmwareVersions := firmwareVersionsCmd.NewSubCommand(f).GetCommand()
	firmwareVersions.AddCommand(firmwareVersionsCreateCmd.NewCreateCmd(f).GetCommand())

	firmware := firmwareCmd.NewSubCommand(f).GetCommand()
	firmware.AddCommand(firmwareListCmd.NewListCmd(f).GetCommand())
	firmware.AddCommand(firmwareVersions)
	firmware.AddCommand(firmwarePatches)
	cmd.AddCommand(firmware)

	// deviceprofilesCmd
	deviceprofiles := deviceprofilesCmd.NewSubCommand(f).GetCommand()
	cmd.AddCommand(deviceprofiles)

	// applications
	applications := applicationsCmd.NewSubCommand(f).GetCommand()
	applications.AddCommand(applicationsCreateHostedCmd.NewCmdCreateHostedApplication(f).GetCommand())
	applications.AddCommand(applicationsOpenCmd.NewOpenCmd(f).GetCommand())
	cmd.AddCommand(applications)

	// smart groups
	smartgroups := smartgroupsCmd.NewSubCommand(f).GetCommand()
	smartgroups.AddCommand(smartgroupsListCmd.NewListCmd(f).GetCommand())
	cmd.AddCommand(smartgroups)

	// Manual commands
	cmd.AddCommand(aliasCmd.NewCmdAlias(f))
	cmd.AddCommand(apiCmd.NewSubCommand(f).GetCommand())

	// Handle errors (not in cobra libary)
	cmd.SilenceErrors = true

	ccmd.Command = cmd
	return ccmd
}

func (c *CmdRoot) Configure(disableEncryptionCheck bool) error {
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
			// Don't silence log levels completely in case of errors
			// mode errors
			logOptions.Silent = false
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
		client, err := factory.CreateCumulocityClient(c.Factory, c.SessionFile, c.SessionUsername, c.SessionPassword, disableEncryptionCheck)()
		if c.SessionUsername != "" || c.SessionPassword != "" {
			client.AuthorizationMethod = c8y.AuthMethodBasic
			c.log.Debug("Forcing basic authentication as user provided username/password")
		}

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
		log.Infof("Loaded session: %s", cfg.HideSensitiveInformationIfActive(client, sessionFile))
		if _, err := os.Stat(sessionFile); err != nil {
			log.Warnf("Failed to verify session file. %s", err)
		}
	}

	if cfg.DisableStdin() {
		// Note: Stdin is disabled elsewhere
		log.Info("Disabling reading from stdin (does not accept piped data)")
	}

	log.Debugf("command str: %s", cmdStr)
	log.Infof("command: c8y %s", utilities.GetCommandLineArgs())
	log.Debugf("output format: %s", cfg.GetOutputFormat().String())

	// print examples
	if cmd.Flags().Changed("examples") {
		examples := fmt.Sprintf("%s\n", cmd.Example)
		// style := markdown.GetStyle(c.Factory.IOStreams.TerminalTheme())
		// log.Debugf("GLAMOR style: %s", style)
		// mdContent, _ := markdown.Render(examples, style)
		fmt.Fprint(c.Factory.IOStreams.Out, examples)
		return cmderrors.ErrHelp
	}

	// TODO: Find more efficient/extensible way of ignoring specific commands in the activity log
	if cmd.Name() != cobra.ShellCompRequestCmd && cmd.CalledAs() != cobra.ShellCompNoDescRequestCmd && !strings.HasPrefix(cmdStr, "activitylog") && !strings.HasPrefix(cmdStr, "completion") && !strings.HasPrefix(cmdStr, "version") {
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
