package extsdk

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ContextVersion is the newest runtime context version which this SDK supports
const ContextVersion = 1

// Environment variables set by c8y for extension commands
const (
	EnvContextVersion       = "C8Y_EXTENSION_CONTEXT_VERSION"
	EnvExtensionName        = "C8Y_EXTENSION_NAME"
	EnvExtensionDir         = "C8Y_EXTENSION_DIR"
	EnvBin                  = "C8Y_BIN"
	EnvHost                 = "C8Y_HOST"
	EnvTenant               = "C8Y_TENANT"
	EnvUser                 = "C8Y_USER"
	EnvToken                = "C8Y_TOKEN"
	EnvAuthorizationHeader  = "C8Y_HEADER_AUTHORIZATION"
	EnvSession              = "C8Y_SESSION"
	EnvSettingsMode         = "C8Y_SETTINGS_MODE"
	EnvSettingsDryRun       = "C8Y_SETTINGS_DEFAULTS_DRY"
	EnvSettingsInsecure     = "C8Y_SETTINGS_DEFAULTS_INSECURE"
	EnvNullInput            = "C8Y_EXTENSION_NULL_INPUT"
	EnvCertificate          = "C8Y_CERTIFICATE"
	EnvCertificateKey       = "C8Y_CERTIFICATE_KEY"
	EnvCumulocityVersion    = "C8Y_VERSION"
	errNotRunByC8yHelpLink  = "https://goc8ycli.netlify.app/docs/concepts/extensions/"
	errExpiredCredentialMsg = "check that the session is still valid, and log in again using 'c8y sessions login' or 'set-session'"
)

// ErrNotRunByC8y is returned by FromEnv when the binary was not started by c8y
var ErrNotRunByC8y = errors.New("the extension must be run through c8y, i.e. 'c8y <extension> <command>'")

// ErrUnsupportedContextVersion is returned by FromEnv when c8y provides a newer runtime context
var ErrUnsupportedContextVersion = errors.New("the c8y runtime context is newer than this extension supports. Update the extension")

// Context is the runtime context which c8y passes to an extension command
type Context struct {
	// Version of the runtime context
	Version int

	// Name of the extension (without the c8y- prefix)
	Name string

	// Dir is the folder where the extension is installed
	Dir string

	// Bin is the path to the c8y executable which started the extension
	Bin string

	// Session values
	Host                string
	Tenant              string
	Username            string
	Token               string
	AuthorizationHeader string
	SessionFile         string
	CumulocityVersion   string
	Certificate         string
	CertificateKey      string

	// Mode is the session mode, e.g. dev, qual or prod
	Mode string

	// Insecure is true if TLS certificates should not be verified
	Insecure bool

	dryRun    bool
	nullInput bool
	env       []string
}

// FromEnv reads the runtime context from the environment variables set by c8y
func FromEnv() (*Context, error) {
	return FromEnvironment(os.Environ())
}

// FromEnvironment reads the runtime context from the given environment ("KEY=value" items)
func FromEnvironment(environ []string) (*Context, error) {
	env := map[string]string{}
	for _, item := range environ {
		if key, value, ok := strings.Cut(item, "="); ok {
			env[key] = value
		}
	}

	rawVersion := env[EnvContextVersion]
	if rawVersion == "" {
		return nil, fmt.Errorf("%w. %s is not set. See %s", ErrNotRunByC8y, EnvContextVersion, errNotRunByC8yHelpLink)
	}
	version, err := strconv.Atoi(rawVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid %s value: %q", EnvContextVersion, rawVersion)
	}
	if version > ContextVersion {
		return nil, fmt.Errorf("%w. context_version=%d, supported=%d", ErrUnsupportedContextVersion, version, ContextVersion)
	}

	return &Context{
		Version:             version,
		Name:                env[EnvExtensionName],
		Dir:                 env[EnvExtensionDir],
		Bin:                 env[EnvBin],
		Host:                env[EnvHost],
		Tenant:              env[EnvTenant],
		Username:            env[EnvUser],
		Token:               env[EnvToken],
		AuthorizationHeader: env[EnvAuthorizationHeader],
		SessionFile:         env[EnvSession],
		CumulocityVersion:   env[EnvCumulocityVersion],
		Certificate:         env[EnvCertificate],
		CertificateKey:      env[EnvCertificateKey],
		Mode:                env[EnvSettingsMode],
		Insecure:            parseBool(env[EnvSettingsInsecure]),
		dryRun:              parseBool(env[EnvSettingsDryRun]),
		nullInput:           parseBool(env[EnvNullInput]),
		env:                 append([]string{}, environ...),
	}, nil
}

func parseBool(v string) bool {
	b, _ := strconv.ParseBool(strings.TrimSpace(v))
	return b
}

// DryRun returns true if c8y was run in dry run mode (--dry). Extensions should not
// change any data in dry run mode
func (c *Context) DryRun() bool {
	return c.dryRun
}

// NullInput returns true if the extension should not read stdin (--nullInput/-n), e.g.
// because stdin is not a terminal, but no input is piped to the command either
func (c *Context) NullInput() bool {
	return c.nullInput
}

// ExtensionFile returns the path to a file in the extension's folder, e.g. "templates/device.jsonnet"
func (c *Context) ExtensionFile(rel string) string {
	return filepath.Join(c.Dir, filepath.FromSlash(rel))
}

// HasSession returns true if c8y has an active session
func (c *Context) HasSession() bool {
	return c.Host != ""
}

// Environ returns the environment which the extension was started with
func (c *Context) Environ() []string {
	return append([]string{}, c.env...)
}
