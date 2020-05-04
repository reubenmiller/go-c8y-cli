// PowerShell completions are based on the amazing work from clap:
// https://github.com/clap-rs/clap/blob/3294d18efe5f264d12c9035f404c7d189d4824e1/src/completions/powershell.rs
//
// The generated scripts require PowerShell v5.0+ (which comes Windows 10, but
// can be downloaded separately for windows 7 or 8.1).

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var powerShellCompletionTemplate = `using namespace System.Management.Automation
using namespace System.Management.Automation.Language
Register-ArgumentCompleter -Native -CommandName '%s' -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $commandElements = $commandAst.CommandElements
    $command = @(
        '%s'
        for ($i = 1; $i -lt $commandElements.Count; $i++) {
            $element = $commandElements[$i]
            if ($element -isnot [StringConstantExpressionAst] -or
                $element.StringConstantType -ne [StringConstantType]::BareWord -or
                $element.Value.StartsWith('-')) {
                break
            }
            if ($commandElements.Count -gt 2) {
                # Support partial completion
                if (!$wordToComplete) {
                    $element.Value
                } else {
                    if ("$($element.Value)".Length -ge $MinimumLength) {
                        $element.Value
                    }
                }
            } else {
                # Don't include the actual command until it is a certain length
                # Because this helps the use "c8y app<tab>" to be expanded to "c8y applications"
                $ValidCommands = @(%s)
                if ($ValidCommands.Contains($element.Value)) {
                    $element.Value
                }
            }
        }
    ) -join ';'
    $completions = @(switch ($command) {%s
    })
    $completions.Where{ $_.CompletionText -like "$wordToComplete*" } |
        Sort-Object -Property ListItemText
}`

type powershellCompletionHelper struct {
	cmd *cobra.Command
}

func newPowershellCompletionHelper(cmd *cobra.Command) *powershellCompletionHelper {
	return &powershellCompletionHelper{
		cmd: cmd,
	}
}

func nonCompletableFlag(flag *pflag.Flag) bool {
	return flag.Hidden || len(flag.Deprecated) > 0
}

func generatePowerShellSubcommandCases(out io.Writer, cmd *cobra.Command, previousCommandName string) {
	var cmdName string
	if previousCommandName == "" {
		cmdName = cmd.Name()
	} else {
		cmdName = fmt.Sprintf("%s;%s", previousCommandName, cmd.Name())
	}

	fmt.Fprintf(out, "\n        '%s' {", cmdName)

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if nonCompletableFlag(flag) {
			return
		}
		usage := escapeStringForPowerShell(flag.Usage)
		if len(flag.Shorthand) > 0 {
			fmt.Fprintf(out, "\n            [CompletionResult]::new('-%s', '%s', [CompletionResultType]::ParameterName, '%s')", flag.Shorthand, flag.Shorthand, usage)
		}
		fmt.Fprintf(out, "\n            [CompletionResult]::new('--%s', '%s', [CompletionResultType]::ParameterName, '%s')", flag.Name, flag.Name, usage)
	})

	for _, subCmd := range cmd.Commands() {
		usage := escapeStringForPowerShell(subCmd.Short)
		fmt.Fprintf(out, "\n            [CompletionResult]::new('%s', '%s', [CompletionResultType]::ParameterValue, '%s')", subCmd.Name(), subCmd.Name(), usage)
	}

	fmt.Fprint(out, "\n            break\n        }")

	for _, subCmd := range cmd.Commands() {
		generatePowerShellSubcommandCases(out, subCmd, cmdName)
	}
}

func escapeStringForPowerShell(s string) string {
	return strings.Replace(s, "'", "''", -1)
}

// GenPowerShellCompletion generates PowerShell completion file and writes to the passed writer.
func (c *powershellCompletionHelper) GenPowerShellCompletion(w io.Writer) error {
	buf := new(bytes.Buffer)

	var subCommandCases bytes.Buffer
	var commandNames bytes.Buffer

	cmd := c.cmd

	generatePowerShellSubcommandCases(&subCommandCases, cmd, "")
	generatePowerShellRootCommandList(&commandNames, cmd)

	fmt.Fprintf(buf, powerShellCompletionTemplate, cmd.Name(), cmd.Name(), commandNames.String(), subCommandCases.String())

	_, err := buf.WriteTo(w)
	return err
}

// generatePowerShellRootCommandList generate a list
func generatePowerShellRootCommandList(out io.Writer, cmd *cobra.Command) {
	prefix := ""
	for i, item := range cmd.Commands() {
		fmt.Fprintf(out, "%s\"%s\"", prefix, item.Name())
		if i == 0 {
			prefix = ","
		}
	}
}

// GenPowerShellCompletionFile generates PowerShell completion file.
func (c *powershellCompletionHelper) GenPowerShellCompletionFile(filename string) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return c.GenPowerShellCompletion(outFile)
}
