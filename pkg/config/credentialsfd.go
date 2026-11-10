package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

// EnvCredentialsFD environment variable with the number of an inherited file
// descriptor to read the credentials from. It is used by the MCP server, so the
// credentials of the commands it runs are not visible in their environment, which
// other processes of the same user can read.
const EnvCredentialsFD = "C8Y_CREDENTIALS_FD"

// maxCredentialsSize limit for the credentials read from the file descriptor
const maxCredentialsSize = 64 * 1024

// Credentials secrets passed via EnvCredentialsFD
type Credentials struct {
	Token    string `json:"token,omitempty"`
	Password string `json:"password,omitempty"`
}

// ReadCredentialsFromFD reads the credentials (json) from the file descriptor given
// by EnvCredentialsFD, if set, and uses them instead of any other credentials.
// The variable is removed, so child processes don't try to read it.
func (c *Config) ReadCredentialsFromFD() error {
	value := os.Getenv(EnvCredentialsFD)
	if value == "" {
		return nil
	}
	_ = os.Unsetenv(EnvCredentialsFD)

	fd, err := strconv.Atoi(value)
	if err != nil || fd < 3 {
		return fmt.Errorf("invalid %s value %q", EnvCredentialsFD, value)
	}
	f := os.NewFile(uintptr(fd), "credentials")
	if f == nil {
		return fmt.Errorf("invalid %s file descriptor %d", EnvCredentialsFD, fd)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxCredentialsSize))
	if err != nil {
		return fmt.Errorf("could not read credentials from %s. %w", EnvCredentialsFD, err)
	}
	credentials := Credentials{}
	if err := json.Unmarshal(data, &credentials); err != nil {
		return fmt.Errorf("invalid credentials from %s. %w", EnvCredentialsFD, err)
	}
	if credentials.Token != "" {
		c.viper.Set(SettingsToken, credentials.Token)
	}
	if credentials.Password != "" {
		c.viper.Set("password", credentials.Password)
	}
	return nil
}
