package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/pkg/prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CumulocitySessions struct {
	Sessions []CumulocitySession `json:"sessions"`
}

type Authentication struct {
	AuthType string         `json:"authType,omitempty"`
	Cookies  []*http.Cookie `json:"cookies,omitempty"`
}

// CumulocitySession contains all settings required to communicate with a Cumulocity service
type CumulocitySession struct {
	// ID          string `json:"id"`
	Host            string `json:"host"`
	Tenant          string `json:"tenant"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Description     string `json:"description"`
	UseTenantPrefix bool   `json:"useTenantPrefix"`

	// Authentication `json:"authentication,omitempty"`

	MicroserviceAliases map[string]string `json:"microserviceAliases,omitempty"`

	Index int    `json:"-"`
	Path  string `json:"-"`
	Name  string `json:"-"`
}

func WriteAuth(v *viper.Viper, savePassword, saveCookies bool) error {
	if savePassword {
		cliConfig.SetPassword(client.Password)
	}
	if saveCookies {
		cliConfig.SetAuthorizationCookies(client.Cookies)
	}
	cliConfig.SetTenant(client.TenantName)
	return cliConfig.WritePersistentConfig()
}

func NewCumulocitySessionFromFile(filePath string) (*CumulocitySession, error) {
	session := &CumulocitySession{}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	if session == nil {
		return nil, fmt.Errorf("Session marshalling failed")
	}

	session.Path = filePath

	basename := filepath.Base(filePath)
	extension := filepath.Ext(basename)
	session.Name = strings.TrimSuffix(basename, extension)
	return session, nil
}

func (s CumulocitySession) GetSessionPassphrase() string {
	return os.Getenv("C8Y_PASSPHRASE")
}

func (s *CumulocitySession) SetPassword(password string) {
	s.Password = password
}

func (s *CumulocitySession) SetHost(host string) {
	s.Host = formatHost(host)
}

func formatHost(host string) string {
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	return host
}

func (s CumulocitySession) GetHost() string {
	return formatHost(s.Host)
}

func (s CumulocitySession) GetPassword() string {
	pass, err := SecureDataAccessor.TryDecryptString(s.Password, s.GetSessionPassphrase())

	if err != nil {
		Logger.Errorf("Could not decrypt password. %s", err)
		return ""
	}

	return pass
}

type newSessionCmd struct {
	host           string
	username       string
	password       string
	description    string
	name           string
	tenant         string
	noTenantPrefix bool

	*baseCmd
}

func newNewSessionCmd() *newSessionCmd {
	ccmd := &newSessionCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Cumulocity session credentials",
		Long:  `Create a new Cumulocity session credentials`,
		Example: `
		c8y sessions create \
			--host "https://mytenant.eu-latest.cumulocity.com" \
			--username "myUser@me.com"
		
		// Create a new session and prompt for the password
		`,
		PersistentPreRunE: ccmd.promptArgs,
		RunE:              ccmd.newSession,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.host, "host", "", "Host. .e.g. test.cumulocity.com. (required)")
	cmd.Flags().StringVar(&ccmd.username, "username", "", "Username (without tenant). (required)")
	cmd.Flags().StringVar(&ccmd.password, "password", "", "Password. If left blank then you will be prompted for the password")
	cmd.Flags().StringVar(&ccmd.tenant, "tenant", "", "Tenant ID")
	cmd.Flags().StringVar(&ccmd.description, "description", "", "Description about the session")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of the session")
	// cmd.Flags().String("microserviceAliases", "", "Name of the session")
	cmd.Flags().BoolVar(&ccmd.noTenantPrefix, "noTenantPrefix", false, "Don't use tenant name as a prefix to the user name when using Basic Authentication. Defaults to false")

	// Required flags
	_ = cmd.MarkFlagRequired("host")
	// cmd.MarkFlagRequired("tenant")
	// cmd.MarkFlagRequired("username")
	// cmd.MarkFlagRequired("password")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newSessionCmd) promptArgs(cmd *cobra.Command, args []string) error {
	prompter := prompt.NewPrompt(Logger)

	if !cmd.Flags().Changed("username") {
		v, err := prompter.Username("Enter username", cliConfig.GetDefaultUsername())

		if err != nil {
			return err
		}
		n.username = v
	}

	if !cmd.Flags().Changed("password") {
		password, err := prompter.Password("Enter c8y password", "")
		if err != nil {
			return err
		}
		n.password = password
	}

	return nil
}

func (n *newSessionCmd) newSession(cmd *cobra.Command, args []string) error {

	session := &CumulocitySession{
		Host:            n.host,
		Tenant:          n.tenant,
		Username:        n.username,
		Description:     n.description,
		UseTenantPrefix: !n.noTenantPrefix,
	}
	session.MicroserviceAliases = make(map[string]string)

	session.SetPassword(n.password)

	// session name (default to host and username)
	hostname := "c8y"
	if u, err := url.Parse(session.GetHost()); err == nil {
		hostname = u.Host
	}

	sessionName := hostname + "-" + session.Username
	if v, err := cmd.Flags().GetString("name"); err == nil && v != "" {
		sessionName = v
	}

	outputDir := getSessionHomeDir()
	outputFile := n.formatFilename(sessionName)

	if err := n.writeSessionFile(outputDir, outputFile, *session); err != nil {
		return err
	}

	fmt.Println(path.Join(outputDir, outputFile))
	return nil
}

func (n *newSessionCmd) formatFilename(name string) string {
	if !strings.HasSuffix(name, ".json") {
		name = fmt.Sprintf("%s.json", name)
	}
	return name
}

func (n *newSessionCmd) writeSessionFile(outputDir, outputFile string, session CumulocitySession) error {
	data, err := json.MarshalIndent(session, "", "  ")

	if err != nil {
		return errors.Wrap(err, "failed to convert session to json")
	}

	outputPath := path.Join(outputDir, outputFile)

	if outputDir != "" {
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			Logger.Errorf("failed to create folder. folder=%s, err=%s", outputDir, err)
			return err
		}
	}
	Logger.Debugf("output file: %s", outputPath)

	if err := ioutil.WriteFile(path.Join(outputDir, outputFile), data, 0644); err != nil {
		return errors.Wrap(err, "failed to write to file")
	}
	return nil
}
