// v2-based firmware version create: uploads a binary file (multipart) or links an
// external URL as a new version of a firmware package via Versions.Create (which
// uploads the binary and links the child addition internally). The firmware flag
// drives iteration (pipe or --firmware) and is resolved to an id via
// Repository.Firmware.ResolveID.
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
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/firmware/firmwareversions"
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

// NewCreateCmd creates a command to Create firmware package version
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create firmware package version",
		Long:  `Create a new firmware package version (managedObject)`,
		Example: heredoc.Doc(`
$ c8y firmware versions create --firmware "linux-os1" --version "1.0.0" --file "./python3.deb"
Create a new version using a binary file and link it to the existing "linux-os1" firmware. The binary will be uploaded to Cumulocity

$ c8y firmware versions create --firmware "linux-os1" --version "1.0.0" --url "https://blob.azure.com/device-firmare/1.0.0/image.mender"
Create a new version with an external URL and link it to the existing "linux-os1" firmware

$ c8y firmware versions create --firmware 12345
Create a new version with an empty version number and url
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("firmware", "", "Firmware package id where the version will be added to (accepts pipeline)")
	cmd.Flags().String("version", "", "Version name, i.e. 1.0.0. If left blank than the version")
	cmd.Flags().String("url", "", "URL to the firmware package")
	cmd.Flags().String("file", "", "File to be uploaded")

	completion.WithOptions(
		cmd,
		completion.WithFirmware("firmware", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("firmware", "firmware", false, "additionParents.references.0.managedObject.id", "id"),
		flags.WithPowershellName("New-FirmwareVersion"),
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

	if err := r.InputFlag("firmware"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		fwID, err := client.Repository.Firmware.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("firmware")), nil)
		if err != nil {
			return nil, err
		}

		// Version defaults to the version extracted from the file name (matching v1).
		file := in.String("file")
		version := in.String("version")
		if version == "" && file != "" {
			version = c8ydata.ExtractVersion(file)
		}

		opt := firmwareversions.CreateOptions{
			Version: version,
			URL:     in.String("url"),
		}
		if file != "" {
			opt.File = firmwareversions.UploadFileOptions{FilePath: file}
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitUpload(ctx, http.MethodPost, fwID, func(ctx context.Context) op.Result[jsonmodels.FirmwareVersion] {
				return client.Repository.Firmware.Versions.Create(ctx, fwID, opt)
			})
		}, nil
	})
}
