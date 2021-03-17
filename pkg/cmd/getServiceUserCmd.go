package cmd

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type getServiceUserCmd struct {
	*subcommand.SubCommand
}

func NewGetServiceUserCmd() *getServiceUserCmd {
	ccmd := &getServiceUserCmd{}

	cmd := &cobra.Command{
		Use:   "getServiceUser",
		Short: "Get microservice service user",
		Long: `Get the service user associated to a microservice
`,
		Example: heredoc.Doc(`
$ c8y microservices getServiceUser --id 12345
Get application service user by app id

$ c8y microservices getServiceUser --id myapp
Get application service user by app name
        `),
		PreRunE: nil,
		RunE:    ccmd.getServiceUser,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Microservice id (required)")

	// Required flags
	_ = cmd.MarkFlagRequired("id")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *getServiceUserCmd) getServiceUser(cmd *cobra.Command, args []string) error {

	commonOptions, err := cliConfig.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// path parameters
	appIDs := []string{}

	if cmd.Flags().Lookup("id") != nil {
		idInputValues, idValue, err := getMicroserviceSlice(cmd, args, "id")

		if err != nil {
			return cmderrors.NewUserError("no matching microservices found", idInputValues, err)
		}

		if len(idValue) == 0 {
			return cmderrors.NewUserError("no matching microservices found", idInputValues)
		}

		for _, item := range idValue {
			if item != "" {
				appIDs = append(appIDs, newIDValue(item).GetID())
			}
		}
	}

	for _, appID := range appIDs {
		// get bootstrap user
		ctx, cancel := getTimeoutContext()
		defer cancel()
		bootstrapUser, _, err := client.Application.GetApplicationUser(ctx, appID)
		if err != nil {
			return cmderrors.NewUserError(err)
		}

		auth := c8y.NewBasicAuthString(bootstrapUser.Tenant, bootstrapUser.Username, bootstrapUser.Password)
		bootstrapCtx := context.WithValue(context.Background(), c8y.GetContextAuthTokenKey(), auth)

		_, resp, err := client.Application.GetCurrentApplicationSubscriptions(bootstrapCtx)
		if _, err := processResponse(resp, err, commonOptions); err != nil {
			return err
		}
	}

	return nil
}
