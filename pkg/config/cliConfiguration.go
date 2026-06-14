package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydefaults"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonfilter"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logintype"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/numbers"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/pathresolver"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/totp"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/vbauerster/mpb/v6"
)

var (
	// EnvSessionHide hides sensitive session information
	EnvSessionHide = "C8Y_SETTINGS_SESSION_HIDE"

	// EnvPassphrase passphrase environment variable name
	EnvPassphrase = "C8Y_PASSPHRASE"

	// EnvPassphraseText passphrase text environment variable name
	EnvPassphraseText = "C8Y_PASSPHRASE_TEXT"

	// EnvSessionMode session mode (short alias)
	EnvSessionMode = "C8Y_MODE"

	// PrefixEncrypted prefix used in encrypted string fields to identify when a string is encrypted or not
	PrefixEncrypted = "{encrypted}"

	// KeyFileName is the name of the reference encryption text
	KeyFileName = ".key"

	// ActivityLogDirName name of the activitylog directory
	ActivityLogDirName = "activitylog"
)

const (
	EnvSettingsPrefix = "c8y"

	// SettingsGlobalName name of the settings file (without extension)
	SettingsGlobalName = "settings"
)

var (
	ProviderTypeAuto        = "auto"
	ProviderTypeFile        = "file"
	ProviderTypeEnv         = "env"
	ProviderTypeExternal    = "external"
	ProviderTypeStdin       = "stdin"
	ProviderTypeInteractive = "interactive"
)

const (
	// SettingsIncludeAllPageSize property name used to control the default page size when using includeAll parameter
	SettingsIncludeAllPageSize = "settings.includeAll.pageSize"

	// SettingEncryptionCachePassphrase setting to cache the passphrase via environment variables
	SettingEncryptionCachePassphrase = "settings.encryption.cachePassphrase"

	// SettingsMaxWorkers property name used to control the hard limit on the maximum workers used in batch operations
	SettingsMaxWorkers = "settings.defaults.maxWorkers"

	// SettingsWorkers number of workers to use
	SettingsWorkers = "settings.defaults.workers"

	// SettingsMaxJobs maximum allowed jobs to be executed
	SettingsMaxJobs = "settings.defaults.maxJobs"

	// SettingsCurrentPage current page
	SettingsCurrentPage = "settings.defaults.currentPage"

	// SettingsTotalPages total pages to return
	SettingsTotalPages = "settings.defaults.totalPages"

	// SettingsIncludeAll include all available results
	SettingsIncludeAll = "settings.defaults.includeAll"

	// SettingsPaginationStrategy controls how results are iterated (auto, offset, id, time)
	SettingsPaginationStrategy = "settings.defaults.paginationStrategy"

	// SettingsIncludeAllDelayMS delay in milliseconds between retrieving the next page when using include all
	SettingsIncludeAllDelayMS = "settings.includeAll.delayMS"

	// SettingsPageSize page size
	SettingsPageSize = "settings.defaults.pageSize"

	// SettingsWithTotalPages include the total pages statistics under statistics.totalPages
	SettingsWithTotalPages = "settings.defaults.withTotalPages"

	// SettingsWithTotalElements include the total pages statistics under statistics.totalPages
	SettingsWithTotalElements = "settings.defaults.withTotalElements"

	// SettingsRawOutput include the raw (original) output instead of only returning the nested array property
	SettingsRawOutput = "settings.defaults.raw"

	// SettingsIgnoreAcceptHeader ignore the accept header / set the Accept header to an empty string
	SettingsIgnoreAcceptHeader = "settings.defaults.noAccept"

	// SettingsHeader custom headers to be added to outgoing requests
	SettingsHeader = "settings.defaults.header"

	// SettingsQueryParameters custom query parameters to be added to outgoing requests
	SettingsQueryParameters = "settings.defaults.customQueryParam"

	// SettingsDryRun dry run. Don't send any requests, just print out the information
	SettingsDryRun = "settings.defaults.dry"

	// SettingsDryRunPattern list of methods which should be conditionally dry, i.e. "PUT POST DELETE"
	SettingsDryRunPattern = "settings.defaults.dryPattern"

	// SettingsUseCompression use compression for HTTP client
	SettingsUseCompression = "settings.http.compression"

	// SettingsHTTPMaxRetries maximum number of retries by the HTTP client
	SettingsHTTPMaxRetries = "settings.defaults.retries"

	// SettingsHTTPRetryWaitMin minimum duration to wait before retrying a failed HTTP request
	SettingsHTTPRetryWaitMin = "settings.http.retryWaitMin"

	// SettingsHTTPRetryWaitMax maximum duration to wait before retrying a failed HTTP request
	SettingsHTTPRetryWaitMax = "settings.http.retryWaitMax"

	// SettingsDryRunFormat dry run output format. Controls how the dry run information is displayed
	SettingsDryRunFormat = "settings.defaults.dryFormat"

	// SettingsDebug Show debug messages
	SettingsDebug = "settings.defaults.debug"

	// SettingsVerbose Show verbose log messages
	SettingsVerbose = "settings.defaults.verbose"

	// SettingsJSONCompact show compact json output
	SettingsJSONCompact = "settings.defaults.compact"

	// SettingsShowProgress show progress bar
	SettingsShowProgress = "settings.defaults.progress"

	SettingsForceTTY = "settings.defaults.forceTTY"

	// SettingsDisableColor don't print progress bar
	SettingsDisableProgress = "settings.defaults.noProgress"

	// SettingsDisableColor don't print console output in color
	SettingsDisableColor = "settings.defaults.noColor"

	// SettingsProxy http/s proxy settings
	SettingsProxy = "settings.defaults.proxy"

	// SettingsIgnoreProxy ignore proxy settings
	SettingsIgnoreProxy = "settings.defaults.noProxy"

	// SettingsWithError return the error response on stdout rather than stderr
	SettingsWithError = "settings.defaults.withError"

	// SettingsWorkerDelay delay in milliseconds to wait after each request before the worker processes a new job (request)
	SettingsWorkerDelay = "settings.defaults.delay"

	// SettingsWorkerDelayBefore delay in milliseconds to wait before each request
	SettingsWorkerDelayBefore = "settings.defaults.delayBefore"

	// SettingsAbortOnErrorCount abort when the number of errors reaches this value
	SettingsAbortOnErrorCount = "settings.defaults.abortOnErrors"

	// SettingsViewOption controls whether views are applied the output or not
	SettingsViewOption = "settings.defaults.view"

	// SettingsTimeout timeout in seconds use when sending requests
	SettingsTimeout = "settings.defaults.timeout"

	// SettingsConfirmText custom confirmation text to use to prompt the user of an action
	SettingsConfirmText = "settings.defaults.confirmText"

	// SettingsProcessingMode default Cumulocity processing mode applied to
	// requests when the --processingMode flag is not given
	SettingsProcessingMode = "settings.defaults.processingMode"

	// SettingsJSONFlatten flatten nested json using dot notation
	SettingsJSONFlatten = "settings.defaults.flatten"

	// SettingsStorageStoreToken controls if the token is saved to the session file or not
	SettingsStorageStoreToken = "settings.storage.storeToken"

	// SettingsStorageStorePassword controls if the password is saved to the session file or not
	SettingsStorageStorePassword = "settings.storage.storePassword"

	// SettingsTemplatePath template folder where the template files are located
	SettingsTemplatePath = "settings.template.path"

	// SettingsTemplateCustomPaths custom template folder where the template files are located
	SettingsTemplateCustomPaths = "settings.template.customPath"

	// SettingsMode controls which commands/actions are enabled, e.g. dev, qual, prod
	SettingsMode = "settings.session.mode"

	// SettingsPinEntry sets the command to run to get the user's passphrase
	SettingsPinEntry = "settings.pinEntry"

	// SettingsModeCI enable continuous integration mode (this will enable all commands)
	SettingsModeCI = "settings.ci"

	// SettingsForce don't prompt for confirmation
	SettingsForce = "settings.defaults.force"

	// SettingsForceConfirm force prompt for confirmation
	SettingsForceConfirm = "settings.defaults.confirm"

	// SettingsModeConfirmation sets the confirm mode
	SettingsModeConfirmation = "settings.session.confirmation"

	// GetOutputFileRaw file path where the raw response will be saved to
	SettingsOutputFileRaw = "settings.defaults.outputFileRaw"

	// SettingsOutputFile file path where the parsed response will be saved to
	SettingsOutputFile = "settings.defaults.outputFile"

	// SettingsOutputFormat Output format i.e. table, json, csv, csvheader
	SettingsOutputFormat = "settings.defaults.output"

	// SettingsOutputTemplate Output jsonnet template
	SettingsOutputTemplate = "settings.defaults.outputTemplate"

	// SettingsEncryptionEnabled enables encryption when storing sensitive session data
	SettingsEncryptionEnabled = "settings.encryption.enabled"

	// SettingsActivityLogPath path where the activity log will be stored
	SettingsActivityLogPath = "settings.activityLog.path"

	// SettingsActivityLogEnabled enables/disables the activity log
	SettingsActivityLogEnabled = "settings.activityLog.enabled"

	// SettingsActivityLogMethodFilter filters the activity log entries by a space delimited methods, i.e. GET POST PUT
	SettingsActivityLogMethodFilter = "settings.activityLog.methodFilter"

	// SettingsConfigPath configuration path
	SettingsConfigPath = "settings.path"

	// SettingsViewsCommonPaths paths to common view definition files
	SettingsViewsCommonPaths = "settings.views.commonPaths"

	// SettingsViewsCustomPaths paths to custom view definition files
	SettingsViewsCustomPaths = "settings.views.customPaths"

	// SettingsFilter json filter to be applied to the output
	SettingsFilter = "settings.defaults.filter"

	// SettingsSelect json properties to be selected from the output. Only the given properties will be returned
	SettingsSelect = "settings.defaults.select"

	// SettingsSilentStatusCodes Status codes which will not print out an error message
	SettingsSilentStatusCodes = "settings.defaults.silentStatusCodes"

	// SettingsSilentExit silent status codes don't affect the exit code
	SettingsSilentExit = "settings.defaults.silentExit"

	// SettingsSessionFile Session file to use for api authentication
	SettingsSessionFile = "settings.defaults.session"

	// SettingsAliases list of aliases
	SettingsAliases = "settings.aliases"

	// SettingsCommonAliases list of common aliases which are usually kept in the global configuration and shared amongst sessions
	SettingsCommonAliases = "settings.commonAliases"

	// SettingsViewMinColumnWidth minimum column width in characters
	SettingsViewMinColumnWidth = "settings.views.columnMinWidth"

	// SettingsViewEmptyValueMinColumnWidth minimum column width in characters when a value is empty
	SettingsViewEmptyValueMinColumnWidth = "settings.views.columnMinWidthEmptyValue"

	// SettingsViewMaxColumnWidth maximum column width in characters
	SettingsViewMaxColumnWidth = "settings.views.columnMaxWidth"

	// SettingsViewColumnPadding column padding
	SettingsViewColumnPadding = "settings.views.columnPadding"

	// SettingsViewRowMode controls row rendering, e.g. wrapping or truncation
	SettingsViewRowMode = "settings.views.rowMode"

	// SettingsViewSampleSize maximum number of rows which are sampled when resolving the table columns and column widths
	SettingsViewSampleSize = "settings.views.sampleSize"

	// SettingsViewSampleTimeout maximum duration to buffer rows whilst waiting for more rows to be sampled, so that streamed output (e.g. realtime subscriptions) is not delayed indefinitely. Accepts a duration, where the default unit is milliseconds
	SettingsViewSampleTimeout = "settings.views.sampleTimeout"

	// SettingsViewNumberFormat number format
	SettingsViewNumberFormat = "settings.views.numberFormat"

	// SettingsViewNumbersMetricPrecision precision to use when using metrics
	SettingsViewNumbersMetricPrecision = "settings.views.metric.precision"

	// SettingsViewNumbersMetricActivateRangeMin minimum number that the metric prefix will be added to
	SettingsViewNumbersMetricActivateRangeMin = "settings.views.metric.rangeMin"

	// SettingsViewNumbersMetricActivateRangeMax maximum number that the metric prefix will be added to
	SettingsViewNumbersMetricActivateRangeMax = "settings.views.metric.rangeMax"

	// SettingsLoggerHideSensitive hide sensitive information in log entries
	SettingsLoggerHideSensitive = "settings.logger.hideSensitive"

	// SettingsDisableInput disable reading from stdin (pipeline input)
	SettingsDisableInput = "settings.defaults.nullInput"

	// SettingsAllowEmptyPipe allow empty piped data
	SettingsAllowEmptyPipe = "settings.defaults.allowEmptyPipe"

	// SettingsSessionProviderType provider to use when setting a session
	SettingsSessionProviderType = "settings.session.provider.type"

	// SettingsSessionProviderCommand sets the command to run when running set-session
	SettingsSessionProviderCommand = "settings.session.provider.command"

	// SettingsSessionProviderSecrets the secrets which should be included as environment variables when calling the external provider command
	SettingsSessionProviderSecrets = "settings.session.provider.secrets"

	// SettingsLoginType preferred login type, i.e. BASIC, OAUTH2_INTERNAL etc.
	SettingsLoginType = "settings.login.type"

	// SettingsSessionAlwaysIncludePassword should the password always be included in the session variables or not
	SettingsSessionAlwaysIncludePassword = "settings.session.alwaysIncludePassword"

	// SettingsSessionTokenValidFor interval which the token must be valid for in order to reuse it
	SettingsSessionTokenValidFor = "settings.session.tokenValidFor"

	// SettingsSessionHide hide sensitive information in the session banner
	SettingsSessionHide = "settings.session.hide"

	// Cache settings
	// SettingsDefaultsCacheEnabled enable caching
	SettingsDefaultsCacheEnabled = "settings.defaults.cache"
	SettingsDefaultsNoCache      = "settings.defaults.noCache"

	// SettingsDefaultsCacheTTL Cache time-to-live setting as a duration
	SettingsDefaultsCacheTTL = "settings.defaults.cacheTTL"

	// SettingsCacheDir Cache directory
	SettingsCacheDir = "settings.cache.path"

	// SettingsCacheMethods HTTP methods which should be cached
	SettingsCacheMethods = "settings.cache.methods"

	// SettingsCacheKeyHost include host in cache key generation
	SettingsCacheKeyHost = "settings.cache.keyhost"

	// SettingsCacheMode cache mode. Only used for testing purposes
	SettingsCacheMode = "settings.cache.mode"

	// SettingsCacheKeyAuth include authorization header in cache key generation
	SettingsCacheKeyAuth = "settings.cache.keyauth"

	// SettingsCacheBodyPaths include only specific json body paths in cache hashing calculation
	SettingsCacheBodyPaths = "settings.defaults.cacheBodyPaths"

	// SettingsDefaultsInsecure allow insecure SSL connections
	SettingsDefaultsInsecure = "settings.defaults.insecure"

	// SettingsBrowser default browser
	SettingsBrowser = "settings.browser"

	// SettingsSSODiscoveryUrl Open ID Connect URL aka. Discovery URL
	SettingsSSODiscoveryUrl = "settings.sso.discoveryUrl"

	// SettingsSSOScopes SSO scopes used to request a device code
	SettingsSSOScopes = "settings.sso.scopes"

	// SettingsBrowserCallbackURL custom redirect URI for the browser flow callback server
	SettingsBrowserCallbackURL = "settings.sso.browserCallbackUrl"

	//
	// Remote Access preferences
	//
	// SettingsRemoteAccessDefaultSSHUser the default ssh user
	SettingsRemoteAccessDefaultSSHUser = "settings.remoteaccess.sshuser"

	// Extensions
	SettingsExtensionDataDir         = "settings.extensions.datadir"
	SettingsExtensionDefaultHost     = "settings.extensions.defaultHost"
	SettingsExtensionDefaultUsername = "settings.extensions.defaultUsername"
)

