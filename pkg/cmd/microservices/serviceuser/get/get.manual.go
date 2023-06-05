package get

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdGet struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdGet(f *cmdutil.Factory) *CmdGet {
	ccmd := &CmdGet{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get microservice service user",
		Long: `Get the service user associated to a microservice
`,
		Example: heredoc.Doc(`
$ c8y microservices serviceuser get --id 12345
Get application service user by app id

$ c8y microservices serviceuser get --id myapp
Get application service user by app name
        `),
		PreRunE: nil,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Microservice id (required)")

	completion.WithOptions(
		cmd,
		completion.WithMicroservice("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true, "id"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdGet) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// path parameters
	appIDs := []string{}

	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// path parameters
	path := flags.NewStringTemplate("{id}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithMicroserviceByNameFirstMatch(n.factory, args, "id", "id"),
	)
	if err != nil {
		return err
	}

	remainingJobs := cfg.GetMaxJobs()

	for {
		v, _, err := path.Execute(false)

		if len(v) > 0 {
			appIDs = append(appIDs, v)
		}

		if err != nil {
			break
		}

		remainingJobs--
		if remainingJobs <= 0 {
			break
		}
	}

	handler, err := n.factory.GetRequestHandler()
	if err != nil {
		return err
	}

	for _, appID := range appIDs {
		// get bootstrap user
		ctx, cancel := handler.GetTimeoutContext()
		defer cancel()
		bootstrapUser, _, err := client.Application.GetApplicationUser(ctx, appID)
		if err != nil {
			return cmderrors.NewUserError(err)
		}

		auth := c8y.NewBasicAuthString(bootstrapUser.Tenant, bootstrapUser.Username, bootstrapUser.Password)
		bootstrapCtx := context.WithValue(context.Background(), c8y.GetContextAuthTokenKey(), auth)

		_, resp, err := client.Application.GetCurrentApplicationSubscriptions(bootstrapCtx)
		if _, err := handler.ProcessResponse(resp, err, nil, commonOptions); err != nil {
			return err
		}
	}

	return nil
}
