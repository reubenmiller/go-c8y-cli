package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ydefaults"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonfilter"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/pkg/pathresolver"
	"github.com/reubenmiller/go-c8y-cli/pkg/prompt"
	"github.com/reubenmiller/go-c8y-cli/pkg/totp"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// EnvPassphrase passphrase environment variable name
	EnvPassphrase = "C8Y_PASSPHRASE"

	// EnvPassphraseText passphrase text environment variable name
	EnvPassphraseText = "C8Y_PASSPHRASE_TEXT"

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

	// SettingsIncludeAllDelayMS delay in milliseconds between retrieving the next page when using include all
	SettingsIncludeAllDelayMS = "settings.includeAll.delayMS"

	// SettingsPageSize page size
	SettingsPageSize = "settings.defaults.pageSize"

	// SettingsWithTotalPages include the total pages statistics under statistics.totalPages
	SettingsWithTotalPages = "settings.defaults.withTotalPages"

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

	// SettingsModeEnableCreate enables create (post) commands
	SettingsModeEnableCreate = "settings.mode.enableCreate"

	// SettingsModeEnableUpdate enables update commands
	SettingsModeEnableUpdate = "settings.mode.enableUpdate"

	// SettingsModeEnableDelete enables delete commands
	SettingsModeEnableDelete = "settings.mode.enableDelete"

	// SettingsModeCI enable continuous integration mode (this will enable all commands)
	SettingsModeCI = "settings.ci"

	// SettingsForce don't prompt for confirmation
	SettingsForce = "settings.defaults.force"

	// SettingsForceConfirm force prompt for confirmation
	SettingsForceConfirm = "settings.defaults.confirm"

	// SettingsModeConfirmation sets the confirm mode
	SettingsModeConfirmation = "settings.mode.confirmation"

	// GetOutputFileRaw file path where the raw response will be saved to
	SettingsOutputFileRaw = "settings.defaults.outputFileRaw"

	// SettingsOutputFile file path where the parsed response will be saved to
	SettingsOutputFile = "settings.defaults.outputFile"

	// SettingsOutputFormat Output format i.e. table, json, csv, csvheader
	SettingsOutputFormat = "settings.defaults.output"

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

	// SettingsViewsCustomPaths paths to custom fiew definition files
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

	// SettingsViewMaxColumnWidth maximum column width in characters
	SettingsViewMaxColumnWidth = "settings.views.columnMaxWidth"

	// SettingsViewColumnPadding column padding
	SettingsViewColumnPadding = "settings.views.columnPadding"

	// SettingsLoggerHideSensitive hide sensitive information in log entries
	SettingsLoggerHideSensitive = "settings.logger.hideSensitive"

	// SettingsDisableInput disable reading from stdin (pipeline input)
	SettingsDisableInput = "settings.defaults.nullInput"

	// SettingsAllowEmptyPipe allow empty piped data
	SettingsAllowEmptyPipe = "settings.defaults.allowEmptyPipe"

	// SettingsLoginType preferred login type, i.e. BASIC, OAUTH_INTERNAL etc.
	SettingsLoginType = "settings.login.type"

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

	// SettingsDefaultsInsecure allow insecure SSL connections
	SettingsDefaultsInsecure = "settings.defaults.insecure"
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
		WithBindEnv(SettingEncryptionCachePassphrase, true),
		WithBindEnv(SettingsMaxWorkers, 50),
		WithBindEnv(SettingsWorkers, 1),
		WithBindEnv(SettingsIncludeAllPageSize, 2000),
		WithBindEnv(SettingsStorageStorePassword, true),
		WithBindEnv(SettingsStorageStoreToken, true),
		WithBindEnv(SettingsModeConfirmation, "PUT POST DELETE"),

		WithBindEnv(SettingsEncryptionEnabled, false),
		WithBindEnv(SettingsActivityLogEnabled, true),
		WithBindEnv(SettingsActivityLogPath, path.Join(c.GetSessionHomeDir(), ActivityLogDirName)),
		WithBindEnv(SettingsActivityLogMethodFilter, "GET PUT POST DELETE"),

		// Dry run options
		WithBindEnv(SettingsDryRunPattern, ""),

		WithBindEnv(SettingsIncludeAllDelayMS, 50),
		WithBindEnv(SettingsTemplatePath, ""),
		WithBindEnv(SettingsModeEnableCreate, false),
		WithBindEnv(SettingsModeEnableUpdate, false),
		WithBindEnv(SettingsModeEnableDelete, false),
		WithBindEnv(SettingsModeCI, false),
		WithBindEnv(SettingsConfigPath, ""),
		WithBindEnv(SettingsViewsCommonPaths, ""),
		WithBindEnv(SettingsViewsCustomPaths, ""),

		WithBindEnv(SettingsViewMinColumnWidth, 2),
		WithBindEnv(SettingsViewMaxColumnWidth, 80),
		WithBindEnv(SettingsViewColumnPadding, 15),

		WithBindEnv(SettingsLoggerHideSensitive, false),

		WithBindEnv(SettingsCacheMethods, "GET"),
		WithBindEnv(SettingsCacheKeyHost, true),
		WithBindEnv(SettingsCacheKeyAuth, true),
		WithBindEnv(SettingsCacheMode, nil),
		WithBindEnv(SettingsCacheDir, filepath.Join(os.TempDir(), "go-c8y-cli-cache")),
	)

	if err != nil {
		c.Logger.Warnf("Could not bind settings. %s", err)
	}
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

