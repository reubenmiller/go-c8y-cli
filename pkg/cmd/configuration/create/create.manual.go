// Code generated from specification version 1.0.0: DO NOT EDIT
package create

import (
	"context"
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
)

// CreateCmd command
type CreateCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateCmd creates a command to Create configuration file
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create configuration file",
		Long:  `Create a new configuration file (managedObject)`,
		Example: heredoc.Doc(`
			$ c8y configuration create --name myExampleConfig01 --configurationType EXAMPLE --file myconfig.json
			Create a configuration from a file

			$ c8y configuration create --name "agent config" --description "Default agent configuration" --configurationType "agentConfig" --url "https://example.com/myfile.json"
			Create a configuration using an external URL
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "name (accepts pipeline)")
	cmd.Flags().String("description", "", "Description of the configuration package")
	cmd.Flags().String("configurationType", "", "Configuration type")
	cmd.Flags().String("url", "", "URL link to the configuration file")
	cmd.Flags().String("deviceType", "", "Device type filter. Only allow configuration to be applied to devices of this type")
	cmd.Flags().String("file", "", "File to be uploaded")

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("name", "name", false, "name"),
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
		flags.WithOverrideValue("name", "name"),
		flags.WithDataFlagValue(),
		flags.WithStringValue("name", "name"),
		flags.WithStringValue("description", "description"),
		flags.WithStringValue("configurationType", "configurationType"),
		flags.WithStringValue("url", "url"),
		flags.WithStringValue("deviceType", "deviceType"),
		flags.WithDefaultTemplateString(`
{type: 'c8y_ConfigurationDump', c8y_Global:{}}`),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("type", "name"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
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
		bodyJSON, err := body.MarshalJSON()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if filename != "" {
			_, resp, respErr = client.Inventory.CreateWithBinary(context.Background(), filename, func(binaryURL string) interface{} {
				// Merge url into body using bytes as the body has already been marshalled to JSON
				bodyUpdated, _ := sjson.SetBytes(bodyJSON, "url", binaryURL)
				return bodyUpdated
			})
		} else {
			_, resp, respErr = client.Inventory.Create(context.Background(), bodyJSON)
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
