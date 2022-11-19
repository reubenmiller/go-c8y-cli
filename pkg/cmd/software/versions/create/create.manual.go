package create

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/pkg/c8y/binary"
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
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("software", []string{""}, "Software package id where the version will be added to (accepts pipeline)")
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
		flags.WithStaticStringValue("c8y_Software.version", ""),
		flags.WithStaticStringValue("c8y_Software.url", ""),
		flags.WithDataFlagValue(),
		flags.WithVersion("file", "version", "c8y_Software.version"),
		flags.WithStringValue("url", "c8y_Software.url"),
		flags.WithDefaultTemplateString(`
{type: 'c8y_SoftwareBinary', c8y_Global:{}}`),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("type", "c8y_Software.version"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("{software}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithSoftwareByNameFirstMatch(client, args, "software", "software"),
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
			file, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer file.Close()

			progress := n.factory.IOStreams.ProgressIndicator()
			bar, err := c8ybinary.NewProgressBar(progress, filename)

			if err != nil {
				return err
			}

			binaryFile, err := binary.NewBinaryFile(
				binary.WithReader(file),
				binary.WithFileProperties(filename),
				binary.WithGlobal(),
			)

			if err != nil {
				return err
			}
			_, resp, respErr = client.Inventory.CreateChildAdditionWithBinary(
				context.Background(), softwareID, binaryFile, func(binaryURL string) interface{} {
					_ = body.Set("c8y_Software.url", binaryURL)
					return body
				}, func(r *http.Request) (*http.Request, error) {
					if bar != nil {
						r.Body = bar.ProxyReader(r.Body)
					}
					return r, nil
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
