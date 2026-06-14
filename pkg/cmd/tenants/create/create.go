// v2-based tenants create: the CLI builds the full body (typed fields +
// --data/--template) and posts it as-is. A tenant has no device references, so
// no resolution is needed (plain CRUD).
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

// NewCreateCmd creates a command to Create tenant
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create tenant",
		Long:  `Create a new tenant`,
		Example: heredoc.Doc(`
$ c8y tenants create --name "mycompany" --domain "mycompany" --adminEmail "admin@example.com" --adminName "admin" --adminPass "mys3curep9d8"
Create a new tenant (from the management tenant)

$ c8y tenants create --name "mycompany" --domain "mycompany" --adminEmail "admin@example.com" --adminName "admin" --sendPasswordResetEmail
Create a new tenant and send a password reset email (from the management tenant)
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("company", "", "Company name. Maximum 256 characters")
	cmd.Flags().String("name", "", "Company name. Maximum 256 characters")
	cmd.Flags().String("domain", "", "Domain name to be used for the tenant. Maximum 256 characters (accepts pipeline)")
	cmd.Flags().String("adminEmail", "", "Email address of the tenant's administrator")
	cmd.Flags().String("adminName", "", "Username of the tenant administrator")
	cmd.Flags().String("adminPass", "", "Password of the tenant administrator")
	cmd.Flags().String("contactName", "", "A contact name, for example an administrator, of the tenant")
	cmd.Flags().String("contactPhone", "", "An international contact phone number")
	cmd.Flags().String("tenantId", "", "The tenant ID. This should be left bank unless you know what you are doing. Will be auto-generated if not present.")
	cmd.Flags().Bool("allowCreateTenants", false, "Allow the tenant to create sub-tenants")
	cmd.Flags().Bool("sendPasswordResetEmail", false, "Send password reset email to the user instead of setting a password")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("domain", "domain", false, "id"),
		flags.WithPowershellName("New-Tenant"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.tenant+json", ""),
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
		flags.WithOverrideValue("domain", "domain"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("company", "company"),
		flags.WithStringValue("name", "company"),
		flags.WithStringValue("domain", "domain"),
		flags.WithStringValue("adminEmail", "adminEmail"),
		flags.WithStringValue("adminName", "adminName"),
		flags.WithStringValue("adminPass", "adminPass"),
		flags.WithStringValue("contactName", "contactName"),
		flags.WithStringValue("contactPhone", "contactPhone"),
		flags.WithStringValue("tenantId", "tenantId"),
		flags.WithBoolValue("allowCreateTenants", "allowCreateTenants", ""),
		flags.WithBoolValue("sendPasswordResetEmail", "sendPasswordResetEmail", ""),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("company", "domain", "adminName", "adminEmail"),
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
			return c8ystream.Submit(ctx, func(ctx context.Context) op.Result[jsonmodels.Tenant] {
				return client.Tenants.Create(ctx, body)
			})
		}, nil
	})
}
