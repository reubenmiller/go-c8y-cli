package cmd

import (
	"bytes"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
)

type executeTemplateCmd struct {
	*baseCmd
}

func newExecuteTemplateCmd() *executeTemplateCmd {
	ccmd := &executeTemplateCmd{}

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a jsonnet template",
		Long:  `Execute a jsonnet template and return the output. Useful when creating new templates`,
		Example: `
Example 1:
$ c8y template execute --template ./mytemplate.jsonnet

Verify a jsonnet template and show the output after it is evaluated

Example 2:
$ c8y template execute --template ./mytemplate.jsonnet --templateVars "name=testme" --data "name=myDeviceName"

Verify a jsonnet template and specify input data to be used as the input when evaluating the template
		`,
		RunE: ccmd.newTemplate,
	}

	cmd.SilenceUsage = true

	addDataFlag(cmd)

	// Required flags
	cmd.MarkFlagRequired(FlagDataTemplateName)

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *executeTemplateCmd) newTemplate(cmd *cobra.Command, args []string) error {

	// body
	body := mapbuilder.NewMapBuilder()
	body.SetMap(getDataFlag(cmd))
	if err := setDataTemplateFromFlags(cmd, body); err != nil {
		return newUserError("Template error. ", err)
	}
	if err := body.Validate(); err != nil {
		return newUserError("Body validation error. ", err)
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

	if globalFlagPrettyPrint && isJSONResponse {
		fmt.Printf("%s%s", pretty.Pretty(bytes.TrimSpace(responseText)), outputEnding)
	} else {
		fmt.Printf("%s%s", bytes.TrimSpace(responseText), outputEnding)
	}
	return nil
}
