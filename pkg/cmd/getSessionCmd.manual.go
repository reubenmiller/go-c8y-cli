package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/prompt"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

// CumulocitySessionDetails public details about the current session
type CumulocitySessionDetails struct {
	CumulocitySession

	Path string `json:"path"`
	Name string `json:"name"`
}

type getSessionCmd struct {
	OutputJSON bool
	prompter   *prompt.Prompt
	*baseCmd
}

func newGetSessionCmd() *getSessionCmd {
	ccmd := &getSessionCmd{}
	ccmd.prompter = prompt.NewPrompt(Logger)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get session information",
		Long:  `Get session information`,
		Example: `
Get the details about the current session
$ c8y sessions get

Get the details about the current session which is specified via the --session argument
$ c8y sessions get --session mycustomsession
		`,
		RunE: ccmd.getSession,
	}

	cmd.Flags().BoolVar(&ccmd.OutputJSON, "json", false, "Output passphrase in json")
	cmd.SilenceUsage = true

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getSessionCmd) getSession(cmd *cobra.Command, args []string) error {

	session := CumulocitySessionDetails{
		Path: cliConfig.GetSessionFilePath(),
		Name: cliConfig.GetName(),
		CumulocitySession: CumulocitySession{
			Host:            cliConfig.GetHost(),
			Tenant:          cliConfig.GetTenant(),
			Username:        cliConfig.GetUsername(),
			Description:     cliConfig.GetDescription(),
			UseTenantPrefix: cliConfig.Persistent.GetBool("useTenantPrefix"),
		},
	}

	if session.CumulocitySession.Host == "" {
		return newUserError("no session is loaded")
	}

	b, err := json.Marshal(session)
	if err != nil {
		return err
	}

	outputEnding := "\n"

	if globalFlagCompact {
		fmt.Printf("%s%s", bytes.TrimSpace(b), outputEnding)
	} else {
		fmt.Printf("%s%s", pretty.Pretty(bytes.TrimSpace(b)), outputEnding)
	}
	return nil
}
