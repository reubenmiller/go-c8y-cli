package c8ysession

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/golang-jwt/jwt/v5"
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

	// authorized
	Authorized *bool `json:"authorized,omitempty"`

	// ID          string `json:"id"`
	Host            string `json:"host,omitempty"`
	Tenant          string `json:"tenant,omitempty"`
	Version         string `json:"version,omitempty"`
	Username        string `json:"username,omitempty"`
	Password        string `json:"password,omitempty"`
	Mode            string `json:"mode,omitempty"`
	TOTP            string `json:"totp,omitempty"`
	Token           string `json:"token,omitempty"`
	Description     string `json:"description,omitempty"`
	UseTenantPrefix bool   `json:"useTenantPrefix"`
	LoginType       string `json:"loginType,omitempty"`

	Settings *config.CommandSettings `json:"settings,omitempty"`

	MicroserviceAliases map[string]string `json:"microserviceAliases,omitempty"`

	Index     int    `json:"-"`
	Path      string `json:"-"`
	Extension string `json:"-"`
	Name      string `json:"-"`

	// How to identify the session
	SessionUri string `json:"sessionUri,omitempty"`

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

func (s *CumulocitySession) SetAuthorized(v bool) {
	s.Authorized = &v
}

func (s CumulocitySession) IsAuthorized() bool {
	return s.Authorized != nil && *s.Authorized
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
	maybeHideMessage := func(client *c8y.Client, message string) string {
		return message
	}
	hideInfo := cfg.HideSessionBanner()
	if hideInfo {
		maybeHideMessage = cfg.HideSensitiveInformation
	}

	if hideInfo {
		labelS.Fprintf(w, "---------------------  Cumulocity Session (sensitive info is hidden)  ---------------------\n")
	} else {
		labelS.Fprintf(w, "---------------------  Cumulocity Session  ---------------------\n")
	}
	// Always show the source of the session
	if session.SessionUri != "" {
		fmt.Fprintf(w, "\n    %s: %s\n\n\n", label("%s", "source"), header(session.SessionUri))
	} else {
		fmt.Fprintf(w, "\n    %s: %s\n\n\n", label("%s", "path"), header(session.Path))
	}

	if session.Mode != "" {
		fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "mode")), value(config.SessionModeProduction.FromString(session.Mode, false).Description()))
	}

	if session.Description != "" {
		fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "description")), value(maybeHideMessage(client, session.Host)))
	}

	fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "host")), value(maybeHideMessage(client, session.Host)))
	if session.Tenant != "" {
		fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "tenant")), value(maybeHideMessage(client, session.Tenant)))
	}
	if session.Version != "" {
		fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "version")), value(maybeHideMessage(client, session.Version)))
	}
	if session.Username != "" {
		fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "username")), value(maybeHideMessage(client, session.Username)))
	}
	if client != nil {
		fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "authType")), value(client.AuthorizationType.String()))
	}
	fmt.Fprintf(w, "\n")
}