var (
	SettingsDefaultsPrefix = "settings.defaults"
)

const (
	ViewsOff  = "off"
	ViewsAuto = "auto"
)

// GetSettingsName get the settings name from a flag name
func GetSettingsName(flagName string) string {
	return SettingsDefaultsPrefix + "." + flagName
}

// GetSettingsNameWithoutPrefix converts the setting name without the settings prefix
func GetSettingsNameWithoutPrefix(name string) string {
	return strings.TrimPrefix(name, SettingsDefaultsPrefix+".")
}

// Config cli configuration settings
type Config struct {
	viper *viper.Viper

	// Persistent settings (stored to file)
	Persistent *viper.Viper

	// SecureData accessor to encrypt/decrypt data
	SecureData *encrypt.SecureData

	// Passphrase used for encrypting/decrypting fields
	Passphrase string

	prompter prompt.Prompt

	// SecretText used to test the encryption passphrase
	SecretText string

	Logger *logger.Logger

	sessionFile string

	// private caching to improve performance
	// by preventing expensive system calls for each iteration
	outputFileRaw    *string
	outputFile       *string
	outputTemplate   *string
	commonOptions    *CommonCommandOptions
	templateResolver flags.Resolver
}

// NewConfig returns a new CLI configuration object
func NewConfig(v *viper.Viper) *Config {

	passphrase := os.Getenv(EnvPassphrase)

	c := &Config{
		viper:      v,
		Passphrase: passphrase,
		SecureData: encrypt.NewSecureData("{encrypted}"),
		Persistent: viper.New(),
		prompter:   prompt.Prompt{},
		Logger:     logger.NewDummyLogger("SecureData"),
	}
	c.prompter.Logger = c.Logger
	c.bindSettings()
	return c
}

// Set non-persisted values
func (c *Config) Set(key string, value any) {
	c.viper.Set(key, value)
}

func (c *Config) RegisterTemplateResolver(resolver flags.Resolver) {
	c.templateResolver = resolver
}

// Option cli configuration option
type Option func(*Config) error

func WithBindEnv(name string, defaultValue interface{}) func(*Config) error {
	return func(c *Config) error {
		return c.bindEnv(name, defaultValue)
	}
}

func WithDefault(name string, defaultValue interface{}) func(*Config) error {
	return func(c *Config) error {
		c.viper.SetDefault(name, defaultValue)
		return nil
	}
}

// WithBoolEnvOverride support optional override boolean variable
func WithBoolEnvOverride(name string, envName string) func(*Config) error {
	return func(c *Config) error {
		if v, err := strconv.ParseBool(os.Getenv(envName)); err == nil {
			c.viper.Set(name, v)
		}
		return nil
	}
}

// WithBoolEnvOverrides sets the configuration if the given env variable function returns true and no error
func WithBoolEnvOverrides(name string, valueFunc func(k, v string) (bool, error), envNames ...string) func(*Config) error {
	return func(c *Config) error {
		for _, envName := range envNames {
			if valueFunc != nil {
				if v, err := valueFunc(envName, os.Getenv(envName)); v && err == nil {
					c.viper.Set(name, v)
				}
			}
		}
		return nil
	}
}

// WithStringEnvOverride supports optional overriding a string value from another env variable
func WithStringEnvOverride(name string, envName string) func(*Config) error {
	return func(c *Config) error {
		if v := os.Getenv(envName); v != "" {
			c.viper.Set(name, v)
		}
		return nil
	}
}

