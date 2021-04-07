package execute

import (
	"bytes"
	"fmt"
	"io"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/spf13/cobra"
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

Example 3:
$ echo '{"name": "external_source"}' | c8y template execute --template "{name: input.value.name, nestedValues: input.value}"
$ => {"name":"external_source","nestedValues":{"name":"external_source"}}

Pass external json data into the template, and reference it via the "input.value" variable
		`),
		RunE: ccmd.newTemplate,
	}

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("input", "input", false),
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
		flags.WithStringValue("input", "input", ""),
	)

	if err != nil {
		return cmderrors.NewUserError(err)
	}

	for {
		responseText, err := body.MarshalJSON()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		isJSONResponse := jsonUtilities.IsValidJSON([]byte(responseText))

		outputEnding := ""
		if len(responseText) > 0 {
			outputEnding = "\n"
		}

		if isJSONResponse {
			if err := n.factory.WriteJSONToConsole(cfg, cmd, "", responseText); err != nil {
				return err
			}
		} else {
			fmt.Printf("%s%s", bytes.TrimSpace(responseText), outputEnding)
		}
	}

	return nil
}