// CheckEncryption checks if the usuer has provided the correct encryption password or not by testing the decryption of the secret text
func (c *Config) CheckEncryption(encryptedText ...string) (string, error) {
	secretText := c.SecretText
	if len(encryptedText) > 0 {
		secretText = encryptedText[0]
	}

	c.Logger.Infof("Checking encryption passphrase against secret text: %s", secretText)
	pass, err := c.prompter.EncryptionPassphrase(secretText, c.Passphrase, "")
	c.Passphrase = pass
	return pass, err
}

// BindAuthorization binds environment variables related to the authrorization to the configuration
func (c *Config) BindAuthorization() error {
	c.viper.SetEnvPrefix(EnvSettingsPrefix)
	auth_variables := [...]string{
		"host",
		"username",
		"tenant",
		"password",
		"token",
		"credential.totp.secret",
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
	return c.viper.GetString("host")
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
	contents, err := ioutil.ReadFile(c.KeyFile())

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

// GetEnvironmentVariables gets all the environment variables associated with the current session
func (c Config) GetEnvironmentVariables(client *c8y.Client, setPassword bool) map[string]interface{} {
	host := c.GetHost()
	tenant := c.GetTenant()
	username := c.GetUsername()
	password := c.MustGetPassword()
	token := c.MustGetToken()
	authHeaderValue := ""
	authHeader := ""

	if client != nil {
		if client.TenantName != "" {
			tenant = client.TenantName
		}
		if client.Username != "" {
			username = client.Username
		}
		if client.BaseURL.Host != "" {
			host = client.BaseURL.Scheme + "://" + client.BaseURL.Host
		}
		if client.Password != "" {
			password = client.Password
		}
		if client.Token != "" {
			token = client.Token
		}
		if dummyReq, err := client.NewRequest("GET", "/", "", nil); err == nil {
			authHeaderValue = dummyReq.Header.Get("Authorization")
			authHeader = "Authorization: " + authHeaderValue
		}
	}

	// hide password if it is not needed
	if !setPassword && token != "" {
		password = ""
	}

	output := map[string]interface{}{
		"C8Y_SESSION":              c.GetSessionFile(),
		"C8Y_URL":                  host,
		"C8Y_BASEURL":              host,
		"C8Y_HOST":                 host,
		"C8Y_TENANT":               tenant,
		"C8Y_USER":                 username,
		"C8Y_TOKEN":                token,
		"C8Y_USERNAME":             username,
		"C8Y_PASSWORD":             password,
		"C8Y_HEADER_AUTHORIZATION": authHeaderValue,
		"C8Y_HEADER":               authHeader,
	}

	cache := c.CachePassphraseVariables()
	c.Logger.Debugf("Cache passphrase: %v", cache)
	if cache {
		if c.Passphrase != "" {
			output[EnvPassphrase] = c.Passphrase
		}
		if c.SecretText != "" {
			output[EnvPassphraseText] = c.SecretText
		}
	}
	return output
}

// GetEnvKey returns the environment key value associated
func (c Config) GetEnvKey(key string) string {
	return "C8Y_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
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
	c.Persistent.Set("$schema", "https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/master/tools/schema/session.schema.json")

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
	value := c.viper.GetString("password")

	if value == "" {
		value = c.Persistent.GetString("password")
	}

	decryptedValue, err := c.DecryptString(value)
	if err != nil {
		return value, err
	}
	return decryptedValue, nil
}

// IsPasswordEncrypted return true if the password is encrypted
// If the password is empty then treat it as encrypted
func (c *Config) IsPasswordEncrypted() bool {
	password := c.viper.GetString("password")
	// return password != "" && c.SecureData.IsEncrypted(password) == 1
	return password == "" || c.SecureData.IsEncrypted(password) == 1
}

func (c *Config) IsTokenEncrypted() bool {
	token := c.viper.GetString("token")
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
func (c *Config) MustGetToken() string {
	decryptedValue, err := c.GetToken()
	if err != nil {
		c.Logger.Warningf("Could not decrypt token. %s", err)
	}
	return decryptedValue
}

// SetPassword sets the password
func (c *Config) SetPassword(p string) {
	c.Persistent.Set("password", p)
}

// SetToken sets the token used for OAUTH authentication
func (c *Config) SetToken(p string) {
	c.Persistent.Set(SettingsToken, p)
}

// SetTenant sets the tenant name
func (c *Config) SetTenant(value string) {
	c.Persistent.Set("tenant", value)
}

// IsCIMode return true if the cli is running in CI mode
func (c *Config) IsCIMode() bool {
	return c.viper.GetBool("settings.ci")
}

// IsEncryptionEnabled indicates if session encryption is enabled or not
func (c *Config) IsEncryptionEnabled() bool {
	return c.viper.GetBool(SettingsEncryptionEnabled)
}

// GetString returns a string from the configuration
func (c *Config) GetString(key string) string {
	return c.viper.GetString(key)
}

// GetDefaultUsername returns the default username
func (c *Config) GetDefaultUsername() string {
	return c.viper.GetString("settings.session.defaultUsername")
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
	c.SetToken(c.MustGetToken())
	return c.WritePersistentConfig()
}

// CommonAliases Get common aliases from the global configuration file
func (c *Config) CommonAliases() map[string]string {
	return c.viper.GetStringMapString(SettingsCommonAliases)
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

// GetIncludeAllDelay include all delay in milliseconds
func (c *Config) GetIncludeAllDelay() int64 {
	return c.viper.GetInt64(SettingsIncludeAllDelayMS)
}

// WithTotalPages return all available results
func (c *Config) WithTotalPages() bool {
	return c.viper.GetBool(SettingsWithTotalPages)
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

// DryRun dont sent any destructive requests. Just print out what would be sent
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
	return c.viper.GetBool(SettingsShowProgress)
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
	return c.viper.GetString(SettingsViewOption)
}

// ViewColumnMinWidth minimum column width in characters
func (c *Config) ViewColumnMinWidth() int {
	return c.viper.GetInt(SettingsViewMinColumnWidth)
}

// ViewColumnMinWidth maximum column width in characters
func (c *Config) ViewColumnMaxWidth() int {
	return c.viper.GetInt(SettingsViewMaxColumnWidth)
}

// ViewColumnPadding column padding
func (c *Config) ViewColumnPadding() int {
	return c.viper.GetInt(SettingsViewColumnPadding)
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

// StoreToken controls if the tokens are saved to the session file or not
func (c *Config) StoreToken() bool {
	return c.viper.GetBool(SettingsStorageStoreToken)
}

// StorePassword controls if the password is saved to the session file or not
func (c *Config) StorePassword() bool {
	return c.viper.GetBool(SettingsStorageStorePassword)
}

// GetTemplatePaths template folders where the template files are located
func (c *Config) GetTemplatePaths() (paths []string) {
	// Prefer custom path over default path
	paths = append(paths, c.GetPathSlice(SettingsTemplateCustomPaths)...)
	paths = append(paths, c.GetPathSlice(SettingsTemplatePath)...)
	return paths
}

// AllowModeCreate enables create (post) commands
func (c *Config) AllowModeCreate() bool {
	return c.viper.GetBool(SettingsModeEnableCreate) || c.CIModeEnabled()
}

// AllowModeUpdate enables update commands
func (c *Config) AllowModeUpdate() bool {
	return c.viper.GetBool(SettingsModeEnableUpdate) || c.CIModeEnabled()
}

// AllowModeDelete enables delete commands
func (c *Config) AllowModeDelete() bool {
	return c.viper.GetBool(SettingsModeEnableDelete) || c.CIModeEnabled()
}

// CIModeEnabled enable continuous integration mode (this will enable all commands)
func (c *Config) CIModeEnabled() bool {
	return c.viper.GetBool(SettingsModeCI)
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
	return c.ExpandHomePath(c.viper.GetString(SettingsOutputFileRaw))
}

// GetOutputFileRaw file path where the parsed response will be saved to
func (c *Config) GetOutputFile() string {
	return c.ExpandHomePath(c.viper.GetString(SettingsOutputFile))
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

// IsCSVOutput check if csv output is enabled
func (c *Config) IsCSVOutput() bool {
	format := c.GetOutputFormat()
	return format == OutputCSV || format == OutputCSVWithHeader
}

// IsResponseOutput check if raw server response should be used
func (c *Config) IsResponseOutput() bool {
	return c.GetOutputFormat() == OutputServerResponse
}

// EncryptionEnabled enables encryption when storing sensitive session data
func (c *Config) EncryptionEnabled() bool {
	return c.viper.GetBool(SettingsEncryptionEnabled)
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

// DisableStdin hide sensitive information in log entries
func (c *Config) DisableStdin() bool {
	return c.viper.GetBool(SettingsDisableInput)
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
	viewPaths := c.GetPathSlice(SettingsViewsCommonPaths)
	viewPaths = append(viewPaths, c.GetPathSlice(SettingsViewsCustomPaths)...)
	return viewPaths
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

// GetLoginType get the preferred login type
func (c *Config) GetLoginType() string {
	return c.viper.GetString(SettingsLoginType)
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

// SkipSSLVerify skip SSL verify
func (c *Config) SkipSSLVerify() bool {
	return c.viper.GetBool(SettingsDefaultsInsecure)
}

// GetJSONSelect get json properties to be selected from the output. Only the given properties will be returned
func (c *Config) GetJSONSelect() []string {
	// Note: select is stored as an cobra Array String, which add special formating of values.
	// so it needs to be converted to an array of strings
	values := c.viper.GetStringSlice(SettingsSelect)
	allitems := []string{}

	for _, item := range values {
		item = strings.Trim(item, "[]")
		item = strings.Trim(item, "\"")
		if item != "" {
			allitems = append(allitems, item)
		}
	}

	// c.Logger.Debugf("json select: len=%d, values=%v", len(allitems), allitems)
	return allitems
}

// GetOutputCommonOptions get common output options which controls how the output should be handled i.e. json filter, selects, csv etc.
func (c *Config) GetOutputCommonOptions(cmd *cobra.Command) (CommonCommandOptions, error) {
	options := CommonCommandOptions{
		OutputFile:    c.GetOutputFile(),
		OutputFileRaw: c.GetOutputFileRaw(),
	}

	// default return property from the raw response
	options.ResultProperty = flags.GetCollectionPropertyFromAnnotation(cmd)

	// Filters and selectors
	filters := jsonfilter.NewJSONFilters(c.Logger)
	filters.AsCSV = c.IsCSVOutput()
	filters.Flatten = c.FlattenJSON()
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

	return options, nil
}

// AllSettings get all the settings as a map
func (c *Config) AllSettings() map[string]interface{} {
	return c.viper.AllSettings()
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

// BindPFlag binds flags to the configuration
// Configuration precendence is:
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
func (c *Config) ReadConfigFiles(client *c8y.Client) (path string, err error) {
	c.Logger.Debugf("Reading configuration files")
	v := c.viper
	v.AddConfigPath(".")
	v.AddConfigPath(c.GetHomeDir())

	sessionFile := c.GetSessionFile("")

	// Load (non-session) preferences
	v.SetConfigName(SettingsGlobalName)

	if err := v.ReadInConfig(); err == nil {
		path = v.ConfigFileUsed()
		c.Logger.Infof("Loaded settings: %s", c.HideSensitiveInformationIfActive(client, path))
	}

	// Load session
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

	err = v.MergeInConfig()
	path = v.ConfigFileUsed()

	if err != nil {
		c.Logger.Debugf("Failed to merge config. %s", err)
	}

	return path, err
}

func (c *Config) HideSensitiveInformationIfActive(client *c8y.Client, message string) string {
	if client == nil {
		return message
	}

	if !c.HideSensitive() {
		return message
	}

	username := os.Getenv("USERNAME")
	if username != "" {
		message = strings.ReplaceAll(message, username, "******")
	}

	if client != nil {
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
			message = strings.ReplaceAll(message, strings.TrimRight(client.BaseURL.Host, "/"), "{host}")
		}
	}

	basicAuthMatcher := regexp.MustCompile(`(Basic\s+)[A-Za-z0-9=]+`)
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