func (c *Config) WithOptions(opts ...Option) error {
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) bindSettings() {
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.viper.SetEnvPrefix(EnvSettingsPrefix)
	err := c.WithOptions(
		WithBindEnv(SettingEncryptionCachePassphrase, false),
		WithBindEnv(SettingsMaxWorkers, 50),
		WithBindEnv(SettingsWorkers, 1),
		WithBindEnv(SettingsIncludeAllPageSize, 2000),
		WithBindEnv(SettingsStorageStorePassword, true),
		WithBindEnv(SettingsStorageStoreToken, true),
		WithBindEnv(SettingsModeConfirmation, "PUT POST DELETE"),
		WithBindEnv(SettingsProcessingMode, nil),

		WithBindEnv(SettingsEncryptionEnabled, true),
		WithBindEnv(SettingsActivityLogEnabled, true),
		WithBindEnv(SettingsActivityLogPath, path.Join(c.GetSessionHomeDir(), ActivityLogDirName)),
		WithBindEnv(SettingsActivityLogMethodFilter, "GET PUT POST DELETE"),

		// HTTP settings
		WithBindEnv(SettingsUseCompression, true),
		WithBindEnv(SettingsHTTPMaxRetries, 0),
		WithBindEnv(SettingsHTTPRetryWaitMax, "50s"),
		WithBindEnv(SettingsHTTPRetryWaitMin, "5s"),

		// Dry run options
		WithBindEnv(SettingsDryRunPattern, ""),

		WithBindEnv(SettingsIncludeAllDelayMS, 50),
		WithBindEnv(SettingsTemplatePath, ""),
		WithBindEnv(SettingsMode, SessionModeProduction.String()),

		WithBindEnv(SettingsModeCI, false),
		// Determines if the current execution context is within a known CI/CD system.
		// This is based on https://github.com/watson/ci-info/blob/HEAD/index.js.
		WithBoolEnvOverrides(
			SettingsModeCI,
			func(k, v string) (bool, error) {
				if k == "CI" {
					return v != "false" && v != "", nil
				}
				return v != "", nil
			},
			"CI",                     // Travis CI, CircleCI, Cirrus CI, Gitlab CI, Appveyor, CodeShip, dsari, Cloudflare Pages
			"BUILD_ID",               // Jenkins, TeamCity
			"BUILD_NUMBER",           // Jenkins, TeamCity
			"RUN_ID",                 // TaskCluster, dsari
			"CI_APP_ID",              // Appflow
			"CI_BUILD_ID",            // Appflow
			"CI_BUILD_NUMBER",        // Appflow
			"CI_NAME",                // Codeship and others
			"CONTINUOUS_INTEGRATION", // Travis CI, Cirrus CI
		),

		// Support overriding the settings.session.mode value with the C8Y_MODE env variable
		WithStringEnvOverride(SettingsMode, EnvSessionMode),

		WithBindEnv(SettingsConfigPath, ""),
		WithBindEnv(SettingsViewsCommonPaths, ""),
		WithBindEnv(SettingsViewsCustomPaths, ""),

		WithBindEnv(SettingsViewMinColumnWidth, 2),
		WithBindEnv(SettingsViewEmptyValueMinColumnWidth, 15),
		WithBindEnv(SettingsViewMaxColumnWidth, 80),
		WithBindEnv(SettingsViewColumnPadding, 15),
		WithBindEnv(SettingsViewRowMode, "truncate"),
		WithBindEnv(SettingsViewSampleSize, 5),
		WithBindEnv(SettingsViewSampleTimeout, "500ms"),

		// Table number formatter
		WithBindEnv(SettingsViewNumberFormat, NumberFormatMetric),
		WithBindEnv(SettingsViewNumbersMetricPrecision, 2),
		WithBindEnv(SettingsViewNumbersMetricActivateRangeMin, 0.00001),
		WithBindEnv(SettingsViewNumbersMetricActivateRangeMax, 100000),

		WithBindEnv(SettingsLoggerHideSensitive, true),

		WithBindEnv(SettingsCacheMethods, "GET PUT POST DELETE"),
		WithBindEnv(SettingsCacheKeyHost, true),
		WithBindEnv(SettingsCacheKeyAuth, true),
		WithBindEnv(SettingsCacheBodyPaths, ""),
		WithBindEnv(SettingsCacheMode, nil),
		WithBindEnv(SettingsCacheDir, filepath.Join(os.TempDir(), "go-c8y-cli-cache")),

		// Console options
		WithBindEnv(SettingsForceTTY, false),

		// Session options
		WithBindEnv(SettingsSessionProviderType, ProviderTypeExternal),
		WithBindEnv(SettingsSessionProviderCommand, "c8y sessions set --no-banner --output json"),
		WithBindEnv(SettingsSessionProviderSecrets, ""),
		WithBindEnv(SettingsPinEntry, ""),
		WithBindEnv(SettingsSessionAlwaysIncludePassword, false),
		WithBindEnv(SettingsSessionTokenValidFor, "8h"),
		WithBindEnv(SettingsSessionHide, false),
		WithBindEnv(SettingsLoginType, ""),

		WithBindEnv(SettingsBrowser, ""),
		WithBindEnv(SettingsBrowserCallbackURL, ""),

		// Extensions
		WithBindEnv(SettingsExtensionDataDir, ""),
		WithBindEnv(SettingsExtensionDefaultHost, "github.com"),
	)

	if err != nil {
		c.Logger.Warnf("Could not bind settings. %s", err)
	}

	// Set pin entry command
	c.prompter.PinEntry = c.PinEntry()
}

// SetLogger sets the logger
func (c *Config) SetLogger(l *logger.Logger) {
	c.Logger = l
	c.prompter.Logger = l
}

// ReadConfig reads the given file and loads it into the persistent session config
func (c *Config) ReadConfig(file string) error {
	c.Persistent.SetConfigFile(file)
	return c.Persistent.ReadInConfig()
}

// PinEntry returns the command to use to request a user's credentials
func (c *Config) PinEntry() string {
	return c.viper.GetString(SettingsPinEntry)
}

// CheckEncryption checks if the user has provided the correct encryption password or not by testing the decryption of the secret text
func (c *Config) CheckEncryption(encryptedText ...string) (string, error) {
	secretText := c.SecretText
	if len(encryptedText) > 0 {
		secretText = encryptedText[0]
	}

	c.Logger.Infof("Checking encryption passphrase against secret text: %s", secretText)
	pass, err := c.prompter.EncryptionPassphrase(secretText, EnvPassphrase, c.Passphrase, "")
	c.Passphrase = pass
	return pass, err
}

// PromptPassphrase prompts the user for the passphrase if it is not already set
func (c *Config) PromptPassphrase() (string, error) {
	if c.Passphrase != "" {
		return c.Passphrase, nil
	}
	prompter, err := c.prompter.GetPassphrasePrompter(EnvPassphrase)
	if err != nil {
		return "", err
	}
	pass, err := prompter.Prompt(0, 1)
	return pass, err
}

// PromptSecret prompts the user for the passphrase if it is not already set
func (c *Config) PromptSecret(key string) (string, error) {
	prompter, err := c.prompter.GetExternalPrompter(key)
	if err != nil {
		return "", err
	}
	pass, err := prompter.Prompt(0, 1)
	return pass, err
}

// BindAuthorization binds environment variables related to the authorization to the configuration
func (c *Config) BindAuthorization() error {
	c.viper.SetEnvPrefix(EnvSettingsPrefix)
	auth_variables := [...]string{
		"host",
		"username",
		"tenant",
		"password",
		"token",
		"credential.totp.secret",
		"certificate",
		"certificate_key",
	}
	for _, name := range auth_variables {
		if err := c.viper.BindEnv(name); err != nil {
			return err
		}
	}
	return nil
}

// GetUsername returns the Cumulocity username for the session
func (c *Config) GetUsername() string {
	if v := c.GetSessionUsername(); v != "" {
		c.Logger.Infof("Using session username override")
		return v
	}
	v := c.viper.GetString("username")

	if v != "" {
		return v
	}
	return os.Getenv("C8Y_USER")
}

// GetName returns the name of the current session
func (c *Config) GetName() string {
	return c.viper.GetString("name")
}

// GetDescription returns the name description of the current session
func (c *Config) GetDescription() string {
	return c.viper.GetString("name")
}

// GetTenant returns the Cumulocity tenant id
func (c *Config) GetTenant() string {
	// check for an empty or "null" tenant name as jq outputs null if
	// a json property is not found, so the user might accidentally provide
	// null without knowing it
	if v := c.viper.GetString("tenant"); v != "" && v != "null" {
		return v
	}
	if v := c.Persistent.GetString("tenant"); v != "" && v != "null" {
		return v
	}
	return ""
}

// GetHost returns the Cumulocity host URL
func (c *Config) GetHost() string {
	return strings.TrimSpace(c.viper.GetString("host"))
}

// GetDomain gets the custom Cumulocity domain for cases where it differs from the Host
func (c *Config) GetDomain() string {
	host := c.GetHost()
	if !strings.Contains(host, "://") {
		host = "https://" + host
	}
	if domain, err := url.Parse(host); err == nil {
		return domain.Host
	}
	return c.GetHost()
}

// GetTOTP returns a TOTP generated by a TOTP secret (if present)
func (c *Config) GetTOTP(t time.Time) (string, error) {
	return totp.GenerateTOTP(c.viper.GetString("credential.totp.secret"), t)
}

// CreateKeyFile creates a file used as reference to validate encryption
func (c *Config) CreateKeyFile(keyText string) error {
	if _, err := os.Stat(c.KeyFile()); os.IsExist(err) {
		c.Logger.Infof("Key file already exists. file=%s", c.KeyFile)
		return nil
	}
	key, err := os.Create(c.KeyFile())
	if err != nil {
		return err
	}

	if _, err := key.WriteString(keyText); err != nil {
		return err
	}
	return nil
}

// KeyFile path to the key file used to test encryption
func (c *Config) KeyFile() string {
	return path.Join(c.GetSessionHomeDir(), KeyFileName)
}

