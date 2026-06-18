// v2-based software version create: uploads a binary file (multipart) or links an
// external URL as a new version of a software package via Versions.Create (which
// uploads the binary and links the child addition internally). The software flag
// drives iteration (pipe or --software) and is resolved to an id via
// Repository.Software.ResolveID.
package create

import (
	"context"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/software/softwareversions"
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

// NewCreateCmd creates a command to Create software package version
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create software package version",
		Long:  `Create a new software package version (managedObject)`,
		Example: heredoc.Doc(`
$ c8y software versions create --software "my-app" --version "1.0.0" --file "./python3.deb"
Create a new version using a binary file. The binary will be uploaded to Cumulocity

$ c8y software versions create --software "my-app" --version "1.0.0" --url "https://"
Create a new version with an external URL

$ c8y software versions create --software 12345
Create a new version with an empty version number and url
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("software", "", "Software package id where the version will be added to (accepts pipeline)")
	cmd.Flags().String("version", "", "Software package version name, i.e. 1.0.0")
	cmd.Flags().String("url", "", "URL to the software package")
	cmd.Flags().String("file", "", "File to be uploaded")

	completion.WithOptions(
		cmd,
		completion.WithSoftware("software", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("software", "software", false, "additionParents.references.0.managedObject.id", "id"),
		flags.WithPowershellName("New-SoftwareVersion"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObject+json", ""),
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

	if err := r.InputFlag("software"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		swID, err := client.Repository.Software.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("software")), nil)
		if err != nil {
			return nil, err
		}

		// Version defaults to the version extracted from the file name (matching v1).
		file := in.String("file")
		version := in.String("version")
		if version == "" && file != "" {
			version = c8ydata.ExtractVersion(file)
		}

		opt := softwareversions.CreateOptions{
			Version: version,
			URL:     in.String("url"),
		}
		if file != "" {
			opt.File = softwareversions.UploadFileOptions{FilePath: file}
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitUpload(ctx, http.MethodPost, swID, func(ctx context.Context) op.Result[jsonmodels.SoftwareVersion] {
				return client.Repository.Software.Versions.Create(ctx, swID, opt)
			})
		}, nil
	})
}
