package create

import (
	"context"
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
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
			`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("firmware", []string{""}, "Firmware package id where the version will be added to (accepts pipeline)")
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
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}
	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// headers
	headers := http.Header{}
	err = flags.WithHeaders(
		cmd,
		headers,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
		flags.WithStaticStringValue("Content-Type", "application/vnd.com.nsn.cumulocity.managedObject+json"),
		flags.WithProcessingModeValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// form data
	formData := make(map[string]io.Reader)
	err = flags.WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// body
	body := mapbuilder.NewInitializedMapBuilder(true)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		flags.WithVersion("file", "version", "c8y_Firmware.version"),
		flags.WithStringValue("url", "c8y_Firmware.url"),
		flags.WithDefaultTemplateString(`
{type: 'c8y_FirmwareBinary', c8y_Global:{}}`),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("type", "c8y_Firmware.version"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("{firmware}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithFirmwareByNameFirstMatch(client, args, "firmware", "firmware"),
	)
	if err != nil {
		return err
	}

	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	var filename string
	filename, err = cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	handler, _ := n.factory.GetRequestHandler()
	var resp *c8y.Response
	var respErr error
	bounded := inputIterators.Total > 0
	for {
		firmwareID, _, err := path.Execute(false)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if filename == "" {
			_, resp, respErr = client.Inventory.CreateChildAddition(context.Background(), firmwareID, body)
		} else {
			_, resp, respErr = client.Inventory.CreateChildAdditionWithBinary(
				context.Background(), firmwareID, filename, func(binaryURL string) interface{} {
					_ = body.Set("c8y_Firmware.url", binaryURL)
					return body
				})
		}

		if _, err := handler.ProcessResponse(resp, respErr, commonOptions); err != nil {
			return err
		}

		if !bounded {
			break
		}
	}

	return nil
}