func WriteOutput(w io.Writer, client *c8y.Client, cfg *config.Config, session *CumulocitySession, format string) error {

	shell, isShell := utilities.ShellType.Parse(utilities.ShellBash, format)
	if isShell {
		output := GetVariablesFromSession(session, cfg, client, cfg.AlwaysIncludePassword())
		utilities.WriteShellVariables(w, output, shell)
		return nil
	}

	if format == "" {
		return nil
	}

	switch format {
	case "json":
		out, err := MarshalSession(session, cfg)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "%s\n", out)
	case "env", "dotenv":
		output := GetVariablesFromSession(session, cfg, client, cfg.AlwaysIncludePassword())
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
func GetVariablesFromSession(session *CumulocitySession, cfg *config.Config, client *c8y.Client, setPassword bool) map[string]interface{} {
	host := session.Host
	domain := session.GetDomain()
	tenant := session.Tenant
	c8yVersion := client.Version
	username := session.Username
	password := session.Password
	mode := session.Mode
	token := session.Token
	authHeaderValue := ""
	authHeader := ""
	loginType := session.LoginType

	if dummyReq, err := client.NewRequest("GET", "/", "", nil); err == nil {
		authHeaderValue = dummyReq.Header.Get("Authorization")
		authHeader = "Authorization: " + authHeaderValue
	}

	// hide password if it is not needed
	if !setPassword && token != "" {
		password = ""
	}

	output := map[string]interface{}{
		"C8Y_SESSION":              "",
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
		"C8Y_SETTINGS_LOGIN_TYPE":  loginType,
	}

	if mode != "" {
		output[config.EnvSessionMode] = mode
	}

	cache := cfg.CachePassphraseVariables()
	cfg.Logger.Debugf("Cache passphrase: %v", cache)
	if cache {
		if cfg.Passphrase != "" {
			output[config.EnvPassphrase] = cfg.Passphrase
		}
		if cfg.SecretText != "" {
			output[config.EnvPassphraseText] = cfg.SecretText
		}
	}

	if client.AuthorizationType != c8y.AuthTypeBearer {
		output["C8Y_TOKEN"] = ""
	}

	// Favor older path style over sessionUri to help with backwards compatibility
	if session.Path != "" {
		output["C8Y_SESSION"] = session.Path
	} else if session.SessionUri != "" {
		output["C8Y_SESSION"] = session.SessionUri
	}
	return output
}

func ShowSessionEnvironmentVariables(session *CumulocitySession, cfg *config.Config, c8yclient *c8y.Client, shell utilities.ShellType) {
	output := GetVariablesFromSession(session, cfg, c8yclient, cfg.AlwaysIncludePassword())
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
		"C8Y_SETTINGS_LOGIN_TYPE",
		config.EnvSessionMode,
	}
	return keys
}

func MarshalSession(session *CumulocitySession, cfg *config.Config) ([]byte, error) {
	// Don't include password if a token is provided
	if !cfg.AlwaysIncludePassword() {
		if session.Token != "" {
			session.Password = ""
		}
	}
	return json.Marshal(session)
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

func shouldRenewToken(log *logger.Logger, t string, validFor time.Duration) (bool, *time.Time) {
	claims := jwt.RegisteredClaims{}
	parser := jwt.NewParser()
	_, _, err := parser.ParseUnverified(t, &claims)

	if err != nil {
		// Invalid token
		return true, nil
	}

	// Recently issued, so don't renew it
	// Check if the token's validity period is too short
	if claims.ExpiresAt != nil && claims.IssuedAt != nil {
		tokenValidityPeriod := claims.ExpiresAt.Sub(claims.IssuedAt.Time)
		if tokenValidityPeriod < validFor {
			log.Warnf("SSO token validity period is less than the given token validFor, so the token will be used regardless. minimumValidFor=%v, tokenValidity=%v", validFor, tokenValidityPeriod)
			return false, nil
		}
	}

	if claims.ExpiresAt != nil {
		limit := claims.ExpiresAt.Add(-1 * validFor)
		expiresSoon := limit.Before(time.Now())
		return expiresSoon, &claims.ExpiresAt.Time
	}
	return true, nil
}

// ShouldReuseToken checks if the token should be reused or not
func ShouldReuseToken(cfg *config.Config, log *logger.Logger, token string) bool {
	if token == "" {
		return false
	}
	reuse := true

	// Check if token is valid for the minimum period
	shouldBeValidFor := cfg.TokenValidFor()
	expiresSoon, expiresAt := shouldRenewToken(log, token, shouldBeValidFor)

	if expiresAt != nil {
		if time.Now().After(*expiresAt) {
			log.Infof("Token has expired. tokenExpiresAt=%s", expiresAt.Format(time.RFC3339))
			reuse = false
		} else if expiresSoon {
			log.Warnf("Ignoring existing token as it will expire soon. minimumValidFor=%s, tokenExpiresAt=%s", shouldBeValidFor, expiresAt.Format(time.RFC3339))
			reuse = false
		} else {
			log.Infof("Token expiresAt: %s", expiresAt.Format(time.RFC3339))
		}
	} else {
		log.Infof("Ignoring invalid token")
		reuse = false
	}
	return reuse
}
