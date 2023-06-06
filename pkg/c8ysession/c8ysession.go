package c8ysession

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
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
	Token           string `json:"token"`
	Description     string `json:"description"`
	UseTenantPrefix bool   `json:"useTenantPrefix"`

	Settings *config.CommandSettings `json:"settings,omitempty"`

	MicroserviceAliases map[string]string `json:"microserviceAliases,omitempty"`

	Index     int    `json:"-"`
	Path      string `json:"-"`
	Extension string `json:"-"`
	Name      string `json:"-"`

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

// PrintSessionInfo print out the session information to writer (i.e. console or file)
func PrintSessionInfo(w io.Writer, client *c8y.Client, cfg *config.Config, session CumulocitySession) {
	labelS := color.New(color.FgWhite, color.Faint)
	label := labelS.SprintfFunc()
	value := color.New(color.FgWhite).SprintFunc()
	header := color.New(color.FgCyan).SprintFunc()

	labelS.Fprintf(w, "---------------------  Cumulocity Session  ---------------------\n")
	fmt.Fprintf(w, "\n    %s: %s\n\n\n", label("%s", "path"), header(cfg.HideSensitiveInformationIfActive(client, session.Path)))
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
