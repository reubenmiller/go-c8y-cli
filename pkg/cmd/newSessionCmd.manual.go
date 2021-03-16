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

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
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
	Schema string `json:"$schema,omitempty"`

	// ID          string `json:"id"`
	Host            string `json:"host"`
	Tenant          string `json:"tenant"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Description     string `json:"description"`
	UseTenantPrefix bool   `json:"useTenantPrefix"`

	Settings *CommandSettings `json:"settings,omitempty"`

	MicroserviceAliases map[string]string `json:"microserviceAliases,omitempty"`

	Index int    `json:"-"`
	Path  string `json:"-"`
	Name  string `json:"-"`
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
	pass, err := cliConfig.SecureData.TryDecryptString(s.Password, s.GetSessionPassphrase())

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
	sessionType    string
	noTenantPrefix bool
	noStorage      bool
	encrypt        bool

	*baseCmd
}

func newNewSessionCmd() *newSessionCmd {
	ccmd := &newSessionCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create session",
		Long:  `Create a new Cumulocity session`,
		Example: heredoc.Doc(`
### Example 1: Create a DEV new session. Prompt for username and password

$ c8y sessions create --type dev --host "https://mytenant.eu-latest.cumulocity.com"

### Example 2: Create a new QA (QUAL) session prompting for password

$ c8y sessions create \
    --type qual \
	--host "https://mytenant.eu-latest.cumulocity.com"
	--username "myUser@me.com"

### Example 3: Create a new production session where only only GET commands are enabled (with no password storage)

$ c8y sessions create --type prod --host "https://mytenant.eu-latest.cumulocity.com" --noStorage
		`),
		PersistentPreRunE: ccmd.promptArgs,
		Args:              cobra.NoArgs,
		RunE:              ccmd.newSession,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.host, "host", "", "Host. .e.g. test.cumulocity.com. (required)")
	cmd.Flags().StringVar(&ccmd.username, "username", "", "Username (without tenant). (required)")
	cmd.Flags().StringVar(&ccmd.password, "password", "", "Password. If left blank then you will be prompted for the password")
	cmd.Flags().StringVar(&ccmd.tenant, "tenant", "", "Tenant ID")
	cmd.Flags().StringVar(&ccmd.description, "description", "", "Description about the session")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of the session")
	cmd.Flags().StringVar(&ccmd.sessionType, "type", "", "Session type. List of predefined session types")
	cmd.Flags().BoolVar(&ccmd.noTenantPrefix, "noTenantPrefix", false, "Don't use tenant name as a prefix to the user name when using Basic Authentication. Defaults to false")
	cmd.Flags().BoolVar(&ccmd.noStorage, "noStorage", false, "Don't store any passwords or cookies in the session file")
	cmd.Flags().BoolVar(&ccmd.encrypt, "encrypt", false, "Encrypt passwords and cookies (occurs when logging in)")

	// Required flags
	_ = cmd.MarkFlagRequired("host")
	completion.WithOptions(cmd,
		completion.WithLazyRequired("type"),
		completion.WithValidateSet(
			"type",
			"prod\tProduction mode (read only)",
			"qual\tQA mode (delete disabled)",
			"dev\tDevelopment mode (no restrictions)",
		),
	)

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newSessionCmd) promptArgs(cmd *cobra.Command, args []string) error {
	prompter := prompt.NewPrompt(Logger)

	// read config
	if _, err := ReadConfigFiles(viper.GetViper()); err != nil {
		Logger.Infof("failed to read configuration files. %s", err)
	}

	if !cmd.Flags().Changed("username") {
		v, err := prompter.Username("Enter username", " "+cliConfig.GetDefaultUsername())

		if err != nil {
			return err
		}
		n.username = strings.TrimSpace(v)
	}

	if !n.noStorage && !cmd.Flags().Changed("password") {
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
		Schema:          "https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/master/tools/schema/session.schema.json",
		Host:            n.host,
		Tenant:          n.tenant,
		Username:        n.username,
		Description:     n.description,
		UseTenantPrefix: !n.noTenantPrefix,
	}

	session.MicroserviceAliases = make(map[string]string)

	settings := &CommandSettings{}
	settings.ActivityLog = &ActivityLogSettings{
		Enabled: settings.Bool(true),
	}

	if n.noStorage {
		settings.Storage = &StorageSettings{
			StorePassword: settings.Bool(false),
			StoreCookies:  settings.Bool(false),
		}
	}

	switch n.sessionType {
	case "dev":
		settings.Mode = &ModeSettings{
			Confirmation: "PUT POST DELETE",
			EnableCreate: settings.Bool(true),
			EnableUpdate: settings.Bool(true),
			EnableDelete: settings.Bool(true),
		}
	case "qual":
		settings.Mode = &ModeSettings{
			Confirmation: "PUT POST DELETE",
			EnableCreate: settings.Bool(true),
			EnableUpdate: settings.Bool(true),
			EnableDelete: settings.Bool(false),
		}
	case "prod":
		settings.Mode = &ModeSettings{
			Confirmation: "PUT POST DELETE",
			EnableCreate: settings.Bool(false),
			EnableUpdate: settings.Bool(false),
			EnableDelete: settings.Bool(false),
		}
	}

	if n.encrypt {
		settings.Encryption = &EncryptionSettings{
			Enabled:         settings.Bool(true),
			CachePassphrase: settings.Bool(true),
		}
	}

	session.Settings = settings

	if !n.noStorage {
		session.SetPassword(n.password)
	}

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
