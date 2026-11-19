// v2-based audit records list: fills the typed auditrecords.ListOptions. Audit
// sources are ids (no device-name resolution). Time-keyset pagination.
package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/auditrecords"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// ListCmd command
type ListCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewListCmd creates a command to Get audit record collection
func NewListCmd(f *cmdutil.Factory) *ListCmd {
	ccmd := &ListCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get audit record collection",
		Long:  `Get a collection of audit records based on filter parameters`,
		Example: heredoc.Doc(`
$ c8y auditrecords list --type Alarm --user admin
Get audit records of type Alarm created by admin

$ c8y auditrecords list --source 12345
Get audit records for a given source object
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("source", "", "Source id of the audit record (accepts pipeline)")
	cmd.Flags().String("type", "", "Audit record type")
	cmd.Flags().String("user", "", "Username")
	cmd.Flags().String("application", "", "Application")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of audit record occurrence")
	cmd.Flags().String("dateTo", "", "End date or date and time of audit record occurrence")
	cmd.Flags().Bool("revert", false, "Return the newest instead of the oldest audit records. Must be used with dateFrom and dateTo parameters")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("source", "source", false, "id", "source.id", "managedObject.id"),
		flags.WithCollectionProperty("auditRecords"),
		flags.WithPowershellName("Get-AuditRecordCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.auditRecordCollection+json", "application/vnd.com.nsn.cumulocity.auditRecord+json"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *ListCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("source"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	common, err := r.Config.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	rawOutput := r.Config.RawOutput()
	paginationStrategy := pagination.StrategyKind(r.Config.PaginationStrategy())

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		opt := auditrecords.ListOptions{
			Source:      in.String("source"),
			Type:        in.String("type"),
			User:        in.String("user"),
			Application: in.String("application"),
			DateFrom:    in.TimeValue("dateFrom"),
			DateTo:      in.TimeValue("dateTo"),
			Revert:      in.Bool("revert"),
		}
		opt.PaginationOptions = pagination.PaginationOptions{
			PageSize:          common.PageSize,
			WithTotalPages:    common.WithTotalPages,
			WithTotalElements: common.WithTotalElements,
			CurrentPage:       int(common.CurrentPage),
			MaxItems:          r.Config.MaxItems(),
			Strategy:          paginationStrategy,
		}
		return c8ystream.ListCall(rawOutput, opt, client.AuditRecords.ListAll), nil
	})
}
