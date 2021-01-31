package cmd

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type getServiceUserCmd struct {
	*baseCmd
}

func NewGetServiceUserCmd() *getServiceUserCmd {
	ccmd := &getServiceUserCmd{}

	cmd := &cobra.Command{
		Use:   "getServiceUser",
		Short: "Get microservice service user",
		Long: `Get the service user associated to a microservice
`,
		Example: `
$ c8y microservices getServiceUser --id 12345
Get application service user by app id

$ c8y microservices getServiceUser --id myapp
Get application service user by app name
        `,
		PreRunE: nil,
		RunE:    ccmd.getServiceUser,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Microservice id (required)")

	// Required flags
	cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getServiceUserCmd) getServiceUser(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// path parameters
	appIDs := []string{}

	if cmd.Flags().Lookup("id") != nil {
		idInputValues, idValue, err := getMicroserviceSlice(cmd, args, "id")

		if err != nil {
			return newUserError("no matching microservices found", idInputValues, err)
		}

		if len(idValue) == 0 {
			return newUserError("no matching microservices found", idInputValues)
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
			return newUserError(err)
		}

		auth := c8y.NewBasicAuthString(bootstrapUser.Tenant, bootstrapUser.Username, bootstrapUser.Password)
		bootstrapCtx := context.WithValue(context.Background(), c8y.GetContextAuthTokenKey(), auth)

		_, resp, err := client.Application.GetCurrentApplicationSubscriptions(bootstrapCtx)
		processResponse(resp, err, commonOptions)
	}

	return nil
}
