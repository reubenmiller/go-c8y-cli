package c8ysession

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/utilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type CumulocitySessions struct {
	Sessions []CumulocitySession `json:"sessions"`
}

// CumulocitySession contains all settings required to communicate with a Cumulocity service
type CumulocitySession struct {
	Schema string `json:"$schema,omitempty"`

	// ID          string `json:"id"`
	Host            string `json:"host"`
	Tenant          string `json:"tenant"`
	Version         string `json:"version"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	TOTP            string `json:"totp"`
	Token           string `json:"token"`
	Description     string `json:"description"`
	UseTenantPrefix bool   `json:"useTenantPrefix"`

	Settings *config.CommandSettings `json:"settings,omitempty"`

	MicroserviceAliases map[string]string `json:"microserviceAliases,omitempty"`

	Index     int    `json:"-"`
	Path      string `json:"-"`
	Extension string `json:"-"`
	Name      string `json:"-"`

	// How to identify the session
	SessionUri string `json:"sessionUri"`

	Logger *logger.Logger `json:"-"`
	Config *config.Config `json:"-"`
}

func (s CumulocitySession) GetSessionPassphrase() string {
	return os.Getenv(config.EnvPassphrase)
}

func (s *CumulocitySession) SetPassword(password string) {
	s.Password = password
}

func (s *CumulocitySession) SetToken(token string) {
	s.Token = token
}

func (s *CumulocitySession) SetHost(host string) {
	s.Host = FormatHost(host)
}

func FormatHost(host string) string {
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	return host
}

func (s CumulocitySession) GetHost() string {
	return FormatHost(s.Host)
}

func (s CumulocitySession) GetPassword() string {
	pass, err := s.Config.SecureData.TryDecryptString(s.Password, s.GetSessionPassphrase())

	if err != nil {
		if s.Logger != nil {
			s.Logger.Errorf("Could not decrypt password. %s", err)
		}
		return ""
	}

	return pass
}

// GetDomain gets the custom Cumulocity domain for cases where it differs from the Host
func (s CumulocitySession) GetDomain() string {
	host := s.Host
	if !strings.Contains(host, "://") {
		host = "https://" + host
	}
	if domain, err := url.Parse(host); err == nil {
		return domain.Host
	}
	return s.Host
}

func PrintSessionInfoAsJSON(w io.Writer, client *c8y.Client, cfg *config.Config, session CumulocitySession) error {
	out, err := json.Marshal(session)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s\n", out)
	return nil
}

// PrintSessionInfo print out the session information to writer (i.e. console or file)
func PrintSessionInfo(w io.Writer, client *c8y.Client, cfg *config.Config, session CumulocitySession) {
	labelS := color.New(color.FgWhite, color.Faint)
	label := labelS.SprintfFunc()
	value := color.New(color.FgWhite).SprintFunc()
	header := color.New(color.FgCyan).SprintFunc()

	labelS.Fprintf(w, "---------------------  Cumulocity Session  ---------------------\n")
	if session.SessionUri != "" {
		fmt.Fprintf(w, "\n    %s: %s\n\n\n", label("%s", "source"), header(cfg.HideSensitiveInformationIfActive(client, session.SessionUri)))
	} else {
		fmt.Fprintf(w, "\n    %s: %s\n\n\n", label("%s", "path"), header(cfg.HideSensitiveInformationIfActive(client, session.Path)))
	}
	if session.Description != "" {
		fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "description")), value(cfg.HideSensitiveInformationIfActive(client, session.Host)))
	}

	fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "host")), value(cfg.HideSensitiveInformationIfActive(client, session.Host)))
	if session.Tenant != "" {
		fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "tenant")), value(cfg.HideSensitiveInformationIfActive(client, session.Tenant)))
	}
	if session.Version != "" {
		fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "version")), value(cfg.HideSensitiveInformationIfActive(client, session.Version)))
	}
	fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "username")), value(cfg.HideSensitiveInformationIfActive(client, session.Username)))
	fmt.Fprintf(w, "\n")
}

func WriteOutput(w io.Writer, client *c8y.Client, cfg *config.Config, session *CumulocitySession, format string) error {

	shell, isShell := utilities.ShellType.Parse(utilities.ShellBash, format)
	if isShell {
		output := GetVariablesFromSession(session, client, cfg.AlwaysIncludePassword())
		utilities.WriteShellVariables(w, output, shell)
		return nil
	}

	if format == "" {
		return nil
	}

	switch format {
	case "json":
		out, err := json.Marshal(session)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "%s\n", out)
	case "env", "dotenv":
		output := GetVariablesFromSession(session, client, cfg.AlwaysIncludePassword())
		for k, v := range output {
			if v != "" {
				fmt.Fprintf(w, "%s=%s\n", k, v)
			}
		}
	default:
		return fmt.Errorf("unsupported output format. %s", format)
	}
	return nil
}

// GetVariablesFromSession gets all the environment variables associated with the current session
func GetVariablesFromSession(session *CumulocitySession, client *c8y.Client, setPassword bool) map[string]interface{} {
	host := session.Host
	domain := session.GetDomain()
	tenant := session.Tenant
	c8yVersion := client.Version
	username := session.Username
	password := session.Password
	token := session.Token
	authHeaderValue := ""
	authHeader := ""

	if dummyReq, err := client.NewRequest("GET", "/", "", nil); err == nil {
		authHeaderValue = dummyReq.Header.Get("Authorization")
		authHeader = "Authorization: " + authHeaderValue
	}

	// hide password if it is not needed
	if !setPassword && token != "" {
		password = ""
	}

	output := map[string]interface{}{
		// "C8Y_SESSION":              c.GetSessionFile(),
		"C8Y_URL":                  host,
		"C8Y_BASEURL":              host,
		"C8Y_HOST":                 host,
		"C8Y_DOMAIN":               domain,
		"C8Y_TENANT":               tenant,
		"C8Y_VERSION":              c8yVersion,
		"C8Y_USER":                 username,
		"C8Y_TOKEN":                token,
		"C8Y_USERNAME":             username,
		"C8Y_PASSWORD":             password,
		"C8Y_HEADER_AUTHORIZATION": authHeaderValue,
		"C8Y_HEADER":               authHeader,
	}
	return output
}

func ShowClientEnvironmentVariables(cfg *config.Config, c8yclient *c8y.Client, shell utilities.ShellType) {
	output := cfg.GetEnvironmentVariables(c8yclient, cfg.AlwaysIncludePassword())
	utilities.WriteShellVariables(os.Stdout, output, shell)
}

func ShowSessionEnvironmentVariables(session *CumulocitySession, cfg *config.Config, c8yclient *c8y.Client, shell utilities.ShellType) {
	output := GetVariablesFromSession(session, c8yclient, cfg.AlwaysIncludePassword())
	utilities.WriteShellVariables(os.Stdout, output, shell)
}

func GetSessionEnvKeys() []string {
	keys := []string{
		"C8Y_HOST",
		"C8Y_URL",
		"C8Y_BASEURL",
		"C8Y_DOMAIN",
		"C8Y_TENANT",
		"C8Y_USER",
		"C8Y_USERNAME",
		"C8Y_PASSWORD",
		"C8Y_TOKEN",
		"C8Y_VERSION",
		"C8Y_SESSION",
		"C8Y_HEADER",
		"C8Y_HEADER_AUTHORIZATION",
		"C8Y_SETTINGS_MODE_ENABLECREATE",
		"C8Y_SETTINGS_MODE_ENABLEUPDATE",
		"C8Y_SETTINGS_MODE_ENABLEDELETE",
	}
	return keys
}

func ClearEnvironmentVariables(shell utilities.ShellType) {
	utilities.ClearEnvironmentVariables(GetSessionEnvKeys(), shell)
}

func ClearProcessEnvironment() {
	utilities.ClearProcessEnvironment(GetSessionEnvKeys())
}

func IsSessionFilePath(path string) bool {
	if path == "" {
		return false
	}
	path = strings.TrimPrefix(path, "file://")
	return !strings.Contains(path, "://")
}
