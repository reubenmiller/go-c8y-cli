// Prototype of a v2-based local (non-API) command: the jsonnet template is
// evaluated per input item entirely client-side, and the results flow through
// the same output pipeline as API commands, so --filter, --select,
// --outputTemplate and --outputFile behave identically.
package execute

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

type CmdExecute2 struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdExecute2(f *cmdutil.Factory) *CmdExecute2 {
	ccmd := &CmdExecute2{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "execute2",
		Short: "Execute a jsonnet template",
		Long:  `Execute a jsonnet template and return the output. Useful when creating new templates`,
		Example: heredoc.Doc(`
$ c8y template execute2 --template ./mytemplate.jsonnet
Verify a jsonnet template and show the output after it is evaluated

$ echo '{"name": "external_source"}' | c8y template execute2 --template "{name: input.value.name}"
Pass external json data into the template, and reference it via the "input.value" variable

$ c8y devices list | c8y template execute2 --template "{id: input.value.id}" --select id --outputFile ids.json
Evaluate a template per piped device and tee the results to a file
		`),
		RunE: ccmd.RunE,
	}

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	cmd.Flags().String("input", "", "input (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("input", "", false),
	)

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd).SetRequiredFlags(flags.FlagDataTemplateName)

	return ccmd
}

func (n *CmdExecute2) RunE(cmd *cobra.Command, args []string) error {
	if !cmd.Flags().Changed(flags.FlagDataTemplateName) {
		return &flags.ParameterError{
			Name: flags.FlagDataTemplateName,
			Err:  flags.ErrParameterMissing,
		}
	}

	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	// body: the jsonnet template plus any bound pipeline input
	err = r.Bind(c8ystream.Body(
		flags.WithOverrideValue("input", "input"),
		flags.WithDataFlagValue(),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithStringValue("input", "input", ""),
	))
	if err != nil {
		return err
	}

	// No API client: the template result is yielded as a document directly,
	// so everything downstream is shared with API commands.
	return r.Run(func(ctx context.Context, args c8ystream.Args) output.Seq {
		return c8ystream.FromDocs(jsondoc.New(formatOutput(args.Body)))
	})
}
