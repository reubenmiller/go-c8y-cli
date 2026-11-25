package serviceuser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/spf13/cobra"
)

type CmdCreate struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	name    string
	tenants []string
	roles   []string
}

func NewCmdCreate(f *cmdutil.Factory) *CmdCreate {
	ccmd := &CmdCreate{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "New application service user",
		Long:  ``,
		Example: heredoc.Doc(`
$ c8y microservices serviceusers create --name my-user
Create new application service user
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of application")
	cmd.Flags().StringSliceVar(&ccmd.tenants, "tenants", []string{}, "Tenant to subscribe to. If left blank than the application will not generate the service user")
	cmd.Flags().StringSliceVar(&ccmd.roles, "roles", []string{}, "Roles which should be assigned to the service user")

	flags.WithOptions(
		cmd,
		flags.WithData(),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd).SetRequiredFlags("name")

	return ccmd
}

// applicationDetails is the body of the new application. It is a struct (not a
// map) to keep the order of the properties.
type applicationDetails struct {
	Key           string   `json:"key,omitempty"`
	Name          string   `json:"name,omitempty"`
	Type          string   `json:"type,omitempty"`
	ContextPath   string   `json:"contextPath,omitempty"`
	RequiredRoles []string `json:"requiredRoles,omitempty"`
}

func (n *CmdCreate) getApplicationDetails() *applicationDetails {
	app := applicationDetails{
		Name: n.name,
		Key:  n.name + "-key",
		Type: "MICROSERVICE",
	}
	app.ContextPath = app.Name

	if len(n.roles) > 0 {
		app.RequiredRoles = n.roles
	}

	return &app
}

// RunE creates the application, and subscribes the tenants to it. The
// application is the result of the command.
func (n *CmdCreate) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		details := n.getApplicationDetails()
		if details.Name == "" {
			return nil, cmderrors.NewUserError("Could not detect application name for the given input")
		}
		body, err := json.Marshal(details)
		if err != nil {
			return nil, err
		}

		return func(ctx context.Context) any {
			return func(yield func(jsondoc.JSONDoc, error) bool) {
				log.Info("Creating new application")
				app := client.Applications.Create(ctx, body)
				if app.Err != nil {
					yield(jsondoc.Empty(), fmt.Errorf("failed to create microservice. %w", app.Err))
					return
				}

				// Subscribe the tenants (in a dry run too, to show the requests)
				for _, tenant := range n.tenants {
					log.Infof("Subscribing to application in tenant %s", tenant)
					// An existing subscription (409) is not an error
					if sub := client.Applications.Subscribe(ctx, tenant, app.Data.Self()); sub.Err != nil {
						yield(jsondoc.Empty(), fmt.Errorf("Failed to subscribe to application. %w", sub.Err))
						return
					}
				}

				for doc, err := range app.Documents() {
					if !yield(doc, err) {
						return
					}
				}
			}
		}, nil
	})
}