// ReadKeyFile reads the key file used as a reference to validate encryption (i.e. when no sessions exist)
func (c *Config) ReadKeyFile() error {

	// read from env variable
	if v := os.Getenv(EnvPassphraseText); v != "" && c.SecureData.IsEncrypted(v) == 1 {
		c.Logger.Infof("Using env variable '%s' as example encryption text", EnvPassphraseText)
		c.SecretText = v
		return c.CreateKeyFile(v)
	}

	// read from file
	contents, err := os.ReadFile(c.KeyFile())

	if err == nil {
		if c.SecureData.IsEncryptedBytes(contents) == 1 {
			c.SecretText = string(contents)
			return nil
		}
		c.Logger.Warningf("Key file is invalid or contains decrypted information")
	}

	// init key file
	passphrase := os.Getenv(EnvPassphrase)

	if passphrase == "" {
		// prompt for passphrase
		passphrase, err = c.prompter.PasswordWithConfirm("new encryption passphrase", "Creating a encryption key for sessions")
		if err != nil {
			return err
		}
	}

	c.Passphrase = passphrase

	keyText, err := c.SecureData.EncryptString("Cumulocity CLI Tool", c.Passphrase)

	if err != nil {
		return err
	}

	if err := c.CreateKeyFile(keyText); err != nil {
		return err
	}

	c.SecretText = keyText
	return nil
}

// HasEncryptedProperties check if some fields are encrypted
func (c Config) HasEncryptedProperties() bool {
	encryptedKeys := []string{}
	for key, value := range c.Persistent.AllSettings() {
		if s, ok := value.(string); ok {
			if c.SecureData.IsEncrypted(s) > 0 {
				encryptedKeys = append(encryptedKeys, key)
			}
		}
	}
	return len(encryptedKeys) > 0
}

// DecryptAllProperties decrypt all properties
func (c Config) DecryptAllProperties() (err error) {
	for key, value := range c.Persistent.AllSettings() {
		if s, ok := value.(string); ok {

			if c.SecureData.IsEncrypted(s) > 0 {
				ds, err := c.SecureData.DecryptString(s, c.Passphrase)

				if err != nil {
					return ErrDecrypt{err}
				}
				c.Persistent.Set(key, ds)
			}
		}
	}
	return err
}

func GetEnvKey(key string) string {
	return "C8Y_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
}

// GetEnvKey returns the environment key value associated
func (c Config) GetEnvKey(key string) string {
	return GetEnvKey(key)
}

// HasEnvSettingsPrefix check if a given env variable name is a settings variable
func (c Config) HasEnvSettingsPrefix(envName string) bool {
	return strings.HasPrefix(envName, "C8Y_SETTINGS_")
}

var SettingsToken = "token"

// GetToken return the decrypted token from the current session
func (c *Config) GetToken() (string, error) {
	value := c.viper.GetString(SettingsToken)

	if value == "" {
		value = c.Persistent.GetString(SettingsToken)
	}

	decryptedValue, err := c.DecryptString(value)
	if err != nil {
		return value, err
	}
	return decryptedValue, nil
}

// DebugViper debug viper configuration
func (c Config) DebugViper() {
	c.viper.Debug()
}

// DecryptString returns the decrypted string if the string is encrypted
func (c *Config) DecryptString(value string) (string, error) {
	if c.SecureData.IsEncrypted(value) > 0 {
		c.Logger.Infof("Decrypting data. %s", value)
	}
	value, err := c.SecureData.TryDecryptString(value, c.Passphrase)
	return value, err
}

// GetEncryptedString returns string value of a potentially encrypted field in the configuration
// If the fields starts with the encrypted prefix, then it will be decrypted using the CLI passphrase,
// otherwise the value will be returned as is.
func (c *Config) GetEncryptedString(key string) string {
	value := c.viper.GetString(key)

	decryptedValue, err := c.DecryptString(value)
	if err != nil {
		return value
	}
	return decryptedValue
}

// SetEncryptedString encrypts and sets a value in the configuration. If the give value is empty, then the value will be read from the configuration file
func (c *Config) SetEncryptedString(key, value string) error {
	if value == "" {
		value = c.Persistent.GetString(key)
	}

	if value == "" {
		c.Logger.Info("Password is not set so nothing to encrypt")
		return nil
	}

	var err error
	password := value
	if c.IsEncryptionEnabled() {
		password, err = c.SecureData.TryEncryptString(value, c.Passphrase)

		if err != nil {
			return err
		}
	}

	c.Persistent.Set(key, password)
	return nil
}

// WritePersistentConfig saves the configuration to file
func (c *Config) WritePersistentConfig() error {
	file := c.viper.ConfigFileUsed()

	if file == "" {
		return fmt.Errorf("No config is being used")
	}
	c.Persistent.Set("$schema", "https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/v2/tools/schema/session.schema.json")

	err := c.SetEncryptedString("password", "")
	if err != nil {
		return err
	}
	err = c.SetEncryptedString("token", "")
	if err != nil {
		return err
	}
	return c.Persistent.WriteConfig()
}

// GetPassword returns the decrypted password of the current session
func (c *Config) GetPassword() (string, error) {
	if v := c.GetSessionPassword(); v != "" {
		c.Logger.Infof("Using session password override")
		return v, nil
	}

	value := c.GetPasswordRaw()

	if value == "" {
		value = c.Persistent.GetString("password")
	}

	decryptedValue, err := c.DecryptString(value)
	if err != nil {
		return value, err
	}
	return decryptedValue, nil
}

func (c *Config) GetPasswordRaw() string {
	return c.viper.GetString("password")
}

// IsPasswordEncrypted return true if the password is encrypted
// If the password is empty then treat it as encrypted
func (c *Config) IsPasswordEncrypted(ignoreEmptyValue ...bool) bool {
	password := c.GetPasswordRaw()
	if len(ignoreEmptyValue) > 0 && ignoreEmptyValue[0] {
		return c.SecureData.IsEncrypted(password) == 1
	}
	return password == "" || c.SecureData.IsEncrypted(password) == 1
}

func (c *Config) IsTokenEncrypted(ignoreEmptyValue ...bool) bool {
	token := c.viper.GetString("token")
	if len(ignoreEmptyValue) > 0 && ignoreEmptyValue[0] {
		return c.SecureData.IsEncrypted(token) == 1
	}
	return token == "" || c.SecureData.IsEncrypted(token) == 1
}

// MustGetPassword returns the decrypted password if there are no encryption errors, otherwise it will return an encrypted password
func (c *Config) MustGetPassword() string {
	decryptedValue, err := c.GetPassword()
	if err != nil {
		c.Logger.Warningf("Could not decrypt password. %s", err)
	}
	return decryptedValue
}

// MustGetToken returns the decrypted token if there are no encryption errors, otherwise it will return an encrypted value
func (c *Config) MustGetToken(silent bool) string {
	decryptedValue, err := c.GetToken()
	if err != nil {
		if !silent {
			c.Logger.Warningf("Could not decrypt token. %s", err)
		}
	}
	return decryptedValue
}

// SetCumulocityVersion sets the Cumulocity version
func (c *Config) SetCumulocityVersion(p string) {
	c.Persistent.Set("version", p)
}

// GetCumulocityVersion gets Cumulocity version
func (c *Config) GetCumulocityVersion() string {
	return c.Persistent.GetString("version")
}

// SetUsername sets the username
func (c *Config) SetUsername(v string) {
	c.Persistent.Set("username", v)
}

// SetPassword sets the password
func (c *Config) SetPassword(p string) {
	c.Persistent.Set("password", p)
}

// SetToken sets the token used for OAUTH authentication
func (c *Config) SetToken(p string) {
	c.Persistent.Set(SettingsToken, p)
}

// SetToken sets the token used for OAUTH authentication
func (c *Config) ClearToken() {
	c.viper.Set(SettingsToken, "")
	c.Persistent.Set(SettingsToken, "")
}

// SetTenant sets the tenant name
func (c *Config) SetTenant(value string) {
	c.Persistent.Set("tenant", value)
}

// IsCIMode return true if the cli is running in CI mode
func (c *Config) IsCIMode() bool {
	return c.viper.GetBool(SettingsModeCI)
}

// IsEncryptionEnabled indicates if session encryption is enabled or not
func (c *Config) IsEncryptionEnabled() bool {
	return c.viper.GetBool(SettingsEncryptionEnabled)
}

// GetString returns a string from the configuration
func (c *Config) GetString(key string) string {
	return c.viper.GetString(key)
}

// GetStringSlice returns a slice of strings
func (c *Config) GetStringSlice(key string) []string {
	return c.viper.GetStringSlice(key)
}

// GetDefaultUsername returns the default username
func (c *Config) GetDefaultUsername() string {
	return c.viper.GetString("settings.session.defaultUsername")
}

// SessionCommand returns session provider
func (c *Config) SessionProvider() string {
	return c.viper.GetString(SettingsSessionProviderType)
}

// SessionProviderCommand returns the command to use when logging into a session
func (c *Config) SessionProviderCommand() string {
	return c.viper.GetString(SettingsSessionProviderCommand)
}

// SessionProviderSecrets returns the env variables to be included in the external command
func (c *Config) SessionProviderSecrets() []string {
	return c.viper.GetStringSlice(SettingsSessionProviderSecrets)
}

// AlwaysIncludePassword password when setting a session
func (c *Config) AlwaysIncludePassword() bool {
	return c.viper.GetBool(SettingsSessionAlwaysIncludePassword)
}

// TokenValidFor minimum validity of a token in order to reuse it
func (c *Config) TokenValidFor() time.Duration {
	value := c.viper.GetString(SettingsSessionTokenValidFor)
	duration, err := flags.GetDuration(value, true, time.Second)
	if err != nil {
		c.Logger.Warnf("Invalid duration. value=%s, err=%s", duration, err)
		return 0
	}
	return duration
}

// CachePassphraseVariables return true if the passphrase variables should be persisted or not
func (c *Config) CachePassphraseVariables() bool {
	return c.viper.GetBool(SettingEncryptionCachePassphrase)
}

func (c *Config) bindEnv(name string, defaultValue interface{}) error {
	err := c.viper.BindEnv(name)
	if defaultValue != nil {
		c.viper.SetDefault(name, defaultValue)
	}
	return err
}

// DecryptSession decrypts a session (as long as the encryption passphrase has already been provided)
func (c *Config) DecryptSession() error {
	c.SetPassword(c.MustGetPassword())
	c.SetToken(c.MustGetToken(false))
	return c.WritePersistentConfig()
}

