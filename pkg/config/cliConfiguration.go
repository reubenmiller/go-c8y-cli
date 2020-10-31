package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/pkg/prompt"
	"github.com/spf13/viper"
)

var (
	// PrefixEncrypted prefix used in encrypted string fields to identify when a string is encrypted or not
	PrefixEncrypted = "{encrypted}"

	KeyFileName = ".key"
)

// CliConfiguration cli configuration settings
type CliConfiguration struct {
	viper *viper.Viper

	// Persistent settings (stored to file)
	Persistent *viper.Viper

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

// NewCliConfiguration returns a new CLI configuration object
func NewCliConfiguration(v *viper.Viper, secureData *encrypt.SecureData, home string, passphrase string) *CliConfiguration {
	c := &CliConfiguration{
		viper:      v,
		Passphrase: passphrase,
		SecureData: secureData,
		Persistent: viper.New(),
		prompter:   prompt.Prompt{},
		KeyFile:    path.Join(home, KeyFileName),
		Logger:     logger.NewDummyLogger("SecureData"),
	}
	c.prompter.Logger = c.Logger
	return c
}

// SetLogger sets the logger
func (c *CliConfiguration) SetLogger(l *logger.Logger) {
	c.Logger = l
	c.prompter.Logger = l
}

// ReadConfig reads the given file and loads it into the persistent session config
func (c *CliConfiguration) ReadConfig(file string) error {
	c.Persistent.SetConfigFile(file)
	return c.Persistent.ReadInConfig()
}

func (c *CliConfiguration) CheckEncryption(encryptedText ...string) (string, error) {
	secretText := c.SecretText
	if len(encryptedText) > 0 {
		secretText = encryptedText[0]
	}
	c.Logger.Infof("SecretText: %s", secretText)
	pass, err := c.prompter.EncryptionPassphrase(secretText, c.Passphrase)
	c.Passphrase = pass
	return pass, err
}

func (c *CliConfiguration) ReadKeyFile() error {

	// read from env variable
	if v := os.Getenv("C8Y_PASSPHRASE_TEXT"); v != "" {
		c.Logger.Debugf("Using env variable 'C8Y_PASSPHRASE_TEXT' as example encryption text")
		c.SecretText = v
		return nil
	}

	// read from file
	if contents, err := ioutil.ReadFile(c.KeyFile); err == nil {
		c.SecretText = string(contents)
		return nil
	}

	// init key file
	passphrase, err := c.prompter.PasswordWithConfirm("encryption")

	if err != nil {
		return err
	}

	c.Passphrase = passphrase

	keyText, err := c.SecureData.EncryptString("Cumulocity CLI Tool", c.Passphrase)

	if err != nil {
		return err
	}

	key, err := os.Create(c.KeyFile)
	if err != nil {
		return err
	}

	if _, err := key.WriteString(keyText); err != nil {
		return nil
	}

	c.SecretText = keyText
	return nil
}

// GetCookies gets the cookies stored in the configuration
func (c CliConfiguration) GetCookies() []*http.Cookie {
	cookies := make([]*http.Cookie, 0)

	for _, cookieValue := range c.viper.GetStringMapString("credential.cookies") {

		if v, err := c.SecureData.TryDecryptString(cookieValue, c.Passphrase); err == nil {
			cookieValue = v
		} else {
			c.Logger.Warningf("Could not decrypt cookie. %s", err)
			continue
		}

		parts := strings.SplitN(cookieValue, "=", 2)
		if len(parts) != 2 {
			continue
		}

		valueParts := strings.SplitN(strings.TrimSpace(parts[1]), ";", 2)

		if len(valueParts) == 0 {
			continue
		}

		cookie := &http.Cookie{
			Name:  parts[0],
			Value: valueParts[0],
			Raw:   cookieValue,
		}
		cookies = append(cookies, cookie)
	}

	return cookies
}

// DecryptString returns the decrypted string if the string is encrypted
func (c *CliConfiguration) DecryptString(value string) (string, error) {
	return c.SecureData.TryDecryptString(value, c.Passphrase)
}

// GetEncryptedString returns string value of a potentially encrypted field in the configuration
// If the fields starts with the encrypted prefix, then it will be decrypted using the CLI passphrase,
// otherwise the value will be returned as is.
func (c *CliConfiguration) GetEncryptedString(key string) string {
	value := c.viper.GetString(key)

	decryptedValue, err := c.DecryptString(value)
	if err != nil {
		return value
	}
	return decryptedValue
}

// SetEncryptedString encryptes and sets a value in the configuration
func (c *CliConfiguration) SetEncryptedString(key, value string) error {
	if value == "" {
		value = c.Persistent.GetString(key)
	}

	encryptedValue, err := c.SecureData.TryEncryptString(value, c.Passphrase)

	if err != nil {
		return err
	}

	c.Persistent.Set(key, encryptedValue)
	return nil
}

// WritePersistentConfig saves the configuration to file
func (c *CliConfiguration) WritePersistentConfig() error {
	file := c.viper.ConfigFileUsed()

	if file == "" {
		return fmt.Errorf("No config is being used")
	}
	c.Persistent.Set("$schema", "https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/master/tools/schema/session.schema.json")

	c.SetEncryptedString("password", "")
	return c.Persistent.WriteConfig()
}

// SetAuthorizationCookies saves the authorization cookies
func (c *CliConfiguration) SetAuthorizationCookies(cookies []*http.Cookie) {
	cookieValues := make([]string, 0)

	encryptedCookies := make(map[string]string)

	for i, cookie := range cookies {
		cookieValues = append(cookieValues, fmt.Sprintf("%s", cookie.Raw))
		c.Persistent.Set(fmt.Sprintf("credential.cookies.%d", i), "cookie")

		encryptedValue, err := c.SecureData.EncryptString(cookie.Raw, c.Passphrase)
		if err != nil {
			continue
		}
		encryptedCookies[fmt.Sprintf("%d", i)] = encryptedValue
	}
	c.Persistent.Set("credential.cookies", encryptedCookies)
}

// GetPassword returns the decrypted password of the current session
func (c *CliConfiguration) GetPassword() (string, error) {
	value := c.viper.GetString("password")

	decryptedValue, err := c.SecureData.TryDecryptString(value, c.Passphrase)
	if err != nil {
		return value, err
	}
	return decryptedValue, nil
}

func (c *CliConfiguration) SetPassword(p string) {
	c.Persistent.Set("password", p)
}

func (c *CliConfiguration) SetTenant(value string) {
	c.Persistent.Set("tenant", value)
}

// IsCIMode return true if the cli is running in CI mode
func (c *CliConfiguration) IsCIMode() bool {
	return c.viper.GetBool("settings.ci")
}

// GetString returns a string from the configuration
func (c *CliConfiguration) GetString(key string) string {
	return c.viper.GetString(key)
}

func (c *CliConfiguration) GetDefaultUsername() string {
	return c.viper.GetString("settings.default.username")
}
