package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/reubenmiller/go-c8y-cli/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/pkg/prompt"
	"github.com/reubenmiller/go-c8y-cli/pkg/totp"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// PrefixEncrypted prefix used in encrypted string fields to identify when a string is encrypted or not
	PrefixEncrypted = "{encrypted}"

	// KeyFileName is the name of the reference encryption text
	KeyFileName = ".key"
)

const (
	EnvSettingsPrefix = "c8y"
)

const (
	// SettingsIncludeAllPageSize property name used to control the default page size when using includeAll parameter
	SettingsIncludeAllPageSize string = "settings.includeAll.pageSize"

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

	// SettingsDryRun dry run. Don't send any requests, just print out the information
	SettingsDryRun = "settings.defaults.dry"

	// SettingsDryRunFormat dry run output format. Controls how the dry run information is displayed
	SettingsDryRunFormat = "settings.defaults.dryFormat"

	// SettingsDebug Debug show debug messages
	SettingsDebug = "settings.defaults.debug"

	// SettingsJSONCompact show compact json output
	SettingsJSONCompact = "settings.defaults.compact"

	// ShowProgress show progress bar
	SettingsShowProgress = "settings.defaults.progess"

	// DisableColor don't print console output in color
	SettingsDisableColor = "settings.defaults.noColor"

	// SettingsProxy http/s proxy settings
	SettingsProxy = "settings.defaults.proxy"

	// IgnoreProxy ignore proxy settings
	SettingsIgnoreProxy = "settings.defaults.noProxy"

	// WithError return the error response on stdout rather than stderr
	SettingsWithError = "settings.defaults.withError"

	// SettingsWorkerDelay delay in milliseconds to wait after each request before the worker processes a new job (request)
	SettingsWorkerDelay = "settings.defaults.delay"

	// SettingsAbortOnErrorCount abort when the number of errors reaches this value
	SettingsAbortOnErrorCount = "settings.defaults.abortOnErrors"

	// ViewOption controls whether views are applied the output or not
	SettingsViewOption = "settings.defaults.view"

	// SettingsTimeout timeout in seconds use when sending requests
	SettingsTimeout = "settings.defaults.timeout"

	// SettingsConfirmText custom confirmation text to use to prompt the user of an action
	SettingsConfirmText = "settings.defaults.confirmText"

	// SettingsJSONFlatten flatten nested json using dot notation
	SettingsJSONFlatten = "settings.defaults.flatten"

	// SettingsStorageStoreCookies controls if the cookies are saved to the session file or not
	SettingsStorageStoreCookies = "settings.storage.storeCookies"

	// SettingsStorageStorePassword controls if the password is saved to the session file or not
	SettingsStorageStorePassword = "settings.storage.storePassword"

	// SettingsTemplatePath template folder where the template files are located
	SettingsTemplatePath = "settings.template.path"

	// SettingsModeEnableCreate enables create (post) commands
	SettingsModeEnableCreate = "settings.mode.enableCreate"

	// SettingsModeEnableUpdate enables update commands
	SettingsModeEnableUpdate = "settings.mode.enableUpdate"

	// SettingsModeEnableDelete enables delete commands
	SettingsModeEnableDelete = "settings.mode.enableDelete"

	// SettingsModeCI enable continuous integration mode (this will enable all commands)
	SettingsModeCI = "settings.ci"

	// Force don't prompt for confirmation
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
)

var (
	SettingsDefaultsPrefix = "settings.defaults"
)

const (
	ViewsNone = "none"
	ViewsAll  = "all"
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

	// Session file path
	path string

	// SecureData accessor to encrypt/decrypt data
	SecureData *encrypt.SecureData

	// Passphrase used for encrypting/decrypting fields
	Passphrase string

	// KeyFile path to the key file used to test encryption
	KeyFile string

	prompter prompt.Prompt

	// SecretText used to test the encryption passphrase
	SecretText string

	Logger *logger.Logger
}