// CommonAliases Get common aliases from the global configuration file
// deprecated in favor of extensions
func (c *Config) CommonAliases() map[string]string {
	return map[string]string{}
	// return c.viper.GetStringMapString(SettingsCommonAliases)
}

// Aliases get aliases configured in the current session
func (c *Config) Aliases() map[string]string {
	return c.Persistent.GetStringMapString(SettingsAliases)
}

// SetAliases set aliases for the current session
func (c *Config) SetAliases(v map[string]string) {
	c.Persistent.Set(SettingsAliases, v)
}

// GetMaxWorkers maximum number of workers allowed. If the number of works is larger than this value then an error will be raised
func (c *Config) GetMaxWorkers() int {
	return c.viper.GetInt(SettingsMaxWorkers)
}

// GetWorkers number of workers to use. If the total workers exceeds the maximum allowed workers then a warning will be logged and the maximum value will be used instead.
func (c *Config) GetWorkers() int {
	workers := c.viper.GetInt(SettingsWorkers)
	maxWorkers := c.GetMaxWorkers()
	if workers > maxWorkers {
		workers = maxWorkers
		c.Logger.Warningf("number of workers exceeds the maximum workers limit of %d. Using maximum value (%d) instead", maxWorkers, maxWorkers)
	}
	return workers
}

// GetMaxJobs maximum number of jobs allowed to run
func (c *Config) GetMaxJobs() int64 {
	return c.viper.GetInt64(SettingsMaxJobs)
}

// GetIncludeAllPageSize get page size used for include all pagination
func (c *Config) GetIncludeAllPageSize() int {
	return c.viper.GetInt(SettingsIncludeAllPageSize)
}

// GetPageSize get page size
func (c *Config) GetPageSize() int {
	return c.viper.GetInt(SettingsPageSize)
}

// GetCurrentPage get current page
func (c *Config) GetCurrentPage() int64 {
	return c.viper.GetInt64(SettingsCurrentPage)
}

// GetTotalPages get total pages to return
func (c *Config) GetTotalPages() int64 {
	return c.viper.GetInt64(SettingsTotalPages)
}

// IncludeAll return all available results
func (c *Config) IncludeAll() bool {
	return c.viper.GetBool(SettingsIncludeAll)
}

// PaginationStrategy returns the configured pagination strategy used when
// iterating results: "" (auto), "offset", "id" or "time". Empty means the SDK
// picks the optimal strategy per entity.
func (c *Config) PaginationStrategy() string {
	return c.viper.GetString(SettingsPaginationStrategy)
}

func (c *Config) MaxItems() int64 {
	if c.IncludeAll() {
		return 0
	}
	if totalPages := c.GetTotalPages(); totalPages > 0 {
		return totalPages * int64(c.GetPageSize())
	}
	return int64(c.GetPageSize())
}

// GetIncludeAllDelay include all delay in milliseconds
func (c *Config) GetIncludeAllDelay() int64 {
	return c.viper.GetInt64(SettingsIncludeAllDelayMS)
}

// WithTotalPages return all available results
func (c *Config) WithTotalPages() bool {
	return c.viper.GetBool(SettingsWithTotalPages)
}

// WithTotalElements return total of all elements
func (c *Config) WithTotalElements() bool {
	return c.viper.GetBool(SettingsWithTotalElements)
}

// RawOutput return raw (original) response
func (c *Config) RawOutput() bool {
	return c.viper.GetBool(SettingsRawOutput)
}

// IgnoreAcceptHeader ignore accept header
func (c *Config) IgnoreAcceptHeader() bool {
	return c.viper.GetBool(SettingsIgnoreAcceptHeader)
}

// GetHeader get custom headers
func (c *Config) GetHeader() []string {
	return c.viper.GetStringSlice(SettingsHeader)
}

// GetQueryParameters get custom query parameters
func (c *Config) GetQueryParameters() []string {
	return c.viper.GetStringSlice(SettingsQueryParameters)
}

// DryRun don't sent any destructive requests. Just print out what would be sent
func (c *Config) DryRun() bool {
	return c.viper.GetBool(SettingsDryRun)
}

// SettingsDryRunFormat dry run output format. Controls how the dry run information is displayed
func (c *Config) DryRunFormat() string {
	return c.viper.GetString(SettingsDryRunFormat)
}

// GetDryRunPattern pattern used to check if a command should be run using dry run or not if dry run is activated
func (c *Config) GetDryRunPattern() string {
	return c.viper.GetString(SettingsDryRunPattern)
}

func (c *Config) ShouldUseCompression() bool {
	return c.viper.GetBool(SettingsUseCompression)
}

// HTTPRetryMax get the maximum number of retries on failed http requests
func (c *Config) HTTPRetryMax() int {
	return c.viper.GetInt(SettingsHTTPMaxRetries)
}

// HTTPRetryWaitMax get the maximum wait time between failed http requests
func (c *Config) HTTPRetryWaitMax() time.Duration {
	value := c.viper.GetString(SettingsHTTPRetryWaitMax)
	duration, err := flags.GetDuration(value, true, time.Second)
	if err != nil {
		c.Logger.Warnf("Invalid duration. value=%s, err=%s", duration, err)
		return 0
	}
	return duration
}

// HTTPRetryWaitMin get the minimum wait time between failed http requests
func (c *Config) HTTPRetryWaitMin() time.Duration {
	value := c.viper.GetString(SettingsHTTPRetryWaitMin)
	duration, err := flags.GetDuration(value, true, time.Second)
	if err != nil {
		c.Logger.Warnf("Invalid duration. value=%s, err=%s", duration, err)
		return 0
	}
	return duration
}

// ShouldUseDryRun returns true of dry run should be applied to the command based on the type of method
func (c *Config) ShouldUseDryRun(commandLine string) bool {
	if c.DryRun() {
		pattern := c.GetDryRunPattern()
		if pattern == "" || commandLine == "" {
			return true
		}
		shouldInvert := false
		if strings.HasPrefix(pattern, "!") {
			shouldInvert = true
			pattern = pattern[1:]
		}
		if m, err := regexp.MatchString(pattern, commandLine); err != nil {
			if c.Logger != nil {
				c.Logger.Warnf("Invalid dry run pattern. pattern=%s, err=%s", commandLine, err)
			}
		} else {

			if shouldInvert {
				c.Logger.Infof("Should use dry run: pattern=%s, result=%v", pattern, !m)
				return !m
			}
			c.Logger.Infof("Should use dry run: pattern=%s, result=%v", pattern, m)
			return m
		}
	}
	return false
}

// Debug show debug messages
func (c *Config) Debug() bool {
	return c.viper.GetBool(SettingsDebug)
}

// Verbose show verbose messages
func (c *Config) Verbose() bool {
	return c.viper.GetBool(SettingsVerbose)
}

// CompactJSON show compact json output
func (c *Config) CompactJSON() bool {
	return c.viper.GetBool(SettingsJSONCompact)
}

// ShowProgress show progress bar
func (c *Config) ShowProgress() bool {
	return c.viper.GetBool(SettingsShowProgress) && !c.DisableProgress()
}

func (c *Config) ForceTTY() bool {
	return c.viper.GetBool(SettingsForceTTY)
}

func (c *Config) GetProgressBar(w io.Writer, enable bool) (progress *mpb.Progress) {
	if enable && !c.DisableProgress() {
		progress = mpb.New(
			mpb.WithOutput(w),
			mpb.WithRefreshRate(180*time.Millisecond),
		)
	}
	return
}

// DisableProgress don't print progress bar
func (c *Config) DisableProgress() bool {
	return c.viper.GetBool(SettingsDisableProgress)
}

// DisableColor don't print console output in color
func (c *Config) DisableColor() bool {
	return c.viper.GetBool(SettingsDisableColor)
}

// Proxy http/s proxy settings
func (c *Config) Proxy() string {
	return strings.TrimSpace(c.viper.GetString(SettingsProxy))
}

// IgnoreProxy ignore proxy settings
func (c *Config) IgnoreProxy() bool {
	return c.viper.GetBool(SettingsIgnoreProxy)
}

// WithError return the error response on stdout rather than stderr
func (c *Config) WithError() bool {
	return c.viper.GetBool(SettingsWithError)
}

// WorkerDelay delay in milliseconds to wait after each request before the worker processes a new job (request)
func (c *Config) WorkerDelay() time.Duration {
	return c.getDuration(SettingsWorkerDelay)
}

// WorkerDelayBefore delay in milliseconds to wait before each request
func (c *Config) WorkerDelayBefore() time.Duration {
	return c.getDuration(SettingsWorkerDelayBefore)
}

func (c *Config) getDuration(name string) time.Duration {
	v := c.viper.GetString(name)
	duration, err := flags.GetDuration(v, true, time.Millisecond)
	if err != nil {
		c.Logger.Warnf("Invalid duration. value=%s, err=%s", v, err)
		return 0
	}
	return duration
}

// AbortOnErrorTotal abort when the number of errors reaches this value
func (c *Config) AbortOnErrorCount() int {
	return c.viper.GetInt(SettingsAbortOnErrorCount)
}

// ViewOption controls whether views are applied the output or not
func (c *Config) ViewOption() string {
	if c.RawOutput() {
		return ViewsOff
	}

	// If view is not set by the user, and an output template is being
	// used, then turn off the views as the output template will most
	// likely change the structure significantly
	if !c.viper.IsSet(SettingsViewOption) && c.GetOutputTemplate() != "" {
		return ViewsOff
	}
	return c.viper.GetString(SettingsViewOption)
}

// ViewColumnMinWidth minimum column width in characters
func (c *Config) ViewColumnMinWidth() int {
	return c.viper.GetInt(SettingsViewMinColumnWidth)
}

// ViewColumnEmptyValueMinWidth minimum column width in characters
func (c *Config) ViewColumnEmptyValueMinWidth() int {
	return c.viper.GetInt(SettingsViewEmptyValueMinColumnWidth)
}

// ViewColumnMinWidth maximum column width in characters
func (c *Config) ViewColumnMaxWidth() int {
	return c.viper.GetInt(SettingsViewMaxColumnWidth)
}

