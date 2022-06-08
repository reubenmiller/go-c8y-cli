package create

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
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
			return f.CreateModeEnabled()
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

func (n *CmdCreate) getApplicationDetails() *c8y.Application {

	app := c8y.Application{}

	// Set application properties
	app.Name = n.name
	app.Key = app.Name + "-key"
	app.Type = "MICROSERVICE"
	app.ContextPath = app.Name

	if len(n.roles) > 0 {
		app.RequiredRoles = n.roles
	}

	return &app
}

func (n *CmdCreate) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	var application *c8y.Application
	var response *c8y.Response

	applicationDetails := n.getApplicationDetails()

	if applicationDetails.Name == "" {
		return cmderrors.NewUserError("Could not detect application name for the given input")
	}

	log.Info("Creating new application")
	application, response, err = client.Application.Create(context.Background(), applicationDetails)

	if err != nil {
		return fmt.Errorf("failed to create microservice. %s", err)
	}

	// App subscription
	if len(n.tenants) > 0 {
		for _, tenant := range n.tenants {
			log.Infof("Subscribing to application in tenant %s", tenant)
			_, resp, err := client.Tenant.AddApplicationReference(context.Background(), tenant, application.Self)

			if err != nil {
				if resp != nil && resp.StatusCode == 409 {
					log.Infof("microservice is already enabled")
				} else {
					return fmt.Errorf("Failed to subscribe to application. %s", err)
				}
			}
		}
	}

	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	handler, err := n.factory.GetRequestHandler()
	if err != nil {
		return err
	}
	_, err = handler.ProcessResponse(response, err, commonOptions)
	return err
}
