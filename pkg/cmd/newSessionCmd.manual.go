package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/howeyc/gopass"
	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tidwall/pretty"
)

type CumulocitySessions struct {
	Sessions []CumulocitySession `json:"sessions"`
}

type Authentication struct {
	AuthType string         `json:"authType"`
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

	Authentication `json:"cookies,omitempty"`

	MicroserviceAliases map[string]string `json:"microserviceAliases"`

	Index int    `json:"-"`
	Path  string `json:"-"`
	Name  string `json:"-"`
}

type LoginInformation struct {
	Cookies  []string `json:"cookies"`
	TenantId string   `json:"tenantId"`
}

func WriteAuth(v *viper.Viper) error {
	cliConfig.SetAuthorizationCookies(client.Cookies)
	cliConfig.SetPassword(client.Password)
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

	session.Path = filePath

	basename := filepath.Base(filePath)
	extension := filepath.Ext(basename)
	session.Name = strings.TrimSuffix(basename, extension)
	return session, nil
}

func (s CumulocitySession) GetSessionPassphrase() string {
	return os.Getenv("C8Y_SESSION_PASSPHRASE")
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
			--tenant "myTenant" \
			--username "myUser@me.com"
		
		// Create a new session and prompt for the password
		`,
		RunE: ccmd.newSession,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("host", "", "Host. .e.g. test.cumulocity.com. (required)")
	cmd.Flags().String("tenant", "", "Tenant. (required)")
	cmd.Flags().String("username", "", "Username (without tenant). (required)")
	cmd.Flags().String("password", "", "Password. If left blank then you will be prompted for the password")
	cmd.Flags().String("description", "", "Description about the session")
	cmd.Flags().String("name", "", "Name of the session")
	cmd.Flags().String("microserviceAliases", "", "Name of the session")
	cmd.Flags().Bool("noTenantPrefix", false, "Don't use tenant name as a prefix to the user name when using Basic Authentication. Defaults to false")

	// Required flags
	cmd.MarkFlagRequired("host")
	cmd.MarkFlagRequired("tenant")
	cmd.MarkFlagRequired("username")
	// cmd.MarkFlagRequired("password")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newSessionCmd) newSession(cmd *cobra.Command, args []string) error {

	session := &CumulocitySession{}
	session.MicroserviceAliases = make(map[string]string)

	if v, err := cmd.Flags().GetString("host"); err == nil && v != "" {
		session.SetHost(v)
	}
	if v, err := cmd.Flags().GetString("tenant"); err == nil && v != "" {
		session.Tenant = v
	}

	if cmd.Flags().Changed("password") {
		if v, err := cmd.Flags().GetString("password"); err == nil && v != "" {
			session.SetPassword(v)
		}
	} else {
		cmd.Printf("Enter password: ")
		password, _ := gopass.GetPasswd() // Silent
		session.SetPassword(string(password))
	}
	if v, err := cmd.Flags().GetString("username"); err == nil && v != "" {
		session.Username = v
	}
	if v, err := cmd.Flags().GetString("description"); err == nil && v != "" {
		session.Description = v
	}

	if v, err := cmd.Flags().GetBool("noTenantPrefix"); err == nil {
		session.UseTenantPrefix = !v
	}

	// session name (default to host and username)
	sessionName := session.Tenant + "-" + session.Username
	if v, err := cmd.Flags().GetString("name"); err == nil && v != "" {
		sessionName = v
	}

	outputDir := getSessionHomeDir()
	outputFile := n.formatFilename(sessionName)

	if err := n.writeSessionFile(outputDir, outputFile, *session); err != nil {
		return err
	}

	fmt.Println(path.Join(outputDir, outputFile))
	// if str, err := json.Marshal(session); err == nil {
	// 	fmt.Printf("%s\n", str)
	// }

	return nil
}

func (n *newSessionCmd) doNewSession(method string, path string, query string, body map[string]interface{}) error {
	resp, err := client.SendRequest(
		context.Background(),
		c8y.RequestOptions{
			Method:       method,
			Path:         path,
			Query:        query,
			Body:         body,
			IgnoreAccept: false,
			DryRun:       globalFlagDryRun,
		})

	if err != nil {
		color.Set(color.FgRed, color.Bold)
	}

	if resp != nil && resp.JSONData != nil {
		if globalFlagPrettyPrint {
			fmt.Printf("%s\n", pretty.Pretty([]byte(*resp.JSONData)))
		} else {
			fmt.Printf("%s\n", *resp.JSONData)
		}
	}

	color.Unset()

	if err != nil {
		return newSystemError("command failed", err)
	}
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

	Logger.Debugf("output file: %s", outputPath)

	if err := ioutil.WriteFile(path.Join(outputDir, outputFile), data, 0644); err != nil {
		return errors.Wrap(err, "failed to write to file")
	}
	return nil
}
