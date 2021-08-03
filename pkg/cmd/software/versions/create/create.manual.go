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
			$ c8y software create --version "1.0.0" --file "./python3.deb"
			Create a new version using a binary file. The binary will be uploaded to Cumulocity

			$ c8y software create --version "1.0.0" --url "https://"
			Create a new version with an external URL
			`),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("softwareId", []string{""}, "Software package id where the version will be added to (accepts pipeline)")
	cmd.Flags().String("version", "", "Software package version name, i.e. 1.0.0")
	cmd.Flags().String("url", "", "URL to the software package")
	cmd.Flags().String("file", "", "File to be uploaded")

	completion.WithOptions(
		cmd,
		completion.WithSoftware("softwareId", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("softwareId", "softwareId", false, "additionParents.references.0.managedObject.id", "id"),
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
	body := mapbuilder.NewInitializedMapBuilder()
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
		flags.WithDataFlagValue(),
		flags.WithStringValue("version", "c8y_Software.version"),
		flags.WithStringValue("url", "c8y_Software.url"),
		// flags.WithBinaryUploadURL(client, "file", "c8y_Software.url"),
		flags.WithDefaultTemplateString(`
{type: 'c8y_SoftwareBinary', c8y_Global:{}}`),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("type"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("{softwareId}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithSoftwareByNameFirstMatch(client, args, "softwareId", "softwareId"),
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
		softwareID, _, err := path.Execute(false)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if filename == "" {
			_, resp, respErr = client.Inventory.CreateChildAddition(context.Background(), softwareID, body)
		} else {
			_, resp, respErr = client.Inventory.CreateChildAdditionWithBinary(
				context.Background(), softwareID, filename, func(binaryURL string) interface{} {
					_ = body.Set("c8y_Software.url", binaryURL)
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
