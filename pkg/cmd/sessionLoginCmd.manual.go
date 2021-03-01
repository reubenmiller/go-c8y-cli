package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ylogin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type sessionLoginCmd struct {
	TFACode              string
	LoginErr             error
	LoginOK              bool
	ClearExistingCookies bool

	*baseCmd
}

func newSessionLoginCmd() *sessionLoginCmd {
	ccmd := &sessionLoginCmd{}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login into a cumulocity session",
		Long:  `Login and test the Cumulocity session and get either OAuth2 token, or using two factor authentication`,
		Example: `
c8y session login

Log into the current session
		`,
		RunE: ccmd.initSession,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.TFACode, "tfaCode", "", "Two Factor Authentication code")
	cmd.Flags().BoolVar(&ccmd.ClearExistingCookies, "clear", false, "Clear any existing cookies")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *sessionLoginCmd) onSave() {
	WriteAuth(viper.GetViper())
}

func (n *sessionLoginCmd) initSession(cmd *cobra.Command, args []string) error {
	if n.ClearExistingCookies {
		client.SetCookies([]*http.Cookie{})
	}

	// If the password is not encrypted, then save it (which will apply the encryption)
	if !cliConfig.IsPasswordEncrypted() {
		Logger.Infof("Password is unencrypted. enforcing encryption")
		n.onSave()
	}

	handler := c8ylogin.NewLoginHandler(client, cmd.ErrOrStderr(), n.onSave)

	handler.TFACode = n.TFACode
	handler.SetLogger(Logger)
	err := handler.Run()

	if err != nil {
		return err
	}

	if handler.C8Yclient.TenantName != "" && cliConfig.GetTenant() != handler.C8Yclient.TenantName {
		Logger.Infof("Saving tenant name")
		n.onSave()
	}

	printSessionInfo(n.cmd.ErrOrStderr(), CumulocitySession{
		Path:     cliConfig.GetSessionFilePath(),
		Host:     handler.C8Yclient.BaseURL.Host,
		Tenant:   cliConfig.GetTenant(),
		Username: handler.C8Yclient.Username,
	})

	return nil
}

func printSessionInfo(w io.Writer, session CumulocitySession) {
	label := color.New(color.FgWhite, color.Faint).SprintfFunc()
	value := color.New(color.FgWhite).SprintFunc()
	header := color.New(color.FgCyan).SprintFunc()
	fmt.Fprintf(w, label("%s", "---------------------  Cumulocity Session  ---------------------\n"))
	fmt.Fprintf(w, "\n    %s: %s\n\n\n", label("%s", "path"), header(hideSensitiveInformationIfActive(session.Path)))
	if session.Description != "" {
		fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "description")), value(hideSensitiveInformationIfActive(session.Host)))
	}

	fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "host")), value(hideSensitiveInformationIfActive(session.Host)))
	if session.Tenant != "" {
		fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "tenant")), value(hideSensitiveInformationIfActive(session.Tenant)))
	}
	fmt.Fprintf(w, "%s : %s\n", label(fmt.Sprintf("%-12s", "username")), value(hideSensitiveInformationIfActive(session.Username)))
	fmt.Fprintf(w, "\n")
}
