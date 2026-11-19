// v2-based bulk tenant option update: the category flag drives iteration (pipe
// or --category) and the --data/--template body supplies the option map, which
// is updated for the category via Tenants.Options.UpdateByCategory. The endpoint
// returns a flat key/value map. The endpoint stores option values as strings, so
// non-string --data values are coerced to their string form.
package updatebulk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/tenants/tenantoptions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// UpdateBulkCmd command
type UpdateBulkCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUpdateBulkCmd creates a command to Update multiple tenant options
func NewUpdateBulkCmd(f *cmdutil.Factory) *UpdateBulkCmd {
	ccmd := &UpdateBulkCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "updateBulk",
		Short: "Update multiple tenant options",
		Long:  `Update multiple tenant options in provided category`,
		Example: heredoc.Doc(`
$ c8y tenantoptions updateBulk --category "c8y_cli_tests" --data "{\"option5\":\"0\",\"option6\":\"1\"}"
Update multiple tenant options
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("category", "", "Tenant Option category (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("category", "category", true, "id"),
		flags.WithPowershellName("Update-TenantOptionBulk"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.option+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateBulkCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("category"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
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
		options, err := bodyToStringMap(body)
		if err != nil {
			return nil, err
		}
		opt := tenantoptions.UpdateByCategoryOption{
			Category: in.String("category"),
			Body:     options,
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromJSONValue(client.Tenants.Options.UpdateByCategory(ctx, opt), apiv2.IsDryRun(ctx))
		}, nil
	})
}

// bodyToStringMap converts the built body document into the string/string map
// the category endpoint expects, coercing non-string values to their string
// form (the server stores all option values as strings).
func bodyToStringMap(body []byte) (map[string]string, error) {
	if len(body) == 0 {
		return map[string]string{}, nil
	}
	raw := map[string]any{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		if s, ok := v.(string); ok {
			out[k] = s
			continue
		}
		out[k] = fmt.Sprintf("%v", v)
	}
	return out, nil
}
