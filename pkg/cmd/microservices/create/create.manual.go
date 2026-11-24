package create

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/microservices"
	"github.com/spf13/cobra"
)

type CmdCreate struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	file             string
	name             string
	key              string
	availability     string
	contextPath      string
	resourceURL      string
	skipSubscription bool
	skipUpload       bool
}

func NewCmdCreate(f *cmdutil.Factory) *CmdCreate {
	ccmd := &CmdCreate{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create microservice",
		Long: heredoc.Doc(`Create a new microservice or update the application binary of an existing microservice

			Note: Named lookups of microservices will only include microservices which 
		`),
		Example: heredoc.Doc(`
$ c8y microservices create --file ./myapp.zip
Create new microservice

$ c8y microservices create --name my-application --file ./myapp.zip
Create or update a microservice using an explicit name

$ c8y microservices create --file ./manifest/cumulocity.json
Create or update an existing microservice using it's manifest file. This will set the requiredRoles and roles defined in the given manifest file
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.file, "file", "", "Microservice file to be uploaded (or Cumulocity.json) file")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of application")
	cmd.Flags().StringVar(&ccmd.key, "key", "", "Shared secret of application")
	cmd.Flags().StringVar(&ccmd.availability, "availability", "", "Access level for other tenants. Possible values are : MARKET, PRIVATE (default)")
	cmd.Flags().StringVar(&ccmd.contextPath, "contextPath", "", "contextPath of the hosted application. Required when application type is HOSTED")
	cmd.Flags().StringVar(&ccmd.resourceURL, "resourcesUrl", "", "URL to application base directory hosted on an external server. Required when application type is HOSTED")

	cmd.Flags().BoolVar(&ccmd.skipSubscription, "skipSubscription", false, "Skip microservice subscription when creating the new microservice")
	cmd.Flags().BoolVar(&ccmd.skipUpload, "skipUpload", false, "Skip uploading the binary to the platform")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd).SetRequiredFlags("file")

	return ccmd
}

// RunE creates the microservice (or finds the existing one with the name),
// uploads the zip file (or sets the roles from a manifest) and subscribes the
// tenant. go-c8y's Microservices.Deploy does all of it.
func (n *CmdCreate) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		path, err := in.File("file")
		if err != nil {
			return nil, err
		}
		opt := microservices.CreateOptions{
			File:             path,
			Name:             n.name,
			Key:              n.key,
			Availability:     n.availability,
			ContextPath:      n.contextPath,
			ResourceURL:      n.resourceURL,
			SkipSubscription: n.skipSubscription,
			SkipUpload:       n.skipUpload,
		}
		return func(ctx context.Context) any {
			return client.Microservices.Deploy(ctx, opt)
		}, nil
	})
}