// ViewColumnPadding column padding
func (c *Config) ViewColumnPadding() int {
	return c.viper.GetInt(SettingsViewColumnPadding)
}

// ViewRowMode get view row rendering mode (truncation or wrapping)
func (c *Config) ViewRowMode() string {
	return strings.ToLower(c.viper.GetString(SettingsViewRowMode))
}

// ViewSampleSize maximum number of rows which are sampled when resolving the table columns and column widths
func (c *Config) ViewSampleSize() int {
	return c.viper.GetInt(SettingsViewSampleSize)
}

// ViewSampleTimeout maximum duration to buffer rows whilst waiting for more rows to be sampled
func (c *Config) ViewSampleTimeout() time.Duration {
	return c.getDuration(SettingsViewSampleTimeout)
}

// RequestTimeout timeout to use when sending requests
func (c *Config) RequestTimeout() time.Duration {
	value := c.viper.GetString(SettingsTimeout)
	duration, err := flags.GetDuration(value, true, time.Second)
	if err != nil {
		c.Logger.Warnf("Invalid duration. value=%s, err=%s", duration, err)
		return 0
	}
	return duration
}

// FlattenJSON flatten nested json using dot notation
func (c *Config) FlattenJSON() bool {
	return c.viper.GetBool(SettingsJSONFlatten)
}

// ConfirmText custom confirmation text to use to prompt the user of an action
func (c *Config) ConfirmText() string {
	return c.viper.GetString(SettingsConfirmText)
}

// GetProcessingMode returns the session/default Cumulocity processing mode
// (uppercased), applied to requests when the --processingMode flag is not
// given. Empty when no default is configured.
func (c *Config) GetProcessingMode() string {
	return strings.ToUpper(c.viper.GetString(SettingsProcessingMode))
}

// StoreToken controls if the tokens are saved to the session file or not
func (c *Config) StoreToken() bool {
	return c.viper.GetBool(SettingsStorageStoreToken)
}

// StorePassword controls if the password is saved to the session file or not
func (c *Config) StorePassword() bool {
	return c.viper.GetBool(SettingsStorageStorePassword)
}

// GetTemplatePaths template folders where the template files are located
func (c *Config) GetTemplatePaths() []string {
	// Prefer custom path over default path
	paths := make([]string, 0)
	paths = append(paths, c.GetPathSlice(SettingsTemplateCustomPaths)...)
	paths = append(paths, c.GetPathSlice(SettingsTemplatePath)...)
	return paths
}

// SSODiscoveryUrl SSO discovery URL (OpenID Connect Configuration URL)
func (c *Config) SSODiscoveryUrl() string {
	return c.viper.GetString(SettingsSSODiscoveryUrl)
}

// GetCertificate returns the path to the PEM-encoded client certificate file for mTLS auth (C8Y_CERTIFICATE)
func (c *Config) GetCertificate() string {
	return c.viper.GetString("certificate")
}

// GetCertificateKey returns the path to the PEM-encoded private key file for mTLS auth (C8Y_CERTIFICATE_KEY)
func (c *Config) GetCertificateKey() string {
	if v := c.viper.GetString("certificate_key"); v != "" {
		return v
	}
	// Fallback for session files that store the key as camelCase certificateKey
	return c.viper.GetString("certificatekey")
}

// BrowserCallbackURL custom redirect URI for the browser flow local callback server
func (c *Config) BrowserCallbackURL() string {
	return c.viper.GetString(SettingsBrowserCallbackURL)
}

// SSOScopes scopes to use in the device code request when using SSO
func (c *Config) SSOScopes() []string {
	// Be flexible with the format, accept either a "," or " " separator
	// * "openid offline_access"
	// * "openid,offline_access"
	// * "openid,offline_access"
	// * "openid, offline_access"
	// * ["openid, "offline_access"]
	// * ["openid,offline_access"]
	rawValues := c.viper.GetStringSlice(SettingsSSOScopes)
	values := make([]string, 0, len(rawValues))
	for _, value := range rawValues {
		for _, item := range strings.FieldsFunc(value, func(r rune) bool {
			return r == ',' || r == ' '
		}) {
			values = append(values, strings.TrimSpace(item))
		}
	}
	return values
}

// SetSessionMode set the session mode (it is not persisted)
func (c *Config) SetSessionMode(mode SessionMode) {
	c.Set(SettingsMode, mode.String())
}

// SessionMode returns the current session mode which controls what the user can do
func (c *Config) SessionMode(defaultMode ...SessionMode) SessionMode {
	mode := SessionModeProduction
	if len(defaultMode) > 0 {
		mode = defaultMode[0]
	}
	return mode.FromString(c.viper.GetString(SettingsMode), c.IsCIMode())
}

func HasLegacySessionMode(v *viper.Viper) (SessionMode, bool) {
	// Prefer newer session value
	if v.IsSet("settings.session.mode") {
		return SessionModeUnset, false
	}

	// Check for legacy settings
	if !(v.IsSet("settings.mode.enablecreate") || v.IsSet("settings.mode.enableupdate") || v.IsSet("settings.mode.enabledelete")) {
		return SessionModeUnset, false
	}

	// Map legacy mode settings to session mode
	enableCreate := v.GetBool("settings.mode.enablecreate")
	enableUpdate := v.GetBool("settings.mode.enableupdate")
	enableDelete := v.GetBool("settings.mode.enabledelete")

	mode := SessionModeUnset
	if !enableCreate && !enableUpdate && !enableDelete {
		mode = SessionModeProduction
	} else if enableCreate && enableUpdate && !enableDelete {
		mode = SessionModeQual
	} else if enableCreate && enableUpdate && enableDelete {
		mode = SessionModeDev
	}
	return mode, true
}

// AllowModeCreate enables create (post) commands
func (c *Config) AllowModeCreate() bool {
	return c.SessionMode().CanCreate()
}

// AllowModeUpdate enables update commands
func (c *Config) AllowModeUpdate() bool {
	return c.SessionMode().CanUpdate()
}

// AllowModeDelete enables delete commands
func (c *Config) AllowModeDelete() bool {
	return c.SessionMode().CanDelete()
}

// Force don't prompt for confirmation
func (c *Config) Force() bool {
	return c.viper.GetBool(SettingsForce)
}

// ForceConfirm force prompt for confirmation
func (c *Config) ForceConfirm() bool {
	return c.viper.GetBool(SettingsForceConfirm)
}

// GetConfirmationMethods get HTTP methods that require confirmation
func (c *Config) GetConfirmationMethods() string {
	return c.viper.GetString(SettingsModeConfirmation)
}

// GetOutputFileRaw file path where the raw output file will be saved to
func (c *Config) GetOutputFileRaw() string {
	if c.outputFileRaw == nil {
		value := c.ExpandHomePath(c.viper.GetString(SettingsOutputFileRaw))
		c.outputFileRaw = &value
	}
	return *c.outputFileRaw
}

// GetOutputFileRaw file path where the parsed response will be saved to
func (c *Config) GetOutputFile() string {
	if c.outputFile == nil {
		value := c.ExpandHomePath(c.viper.GetString(SettingsOutputFile))
		c.outputFile = &value
	}
	return *c.outputFile
}

// GetOutputTemplate returns the output template to use when processing the output
func (c *Config) GetOutputTemplate() string {
	if c.outputTemplate == nil {
		contents := flags.ResolveTemplate(c.viper.GetString(SettingsOutputTemplate), c.templateResolver)
		c.outputTemplate = &contents
	}
	return *c.outputTemplate
}

// GetOutputFormat Get output format type, i.e. json, csv, table etc.
func (c *Config) GetOutputFormat() OutputFormat {
	if c.RawOutput() {
		return OutputJSON
	}
	format := c.viper.GetString(SettingsOutputFormat)
	outputFormat := OutputJSON.FromString(format)
	// c.Logger.Debugf("output format: %s", outputFormat.String())
	return outputFormat
}

// GetOutputFormat Get output format type, i.e. json, csv, table etc.
func (c *Config) GetOutputFormatWithDefault(cmd *cobra.Command, fallback OutputFormat) OutputFormat {
	if !cmd.Flags().Changed("output") {
		return fallback
	}
	value, err := cmd.Flags().GetString("output")
	if err != nil {
		return fallback
	}
	return fallback.FromString(value)
}

// IsCSVOutput check if csv output is enabled
func (c *Config) IsCSVOutput() bool {
	format := c.GetOutputFormat()
	return format == OutputCSV || format == OutputCSVWithHeader
}

func (c *Config) IsTSVOutput() bool {
	format := c.GetOutputFormat()
	return format == OutputTSV
}

func (c *Config) IsCompletionOutput() bool {
	format := c.GetOutputFormat()
	return format == OutputCompletion
}

// IsResponseOutput check if raw server response should be used
func (c *Config) IsResponseOutput() bool {
	return c.GetOutputFormat() == OutputServerResponse
}

// EncryptionEnabled enables encryption when storing sensitive session data
func (c *Config) EncryptionEnabled() bool {
	return c.viper.GetBool(SettingsEncryptionEnabled)
}

// Enable/Disable encryption
func (c *Config) SetEncryptionEnabled(v bool) {
	c.viper.Set(SettingsEncryptionEnabled, v)
}

// GetActivityLogPath path where the activity log will be stored
func (c *Config) GetActivityLogPath() string {
	return c.ExpandHomePath(c.viper.GetString(SettingsActivityLogPath))
}

// ActivityLogEnabled enables/disables the activity log
func (c *Config) ActivityLogEnabled() bool {
	return c.viper.GetBool(SettingsActivityLogEnabled)
}

// GetActivityLogMethodFilter filters the activity log entries by a space delimited methods, i.e. GET POST PUT
func (c *Config) GetActivityLogMethodFilter() string {
	return c.viper.GetString(SettingsActivityLogMethodFilter)
}

// HideSensitive hide sensitive information in log entries
func (c *Config) HideSensitive() bool {
	return c.viper.GetBool(SettingsLoggerHideSensitive)
}

