package createhostedapplication

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/applications"
	"github.com/spf13/cobra"
)

type CmdCreateHostedApplication struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory

	file           string
	name           string
	key            string
	availability   string
	contextPath    string
	resourcesURL   string
	skipActivation bool
	skipUpload     bool
}

func NewCmdCreateHostedApplication(f *cmdutil.Factory) *CmdCreateHostedApplication {
	ccmd := &CmdCreateHostedApplication{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "createHostedApplication",
		Short: "Create hosted application",
		Long: heredoc.Doc(`
			Create a new hosted web application or update the binary of an existing hosted application

			If the zip file or folder contains a 'cumulocity.json' manifest file, then the key will be automatically read from it.
		`),
		Example: heredoc.Doc(`
			$ c8y applications createHostedApplication --file ./myapp.zip
			Create new hosted application from a given zip file. The application will be called "myapp". If the application placeholder does not exist then it will be created

			$ c8y applications createHostedApplication --file simple-helloworld/build --name custom_helloworld
			Create/update hosted web application from a build folder and specify a custom application name

			$ c8y applications createHostedApplication --file myapp.zip --skipActivation
			Create/update hosted web application but don't activate it, so the current version (if any) will be untouched
		`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.file, "file", "", "File or Folder of the web application. It should contain a index.html file in the root folder/ or zip file")
	cmd.Flags().StringVar(&ccmd.name, "name", "", "Name of application")
	cmd.Flags().StringVar(&ccmd.key, "key", "", "Shared secret of application. Defaults to the value inside the cumulocity.json file (if present)")
	cmd.Flags().StringVar(&ccmd.availability, "availability", "", "Access level for other tenants. Possible values are : MARKET, PRIVATE (default)")
	cmd.Flags().StringVar(&ccmd.contextPath, "contextPath", "", "contextPath of the hosted application")
	cmd.Flags().StringVar(&ccmd.resourcesURL, "resourcesUrl", "/", "URL to application base directory hosted on an external server. Required when application type is HOSTED")

	cmd.Flags().BoolVar(&ccmd.skipActivation, "skipActivation", false, "Don't activate to the application after it has been created and uploaded")
	cmd.Flags().BoolVar(&ccmd.skipUpload, "skipUpload", false, "Don't uploaded the web app binary. Only the application placeholder will be created")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("availability", "MARKET", "PRIVATE"),
	)

	flags.WithOptions(
		cmd,
		flags.WithData(),
		f.WithTemplateFlag(cmd),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE creates the hosted application (or finds the existing one with the
// name), uploads the web application and activates it. go-c8y's DeployHosted
// does all of it; the name, key and context path default to the manifest
// (cumulocity.json) and then to the file name.
func (n *CmdCreateHostedApplication) RunE(cmd *cobra.Command, args []string) error {
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
		opt := applications.DeployHostedOptions{
			// Only the tenant's own applications are updated, never a
			// subscribed one with the same name (e.g. cockpit).
			Owner:          client.Auth.Tenant,
			Path:           path,
			Name:           n.name,
			Key:            n.key,
			ContextPath:    n.contextPath,
			Availability:   n.availability,
			ResourcesURL:   n.resourcesURL,
			SkipUpload:     n.skipUpload || path == "",
			SkipActivation: n.skipActivation,
		}
		return func(ctx context.Context) any {
			return client.Applications.DeployHosted(ctx, opt)
		}, nil
	})
}