// NewConfig returns a new CLI configuration object
func NewConfig(v *viper.Viper, home string, passphrase string) *Config {

	c := &Config{
		viper:      v,
		Passphrase: passphrase,
		SecureData: encrypt.NewSecureData("{encrypted}"),
		Persistent: viper.New(),
		prompter:   prompt.Prompt{},
		KeyFile:    path.Join(home, KeyFileName),
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
		WithBindEnv(SettingsMaxWorkers, 5),
		WithBindEnv(SettingsWorkers, 1),
		WithBindEnv(SettingsIncludeAllPageSize, 2000),
		WithBindEnv(SettingsStorageStorePassword, true),
		WithBindEnv(SettingsStorageStoreCookies, true),
		WithBindEnv(SettingsModeConfirmation, "PUT POST DELETE"),

		WithBindEnv(SettingsEncryptionEnabled, false),
		WithBindEnv(SettingsActivityLogEnabled, true),
		WithBindEnv(SettingsActivityLogPath, ""),
		WithBindEnv(SettingsActivityLogMethodFilter, "GET PUT POST DELETE"),
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
	c.viper.SetEnvPrefix("c8y")
	auth_variables := [...]string{
		"host",
		"username",
		"tenant",
		"password",
		"credential.totp.secret",
		"credential.cookies.0",
		"credential.cookies.1",
		"credential.cookies.2",
		"credential.cookies.3",
		"credential.cookies.4",
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

// SetSessionFilePath sets the session's file path
func (c *Config) SetSessionFilePath(v string) {
	c.path = v
}

// GetSessionFilePath returns the session file path
func (c *Config) GetSessionFilePath() string {
	return c.path
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
	if _, err := os.Stat(c.KeyFile); os.IsExist(err) {
		c.Logger.Infof("Key file already exists. file=%s", c.KeyFile)
		return nil
	}
	key, err := os.Create(c.KeyFile)
	if err != nil {
		return err
	}

	if _, err := key.WriteString(keyText); err != nil {
		return err
	}
	return nil
}

// ReadKeyFile reads the key file used as a reference to validate encryption (i.e. when no sessions exist)
func (c *Config) ReadKeyFile() error {

	// read from env variable
	if v := os.Getenv("C8Y_PASSPHRASE_TEXT"); v != "" && c.SecureData.IsEncrypted(v) == 1 {
		c.Logger.Infof("Using env variable 'C8Y_PASSPHRASE_TEXT' as example encryption text")
		c.SecretText = v
		return c.CreateKeyFile(v)
	}

	// read from file
	contents, err := ioutil.ReadFile(c.KeyFile)

	if err == nil {
		if c.SecureData.IsEncryptedBytes(contents) == 1 {
			c.SecretText = string(contents)
			return nil
		}
		c.Logger.Warningf("Key file is invalid or contains decrypted information")
	}

	// init key file
	passphrase, err := c.prompter.PasswordWithConfirm("new encryption passphrase", "Creating a encryption key for sessions")

	if err != nil {
		return err
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

// GetEnvironmentVariables gets all the environment variables associated with the current session
func (c Config) GetEnvironmentVariables(client *c8y.Client, setPassword bool) map[string]interface{} {
	host := c.GetHost()
	tenant := c.GetTenant()
	username := c.GetUsername()
	password := c.MustGetPassword()
	cookies := c.GetCookies()

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
		if len(client.Cookies) > 0 {
			cookies = client.Cookies
		}
	}

	// hide password if it is not needed
	if !setPassword && len(cookies) > 0 {
		password = ""
	}

	output := map[string]interface{}{
		"C8Y_URL":      host,
		"C8Y_BASEURL":  host,
		"C8Y_HOST":     host,
		"C8Y_TENANT":   tenant,
		"C8Y_USER":     username,
		"C8Y_USERNAME": username,
		"C8Y_PASSWORD": password,
	}

	if c.CachePassphraseVariables() {
		c.Logger.Info("Caching passphrase")
		if c.Passphrase != "" {
			output["C8Y_PASSPHRASE"] = c.Passphrase
		}
		if c.SecretText != "" {
			output["C8Y_PASSPHRASE_TEXT"] = c.SecretText
		}
	}

	// add decrypted cookies
	for i, cookie := range cookies {
		output[c.GetEnvKey(fmt.Sprintf("credential.cookies.%d", i))] = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
	}

	return output
}

// GetEnvKey returns the environment key value associated
func (c Config) GetEnvKey(key string) string {
	return "C8Y_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
}

// GetCookies gets the cookies stored in the configuration
func (c Config) GetCookies() []*http.Cookie {
	cookies := make([]*http.Cookie, 0)
	for i := 0; i < 5; i++ {
		cookieValue := c.viper.GetString(fmt.Sprintf("credential.cookies.%d", i))

		if cookieValue == "" {
			cookieValue = c.Persistent.GetString(fmt.Sprintf("credential.cookies.%d", i))
		}

		if cookieValue == "" {
			break
		}

		if v, err := c.DecryptString(cookieValue); err == nil {
			cookieValue = v
		} else {
			c.Logger.Warningf("Could not decrypt cookie. %s", err)
			continue
		}

		cookie := newCookie(cookieValue)
		if cookies != nil {
			cookies = append(cookies, cookie)
		}
	}
	return cookies
}

func newCookie(raw string) *http.Cookie {
	parts := strings.SplitN(raw, "=", 2)
	if len(parts) != 2 {
		return nil
	}

	valueParts := strings.SplitN(strings.TrimSpace(parts[1]), ";", 2)

	if len(valueParts) == 0 {
		return nil
	}

	return &http.Cookie{
		Name:  parts[0],
		Value: valueParts[0],
		Raw:   raw,
	}
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
	return c.Persistent.WriteConfig()
}

// SetAuthorizationCookies saves the authorization cookies
func (c *Config) SetAuthorizationCookies(cookies []*http.Cookie) {
	encryptedCookies := make(map[string]string)
	var err error

	for i, cookie := range cookies {
		c.Persistent.Set(fmt.Sprintf("credential.cookies.%d", i), "cookie")

		cookieData := cookie.Raw
		if c.IsEncryptionEnabled() {
			cookieData, err = c.SecureData.EncryptString(cookie.Raw, c.Passphrase)
			if err != nil {
				continue
			}
		}

		encryptedCookies[fmt.Sprintf("%d", i)] = cookieData
	}
	c.Persistent.Set("credential.cookies", encryptedCookies)
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
	return password == "" || c.SecureData.IsEncrypted(password) == 1
}

// MustGetPassword returns the decrypted password if there are no encryption errors, otherwise it will return an encrypted password
func (c *Config) MustGetPassword() string {
	decryptedValue, err := c.GetPassword()
	if err != nil {
		c.Logger.Warningf("Could not decrypt password. %s", err)
	}
	return decryptedValue
}

// SetPassword sets the password
func (c *Config) SetPassword(p string) {
	c.Persistent.Set("password", p)
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
	return c.viper.GetBool("settings.encryption.enabled")
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
	c.viper.SetDefault(name, defaultValue)
	return err
}

// DecryptSession decrypts a session (as long as the encryption passphrase has already been provided)
func (c *Config) DecryptSession() error {
	c.SetPassword(c.MustGetPassword())
	c.SetAuthorizationCookies(c.GetCookies())
	return c.WritePersistentConfig()
}

func (c *Config) Aliases() map[string]string {
	return c.viper.GetStringMapString("settings.aliases")
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
		c.Logger.Warningf("number of workers exceeds the maximum workers limit of %d. Use max instead", maxWorkers)
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

// DryRun dont sent any destructive requests. Just print out what would be sent
func (c *Config) DryRun() bool {
	return c.viper.GetBool(SettingsDryRun)
}

// SettingsDryRunFormat dry run output format. Controls how the dry run information is displayed
func (c *Config) DryRunFormat() string {
	return c.viper.GetString(SettingsDryRunFormat)
}

// Debug show debug messages
func (c *Config) Debug() bool {
	return c.viper.GetBool(SettingsDebug)
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
func (c *Config) WorkerDelay() int {
	return c.viper.GetInt(SettingsWorkerDelay)
}

// AbortOnErrorTotal abort when the number of errors reaches this value
func (c *Config) AbortOnErrorCount() int {
	return c.viper.GetInt(SettingsAbortOnErrorCount)
}

// ViewOption controls whether views are applied the output or not
func (c *Config) ViewOption() string {
	return c.viper.GetString(SettingsViewOption)
}

// RequestTimeout timeout in seconds use when sending requests
func (c *Config) RequestTimeout() float64 {
	return c.viper.GetFloat64(SettingsTimeout)
}

// FlattenJSON flatten nested json using dot notation
func (c *Config) FlattenJSON() bool {
	return c.viper.GetBool(SettingsJSONFlatten)
}

// ConfirmText custom confirmation text to use to prompt the user of an action
func (c *Config) ConfirmText() string {
	return c.viper.GetString(SettingsConfirmText)
}

// StoreCookies controls if the cookies are saved to the session file or not
func (c *Config) StoreCookies() bool {
	return c.viper.GetBool(SettingsStorageStoreCookies)
}

// StorePassword controls if the password is saved to the session file or not
func (c *Config) StorePassword() bool {
	return c.viper.GetBool(SettingsStorageStorePassword)
}

// GetTemplatePath template folder where the template files are located
func (c *Config) GetTemplatePath() string {
	return c.ExpandHomePath(c.viper.GetString(SettingsTemplatePath))
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

// GetOutputFormat Get output format
func (c *Config) GetOutputFormat() string {
	return c.viper.GetString(SettingsOutputFormat)
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

// GetConfigPath get global settings file path
func (c *Config) GetConfigPath() string {
	return c.ExpandHomePath(c.viper.GetString(SettingsConfigPath))
}

// GetViewPaths get list of view paths
func (c *Config) GetViewPaths() []string {
	viewPaths := c.viper.GetStringSlice(SettingsViewsCommonPaths)
	viewPaths = append(viewPaths, c.viper.GetStringSlice(SettingsViewsCustomPaths)...)
	for i, path := range viewPaths {
		viewPaths[i] = c.ExpandHomePath(path)
	}
	return viewPaths
}

// AllSettings get all the settings as a map
func (c *Config) AllSettings() map[string]interface{} {
	return c.viper.AllSettings()
}

// SaveClientConfig save client settings to the session configuration
func (c *Config) SaveClientConfig(client *c8y.Client) error {
	if c.StorePassword() {
		c.SetPassword(client.Password)
	}
	if c.StoreCookies() {
		c.SetAuthorizationCookies(client.Cookies)
	}
	c.SetTenant(client.TenantName)
	return c.WritePersistentConfig()
}

func (c *Config) ShouldConfirm(methods ...string) bool {
	if c.ForceConfirm() {
		return true
	}

	if c.IsCIMode() || c.Force() || c.DryRun() {
		c.Logger.Debugf("no confirmation required. ci_mode=%v, force=%v, dry=%v", c.IsCIMode(), c.Force(), c.DryRun())
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
		c.Logger.Warnf("Could not expand path to home directory. %s", err)
		return path
	}
	return expanded
}