// HideSessionBanner hide sensitive information in the session banner
func (c *Config) HideSessionBanner() bool {
	return c.viper.GetBool(SettingsSessionHide)
}

// DisableStdin hide sensitive information in log entries
func (c *Config) DisableStdin() bool {
	return c.viper.GetBool(SettingsDisableInput)
}

// Change the disable stdin value
func (c *Config) SetDisableStdin(v bool) {
	c.viper.Set(SettingsDisableInput, v)
}

// AllowEmptyPipe check if empty piped data is allowed
func (c *Config) AllowEmptyPipe() bool {
	return c.viper.GetBool(SettingsAllowEmptyPipe)
}

// GetConfigPath get global settings file path
func (c *Config) GetConfigPath() string {
	return c.ExpandHomePath(c.viper.GetString(SettingsConfigPath))
}

// GetViewPaths get list of view paths
func (c *Config) GetViewPaths() []string {
	paths := c.GetPathSlice(SettingsViewsCommonPaths)
	paths = append(paths, c.GetPathSlice(SettingsViewsCustomPaths)...)
	return paths
}

// GetJSONFilter get json filter to be applied to the output
func (c *Config) GetJSONFilter() []string {
	return c.viper.GetStringSlice(SettingsFilter)
}

// GetSilentStatusCodes Status codes which will not print out an error message
func (c *Config) GetSilentStatusCodes() string {
	return c.viper.GetString(SettingsSilentStatusCodes)
}

// GetSilentExit silent status codes don't affect the exit code
func (c *Config) GetSilentExit() bool {
	return c.viper.GetBool(SettingsSilentExit)
}

func ParseLoginTypeWithDefault(v string) string {
	// Pass v2-only types through unchanged – c8y.ParseLoginType doesn't know them.
	switch strings.ToUpper(v) {
	case logintype.Browser, logintype.Certificate, logintype.Device:
		return strings.ToUpper(v)
	}
	value, err := c8y.ParseLoginType(v)
	if err != nil {
		value = ""
	}
	return value
}

// GetLoginTypeWithDefault get the preferred login type
func (c *Config) GetLoginTypeWithDefault() string {
	v := c.GetLoginTypeRaw()
	return ParseLoginTypeWithDefault(v)
}

// GetLoginTypeRaw get the raw value, where it could also be an empty value
func (c *Config) GetLoginTypeRaw() string {
	if c.HasSessionUsernameOrPassword() {
		// Force BASIC AUTH
		return c8y.LoginTypeBasic
	}
	v := c.viper.GetString(SettingsLoginType)
	return strings.ToUpper(v)
}

func (c *Config) HasSessionUsernameOrPassword() bool {
	return c.GetSessionUsername() != "" || c.GetSessionPassword() != ""
}

func (c *Config) GetSessionUsername() string {
	return c.viper.GetString("settings.defaults.sessionUsername")
}

func (c *Config) GetSessionPassword() string {
	return c.viper.GetString("settings.defaults.sessionPassword")
}

// SetLoginType sets the authorization method, e.g. BASIC, OAUTH2_INTERNAL, NONE
func (c *Config) SetLoginType(v string) {
	// Pass v2-only types through unchanged.
	switch strings.ToUpper(v) {
	case logintype.Browser, logintype.Certificate, logintype.Device:
		c.Set(SettingsLoginType, strings.ToUpper(v))
		return
	}
	value, err := c8y.ParseLoginType(v)
	if err != nil {
		value = c8y.LoginTypeOAuth2Internal
	}
	c.Set(SettingsLoginType, value)
}

// CacheEnabled shows if caching is enabled or not
func (c *Config) CacheEnabled() bool {
	return c.viper.GetBool(SettingsDefaultsCacheEnabled) && !c.viper.GetBool(SettingsDefaultsNoCache)
}

// CacheTTL cache time-to-live. After the duration then the cache will no longer be used.
func (c *Config) CacheTTL() time.Duration {
	return c.getDuration(SettingsDefaultsCacheTTL)
}

// CacheDir get the cache directory
func (c *Config) CacheDir() string {
	return c.viper.GetString(SettingsCacheDir)
}

// CacheMethods HTTP methods which should be cached
func (c *Config) CacheMethods() string {
	return c.viper.GetString(SettingsCacheMethods)
}

// CacheMode caching mode which controls
func (c *Config) CacheMode() c8y.StoreMode {
	rawValue := c.viper.GetString(SettingsCacheMode)
	if strings.EqualFold(rawValue, "storeonly") {
		return c8y.StoreModeWrite
	}
	return c8y.StoreModeReadWrite
}

// CacheKeyIncludeHost include full host name in cache key generation
func (c *Config) CacheKeyIncludeHost() bool {
	return c.viper.GetBool(SettingsCacheKeyHost)
}

// CacheKeyIncludeAuth include authorization cache key generation
func (c *Config) CacheKeyIncludeAuth() bool {
	return c.viper.GetBool(SettingsCacheKeyAuth)
}

func (c *Config) CacheBodyKeys() []string {
	return c.viper.GetStringSlice(SettingsCacheBodyPaths)
}

// SkipSSLVerify skip SSL verify
func (c *Config) SkipSSLVerify() bool {
	return c.viper.GetBool(SettingsDefaultsInsecure)
}

// Browser get default web browser
func (c *Config) Browser() string {
	return c.viper.GetString(SettingsBrowser)
}

// Get Extension Data Directory
func (c *Config) ExtensionsDataDir() string {
	dir := c.viper.GetString(SettingsExtensionDataDir)
	if dir == "" {
		dir = c.GetSessionHomeDir()
	}
	return filepath.Join(dir, "extensions")
}

func (c *Config) DefaultHost() string {
	return c.viper.GetString(SettingsExtensionDefaultHost)
}

func (c *Config) ExtensionDefaultUsername() string {
	return c.viper.GetString(SettingsExtensionDefaultUsername)
}

func (c *Config) GetRemoteAccessDefaultSSHUser() string {
	return c.viper.GetString(SettingsRemoteAccessDefaultSSHUser)
}

// GetJSONSelect get json properties to be selected from the output. Only the given properties will be returned
func (c *Config) GetJSONSelect() []string {
	// Note: select is stored as an cobra Array String, which add special formatting of values.
	// so it needs to be converted to an array of strings
	values := c.viper.GetStringSlice(SettingsSelect)
	allitems := []string{}

	for _, item := range values {
		item = strings.Trim(item, "[]")
		item = strings.Trim(item, "\"")
		if item != "" {
			for v := range strings.SplitSeq(item, ",") {
				allitems = append(allitems, strings.TrimSpace(v))
			}
		}
	}

	// c.Logger.Debugf("json select: len=%d, values=%v", len(allitems), allitems)
	return allitems
}

func (c *Config) MustGetOutputCommonOptions(cmd *cobra.Command) *CommonCommandOptions {
	opts, err := c.GetOutputCommonOptions(cmd)
	if err != nil {
		panic(err)
	}
	return &opts
}

// GetOutputCommonOptions get common output options which controls how the output should be handled i.e. json filter, selects, csv etc.
func (c *Config) GetOutputCommonOptions(cmd *cobra.Command) (CommonCommandOptions, error) {
	if c.commonOptions != nil {
		return *c.commonOptions, nil
	}
	options := CommonCommandOptions{
		OutputFile:     c.GetOutputFile(),
		OutputFileRaw:  c.GetOutputFileRaw(),
		OutputTemplate: c.GetOutputTemplate(),
		WithError:      c.WithError(),
	}

	// Store flag values for usage in the output template
	commandFlags := make(map[string]string)
	cmd.Flags().Visit(func(f *pflag.Flag) {
		commandFlags[f.Name] = strings.Trim(f.Value.String(), "[]")
	})
	options.CommandFlags = commandFlags

	// default return property from the raw response
	options.ResultProperty = flags.GetCollectionPropertyFromAnnotation(cmd)

	// Filters and selectors
	filters := jsonfilter.NewJSONFilters(c.Logger)
	filters.AsCSV = c.IsCSVOutput()
	filters.AsTSV = c.IsTSVOutput()
	filters.AsCompletionFormat = c.IsCompletionOutput()
	filters.Flatten = c.FlattenJSON()
	filters.SkipHeaders = c.GetOutputFormat() != OutputTable
	filters.Pluck = c.GetJSONSelect()
	if err := filters.AddRawFilters(c.GetJSONFilter()); err != nil {
		return options, err
	}
	options.Filters = filters

	pageSize := c.GetPageSize()
	if pageSize > 0 && pageSize != c8ydefaults.PageSize {
		options.PageSize = pageSize
	}

	options.WithTotalPages = c.WithTotalPages()
	options.WithTotalElements = c.WithTotalElements()

	options.IncludeAll = c.IncludeAll()

	if options.IncludeAll {
		options.PageSize = c.GetIncludeAllPageSize()
		// c.Logger.Debugf("Setting pageSize to maximum value to limit number of requests. value=%d", options.PageSize)
	}

	options.CurrentPage = c.GetCurrentPage()
	options.TotalPages = c.GetTotalPages()

	options.ConfirmText = c.ConfirmText()
	if options.ConfirmText == "" {
		options.ConfirmText = cmd.Short
	}

	c.commonOptions = &options
	return options, nil
}

// AllSettings get all the settings as a map
func (c *Config) AllSettings() map[string]interface{} {
	return c.viper.AllSettings()
}

// MarshalSettings marshals all of the settings into json for debugging purposes
func (c *Config) MarshalSettings() ([]byte, error) {
	values := map[string]any{}
	if c.Persistent != nil {
		values["all"] = c.viper.AllSettings()
	}
	if c.Persistent != nil {
		values["persistent"] = c.Persistent.AllSettings()
	}
	return json.Marshal(values)
}

