// v2-based audit records create: the CLI builds the full body (typed fields +
// --data/--template) and posts it as-is. Audit sources are ids (no device-name
// resolution), so no source-reference resolution is needed.
package create

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// CreateCmd command
type CreateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateCmd creates a command to Create audit record
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create audit record",
		Long:  `Create a new audit record`,
		Example: heredoc.Doc(`
$ c8y auditrecords create --type "Custom" --source 12345 --activity "Did something" --text "details"
Create an audit record for a source object
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("type", "", "Identifies the type of this audit record")
	cmd.Flags().String("time", "", "Time of the audit record. Defaults to current timestamp")
	cmd.Flags().String("text", "", "Text description of the audit record")
	cmd.Flags().String("source", "", "The ManagedObject id that the audit record originated from (accepts pipeline)")
	cmd.Flags().String("activity", "", "The activity that was carried out")
	cmd.Flags().String("severity", "", "The severity: critical, major, minor, warning or information")
	cmd.Flags().String("user", "", "The user responsible for the audited action")
	cmd.Flags().String("application", "", "The application used to carry out the audited action")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("source", "source.id", false, "id", "source.id", "managedObject.id"),
		flags.WithPowershellName("New-AuditRecord"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.auditRecord+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.Input(); err != nil {
		return err
	}

	err = r.Body(
		flags.WithOverrideValue("source", "source.id"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("type", "type"),
		flags.WithRelativeTimestamp("time", "time"),
		flags.WithStringValue("text", "text"),
		flags.WithStringValue("source", "source.id"),
		flags.WithStringValue("activity", "activity"),
		flags.WithStringValue("severity", "severity"),
		flags.WithStringValue("user", "user"),
		flags.WithStringValue("application", "application"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithDefaultTemplateString(`
{time: _.Now('0s')}`),
		flags.WithRequiredProperties("activity", "source.id", "type", "time"),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.AuditRecord] {
				return client.AuditRecords.Create(ctx, body)
			})
		}, nil
	})
}
