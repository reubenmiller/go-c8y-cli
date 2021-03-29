package create

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/pkg/prompt"
	"github.com/spf13/cobra"
)

func NewCumulocitySessionFromFile(filePath string, log *logger.Logger, cfg *config.Config) (*c8ysession.CumulocitySession, error) {
	session := &c8ysession.CumulocitySession{
		Config: cfg,
		Logger: log,
	}
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

type CmdCreate struct {
	host           string
	username       string
	password       string
	token          string
	description    string
	name           string
	tenant         string
	sessionType    string
	noTenantPrefix bool
	noStorage      bool
	encrypt        bool

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdCreate(f *cmdutil.Factory) *CmdCreate {
	ccmd := &CmdCreate{
		factory: f,
	}

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
		RunE:              ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.host, "host", "", "Host. .e.g. test.cumulocity.com. (required)")
	cmd.Flags().StringVar(&ccmd.username, "username", "", "Username (without tenant). (required)")
	cmd.Flags().StringVar(&ccmd.password, "password", "", "Password. If left blank then you will be prompted for the password")
	cmd.Flags().StringVar(&ccmd.token, "token", "", "Token")
	cmd.Flags().StringVar(&ccmd.tenant, "tenant", "", "Tenant ID")
	cmd.Flags().StringVar(&ccmd.description, "description", "", "Description about the session")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of the session")
	cmd.Flags().StringVar(&ccmd.sessionType, "type", "", "Session type. List of predefined session types")
	cmd.Flags().BoolVar(&ccmd.noTenantPrefix, "noTenantPrefix", false, "Don't use tenant name as a prefix to the user name when using Basic Authentication. Defaults to false")
	cmd.Flags().BoolVar(&ccmd.noStorage, "noStorage", false, "Don't store any passwords or tokens in the session file")
	cmd.Flags().BoolVar(&ccmd.encrypt, "encrypt", false, "Encrypt passwords and tokens (occurs when logging in)")

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

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdCreate) promptArgs(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	prompter := prompt.NewPrompt(log)

	if !cmd.Flags().Changed("username") {
		v, err := prompter.Username("Enter username", " "+cfg.GetDefaultUsername())

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

func (n *CmdCreate) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	session := &c8ysession.CumulocitySession{
		Schema:          "https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/master/tools/schema/session.schema.json",
		Host:            n.host,
		Tenant:          n.tenant,
		Username:        n.username,
		Description:     n.description,
		UseTenantPrefix: !n.noTenantPrefix,
		Config:          cfg,
		Logger:          log,
	}

	session.MicroserviceAliases = make(map[string]string)

	settings := &config.CommandSettings{}
	settings.ActivityLog = &config.ActivityLogSettings{
		Enabled: settings.Bool(true),
	}

	if n.noStorage {
		settings.Storage = &config.StorageSettings{
			StorePassword: settings.Bool(false),
			StoreToken:    settings.Bool(false),
		}
	}

	switch n.sessionType {
	case "dev":
		settings.Mode = &config.ModeSettings{
			Confirmation: "PUT POST DELETE",
			EnableCreate: settings.Bool(true),
			EnableUpdate: settings.Bool(true),
			EnableDelete: settings.Bool(true),
		}
	case "qual":
		settings.Mode = &config.ModeSettings{
			Confirmation: "PUT POST DELETE",
			EnableCreate: settings.Bool(true),
			EnableUpdate: settings.Bool(true),
			EnableDelete: settings.Bool(false),
		}
	case "prod":
		settings.Mode = &config.ModeSettings{
			Confirmation: "PUT POST DELETE",
			EnableCreate: settings.Bool(false),
			EnableUpdate: settings.Bool(false),
			EnableDelete: settings.Bool(false),
		}
	}

	if n.encrypt {
		settings.Encryption = &config.EncryptionSettings{
			Enabled:         settings.Bool(true),
			CachePassphrase: settings.Bool(true),
		}
	}

	session.Settings = settings

	if !n.noStorage {
		session.SetPassword(n.password)
		session.SetToken(n.token)
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

	outputDir := cfg.GetSessionHomeDir()
	outputFile := n.formatFilename(sessionName)

	if err := n.writeSessionFile(outputDir, outputFile, *session); err != nil {
		return err
	}

	fmt.Println(path.Join(outputDir, outputFile))
	return nil
}

func (n *CmdCreate) formatFilename(name string) string {
	if !strings.HasSuffix(name, ".json") {
		name = fmt.Sprintf("%s.json", name)
	}
	return name
}

func (n *CmdCreate) writeSessionFile(outputDir, outputFile string, session c8ysession.CumulocitySession) error {
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")

	if err != nil {
		return errors.Wrap(err, "failed to convert session to json")
	}

	outputPath := path.Join(outputDir, outputFile)

	if outputDir != "" {
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			log.Errorf("failed to create folder. folder=%s, err=%s", outputDir, err)
			return err
		}
	}
	log.Debugf("output file: %s", outputPath)

	if err := ioutil.WriteFile(path.Join(outputDir, outputFile), data, 0644); err != nil {
		return errors.Wrap(err, "failed to write to file")
	}
	return nil
}