// SaveClientConfig save client settings to the session configuration
func (c *Config) SaveClientConfig(client *c8y.Client) error {
	if client != nil {
		if c.StorePassword() {
			c.SetPassword(client.Password)
		}

		if c.StoreToken() {
			c.SetToken(client.Token)
		}
		c.SetTenant(client.TenantName)

		if client.Version != "" {
			c.SetCumulocityVersion(client.Version)
		}

		if client.Username != "" {
			c.SetUsername(client.Username)
		}
	}
	return c.WritePersistentConfig()
}

func (c *Config) ShouldConfirm(methods ...string) bool {
	if c.ForceConfirm() {
		return true
	}

	useDryRun := c.ShouldUseDryRun("")
	if c.IsCIMode() || c.Force() || useDryRun {
		c.Logger.Debugf("no confirmation required. ci_mode=%v, force=%v, dry=%v", c.IsCIMode(), c.Force(), useDryRun)
		return false
	}

	if len(methods) == 0 {
		return true
	}

	confirmMethods := strings.ToUpper(c.GetConfirmationMethods())
	for _, method := range methods {
		if strings.Contains(confirmMethods, strings.ToUpper(method)) {
			c.Logger.Debugf("confirmation required due to method=%s", method)
			return true
		}
	}
	return false
}

const (
	NumberFormatNone   string = "none"
	NumberFormatMetric string = "metric"
)

// GetTableViewNumberFormatter get the number formatter to be used when rendering numbers
func (c *Config) GetTableViewNumberFormatter() numbers.NumberFormatter {
	formatName := strings.ToLower(c.viper.GetString(SettingsViewNumberFormat))
	if formatName == "" {
		formatName = NumberFormatNone
	}

	switch formatName {
	case NumberFormatMetric:
		return numbers.NewNumberViewOptions(
			c.viper.GetInt(SettingsViewNumbersMetricPrecision),
			c.viper.GetFloat64(SettingsViewNumbersMetricActivateRangeMin),
			c.viper.GetFloat64(SettingsViewNumbersMetricActivateRangeMax),
		)
	default:
		return &numbers.RawNumber{}
	}
}

// BindPFlag binds flags to the configuration
// Configuration precedence is:
// 1. Arguments
// 2. Environment variables
// 3. Session configuration
// 4. Global configuration
func (c *Config) BindPFlag(flags *pflag.FlagSet) error {
	var lastError error
	flags.VisitAll(func(f *pflag.Flag) {
		settingsName := GetSettingsName(f.Name)

		if err := c.viper.BindEnv(settingsName); err != nil {
			c.Logger.Warnf("Could not bind to environment variable. name=%s, err=%s", settingsName, err)
			lastError = err
		}

		if err := c.viper.BindPFlag(settingsName, flags.Lookup(f.Name)); err != nil {
			c.Logger.Warnf("Could not set flag. name=%s, err=%s", settingsName, err)
			lastError = err
		}
	})
	return lastError
}

// ExpandHomePath expand home path references found in the path
func (c *Config) ExpandHomePath(path string) string {
	expanded, err := homedir.Expand(path)
	if err != nil {
		if c.Logger != nil {
			c.Logger.Warnf("Could not expand path to home directory. %s", err)
		}
		expanded = path
	}
	// replace special variables
	expanded = strings.ReplaceAll(expanded, "$C8Y_HOME", c.GetHomeDir())
	expanded = strings.ReplaceAll(expanded, "$C8Y_SESSION_HOME", c.GetSessionHomeDir())
	return os.ExpandEnv(expanded)
}

// LogErrorF dynamically changes where the error is logged based on the users Silent Status Codes preferences
// Silent errors are only logged on the INFO level, where as non-silent errors are logged on the ERROR level
func (c *Config) LogErrorF(err error, format string, args ...interface{}) {
	errorLogger := c.Logger.Infof
	silentStatusCodes := c.GetSilentStatusCodes()
	if errors.Is(err, cmderrors.ErrNoMatchesFound) {
		if strings.Contains(silentStatusCodes, "404") {
			errorLogger = c.Logger.Infof
		}
	} else if cErr, ok := err.(cmderrors.CommandError); ok {

		// format errors as json messages
		// only log users errors
		if strings.Contains(silentStatusCodes, fmt.Sprintf("%d", cErr.StatusCode)) {
			errorLogger = c.Logger.Infof
		}
	}
	errorLogger(format, args...)
}

var ConfigExtensions = []string{"json", "yaml", "yml", "env", "toml", "properties"}

// SupportsFileExtension check if a filepath is using a supported extension or not
func SupportsFileExtension(p string) bool {
	ext := strings.TrimLeft(filepath.Ext(p), ".")
	for _, iExt := range ConfigExtensions {
		if strings.EqualFold(iExt, ext) {
			return true
		}
	}
	return false
}

func (c *Config) ClearSessionFile() {
	c.sessionFile = ""
}

func (c *Config) SetSessionFile(path string) {
	if _, fileErr := os.Stat(path); fileErr != nil {
		home := c.GetSessionHomeDir()
		c.Logger.Debugf("Resolving session %s in %s", path, home)
		matches, err := pathresolver.ResolvePaths([]string{home}, path, ConfigExtensions, "ignore")
		if err != nil {
			c.Logger.Warnf("Failed to find session. %s", err)
		}
		if len(matches) > 0 {
			path = matches[0]
			c.Logger.Debugf("Resolved session. %s", path)
		}
	}
	c.sessionFile = c.ExpandHomePath(path)
}

// GetSessionFile detect the session file path
func (c *Config) GetSessionFile(overrideSession ...string) string {
	var sessionFile string

	if len(overrideSession) > 0 {
		sessionFile = overrideSession[0]
	}

	if sessionFile == "" && c.sessionFile != "" {
		return c.sessionFile
	}

	if sessionFile == "" {
		sessionFile = c.viper.GetString(SettingsSessionFile)
	}

	if sessionFile == "" {
		// TODO: Create viper env alias rather than checking it manually
		sessionFile = os.Getenv("C8Y_SESSION")
	}

	sessionFile = strings.TrimPrefix(sessionFile, "file://")
	if _, fileErr := os.Stat(sessionFile); fileErr != nil {
		home := c.GetSessionHomeDir()
		c.Logger.Debugf("Resolving session %s in %s", sessionFile, home)
		matches, err := pathresolver.ResolvePaths([]string{home}, sessionFile, ConfigExtensions, "ignore")
		if err != nil {
			c.Logger.Warnf("Failed to find session. %s", err)
		}
		if len(matches) > 0 {
			sessionFile = matches[0]
			c.Logger.Debugf("Resolved session. %s", sessionFile)
		}
	}

	c.sessionFile = c.ExpandHomePath(sessionFile)
	return c.sessionFile
}

// ReadConfigFiles reads multiple configuration files to load the c8y session and other settings
//
// The session files are
// 1. load settings (from C8Y_SESSION_HOME path)
// 2. load session file (by path)
// 3. load session file (by name)
func (c *Config) ReadConfigFiles(client *c8y.Client, ignoreSessionFile ...bool) (path string, err error) {
	c.Logger.Debugf("Reading configuration files")
	v := c.viper
	v.AddConfigPath(".")
	v.AddConfigPath(c.GetHomeDir())

	// Load (non-session) preferences
	v.SetConfigName(SettingsGlobalName)

	if err := v.ReadInConfig(); err == nil {
		path = v.ConfigFileUsed()
		c.Logger.Infof("Loaded settings: %s", c.HideSensitiveInformationIfActive(client, path))
	}

	// Load session
	if len(ignoreSessionFile) == 0 || !ignoreSessionFile[0] {
		sessionFile := c.GetSessionFile("")

		if _, err := os.Stat(sessionFile); err == nil {
			// Load config by file path
			v.SetConfigFile(sessionFile)

			if err := c.ReadConfig(sessionFile); err != nil {
				c.Logger.Warnf("Could not read global settings file. file=%s, err=%s", sessionFile, err)
			}
		} else {
			// Load config by name
			sessionName := "session"
			if sessionFile != "" {
				sessionName = sessionFile
			}

			if sessionName != "" {
				v.SetConfigName(sessionName)
			}
		}
	}

	err = v.MergeInConfig()
	path = v.ConfigFileUsed()

	if err != nil {
		c.Logger.Debugf("Failed to merge config. %s", err)
	}

	return path, err
}

func (c *Config) HideSensitiveInformationIfActive(client *c8y.Client, message string) string {
	if !c.HideSensitive() {
		return message
	}
	return c.HideSensitiveInformation(client, message)
}

func (c *Config) HideSensitiveInformation(client *c8y.Client, message string) string {
	if client == nil {
		return message
	}

	username := os.Getenv("USERNAME")
	if username != "" {
		message = strings.ReplaceAll(message, username, "******")
	}

	if client.TenantName != "" {
		message = strings.ReplaceAll(message, client.TenantName, "{tenant}")
	}
	if client.Username != "" {
		message = strings.ReplaceAll(message, client.Username, "{username}")
	}
	if client.Password != "" {
		message = strings.ReplaceAll(message, client.Password, "{password}")
	}
	if client.Token != "" {
		message = strings.ReplaceAll(message, client.Token, "{token}")
	}
	if client.BaseURL != nil {
		if host := client.BaseURL.Host; host != "" {
			message = strings.ReplaceAll(message, strings.TrimRight(host, "/"), "{host}")
		}
	}

	basicAuthMatcher := regexp.MustCompile(`(Basic)\s+[A-Za-z0-9=]+`)
	message = basicAuthMatcher.ReplaceAllString(message, "$1 {base64 tenant/username:password}")

	return message
}

var PathSplitChar = ":"

// GetPathSlice get a slice of paths
func (c *Config) GetPathSlice(name string) (paths []string) {
	rawPaths := []string{}
	if v := c.viper.GetString(name); v != "" {
		rawPaths = append(rawPaths, strings.Split(v, PathSplitChar)...)
	} else if v := c.viper.GetStringSlice(name); len(v) > 0 {
		rawPaths = append(rawPaths, v...)
	}
	for _, p := range rawPaths {
		p = c.ExpandHomePath(p)
		if p != "" {
			paths = append(paths, p)
		}
	}
	return
}
