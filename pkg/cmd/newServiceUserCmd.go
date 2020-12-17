package cmd

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type newServiceUserCmd struct {
	*baseCmd

	name        string
	key         string
	contextPath string
	tenants     []string
	roles       []string
}

func newNewServiceUserCmd() *newServiceUserCmd {
	ccmd := &newServiceUserCmd{}

	cmd := &cobra.Command{
		Use:   "createServiceUser",
		Short: "New application service user",
		Long:  ``,
		Example: `
$ c8y microservices createServiceUser --name my-user
Create new application service user
		`,
		PreRunE: validateCreateMode,
		RunE:    ccmd.doProcedure,
	}

	cmd.SilenceUsage = true

	addDataFlagWithoutTemplates(cmd)
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of application")
	cmd.Flags().StringArrayVar(&ccmd.tenants, "tenants", []string{}, "Tenant to subscribe to. If left blank than the application will not generate the service user")
	cmd.Flags().StringArrayVar(&ccmd.roles, "roles", []string{}, "Roles which should be assigned to the service user")

	// Required flags
	cmd.MarkFlagRequired("name")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *newServiceUserCmd) getApplicationDetails() *c8y.Application {

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

func (n *newServiceUserCmd) doProcedure(cmd *cobra.Command, args []string) error {
	var application *c8y.Application
	var response *c8y.Response
	var err error

	applicationDetails := n.getApplicationDetails()

	if applicationDetails.Name == "" {
		return newUserError("Could not detect application name for the given input")
	}

	Logger.Info("Creating new application")
	application, response, err = client.Application.Create(context.Background(), applicationDetails)

	if err != nil {
		return fmt.Errorf("failed to create microservice. %s", err)
	}

	// App subscription
	if len(n.tenants) > 0 {
		if !globalFlagDryRun {
			for _, tenant := range n.tenants {
				Logger.Infof("Subscribing to application in tenant %s", tenant)
				_, resp, err := client.Tenant.AddApplicationReference(context.Background(), tenant, application.Self)

				if err != nil {
					if resp != nil && resp.StatusCode == 409 {
						Logger.Infof("microservice is already enabled")
					} else {
						return fmt.Errorf("Failed to subscribe to application. %s", err)
					}
				}
			}
		}
	}

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	_, err = processResponse(response, err, commonOptions)
	return err
}
