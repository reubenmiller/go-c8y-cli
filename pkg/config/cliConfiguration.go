package config

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/encrypt"
	"github.com/spf13/viper"
)

// PrefixEncrypted prefix used in encrypted string fields to identify when a string is encrypted or not
var PrefixEncrypted = "{encrypted}"

// CliConfiguration cli configuration settings
type CliConfiguration struct {
	viper *viper.Viper

	// Persistent settings (stored to file)
	Persistent *viper.Viper

	// Passphrase used for encrypting/decrypting fields
	Passphrase string
}

// NewCliConfiguration returns a new CLI configuration object
func NewCliConfiguration(v *viper.Viper, passphrase string) *CliConfiguration {
	return &CliConfiguration{
		viper:      v,
		Passphrase: passphrase,
		Persistent: viper.New(),
	}
}

// ReadConfig reads the given file and loads it into the persistent session config
func (c *CliConfiguration) ReadConfig(file string) error {
	c.Persistent.SetConfigFile(file)
	return c.Persistent.ReadInConfig()
}

// GetCookies gets the cookies stored in the configuration
func (c CliConfiguration) GetCookies() []*http.Cookie {
	cookies := make([]*http.Cookie, 0)
	for _, cookieValue := range c.viper.GetStringSlice("credential.cookies") {
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
	if strings.HasPrefix(value, PrefixEncrypted) {
		hexVal, err := hex.DecodeString(strings.TrimPrefix(value, PrefixEncrypted))

		if err != nil {
			return "", err
		}
		b, err := encrypt.Decrypt(hexVal, c.Passphrase)

		if err != nil {
			return "", err
		}
		value = string(b)
	}
	return value, nil
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
		value = c.viper.GetString(key)
	}

	if !strings.HasPrefix(value, PrefixEncrypted) {
		value = PrefixEncrypted + fmt.Sprintf("%x", encrypt.EncryptString(value, c.Passphrase))
	}

	c.Persistent.Set(key, value)
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

	for i, cookie := range cookies {
		cookieValues = append(cookieValues, fmt.Sprintf("%s", cookie.Raw))
		c.Persistent.Set(fmt.Sprintf("credential.cookies.%d", i), "cookie")
	}
	c.Persistent.Set("credential.cookies", cookieValues)
}
