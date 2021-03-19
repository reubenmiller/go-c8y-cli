package cmd

import (
	"bytes"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

type CmdExecute struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdExecute(f *cmdutil.Factory) *CmdExecute {
	ccmd := &CmdExecute{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a jsonnet template",
		Long:  `Execute a jsonnet template and return the output. Useful when creating new templates`,
		Example: heredoc.Doc(`
Example 1:
$ c8y template execute --template ./mytemplate.jsonnet

Verify a jsonnet template and show the output after it is evaluated

Example 2:
$ c8y template execute --template ./mytemplate.jsonnet --templateVars "name=testme" --data "name=myDeviceName"

Verify a jsonnet template and specify input data to be used as the input when evaluating the template
		`),
		RunE: ccmd.newTemplate,
	}

	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithData(),
		f.WithTemplateFlag(cmd),
	)

	// Required flags
	_ = cmd.MarkFlagRequired(flags.FlagDataTemplateName)

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdExecute) newTemplate(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
	)

	if err != nil {
		return cmderrors.NewUserError(err)
	}

	responseText, err := body.MarshalJSON()
	if err != nil {
		return err
	}

	isJSONResponse := jsonUtilities.IsValidJSON([]byte(responseText))

	outputEnding := ""
	if len(responseText) > 0 {
		outputEnding = "\n"
	}

	if isJSONResponse {
		formatter := pretty.Pretty
		if cfg.CompactJSON() {
			formatter = pretty.UglyInPlace
		}
		fmt.Printf("%s%s", formatter(bytes.TrimSpace(responseText)), outputEnding)
	} else {
		fmt.Printf("%s%s", bytes.TrimSpace(responseText), outputEnding)
	}
	return nil
}