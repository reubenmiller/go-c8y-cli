// Code generated from specification version 1.0.0: DO NOT EDIT
package createpatch

import (
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

// CreatePatchCmd command
type CreatePatchCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreatePatchCmd creates a command to Create firmware package version patch
func NewCreatePatchCmd(f *cmdutil.Factory) *CreatePatchCmd {
	ccmd := &CreatePatchCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "createPatch",
		Short: "Create firmware package version patch",
		Long:  `Create a new firmware package (managedObject)`,
		Example: heredoc.Doc(`
$ c8y firmware create --name "python3-requests" --description "python requests library"
Create a new version to an existing firmware package
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("firmwareId", []string{""}, "Firmware package id where the version will be added to (accepts pipeline)")
	cmd.Flags().String("version", "", "Patch version, i.e. 1.0.0")
	cmd.Flags().String("url", "", "URL to the firmware patch")
	cmd.Flags().String("dependencyVersion", "", "Existing firmware version that the patch is dependent on")

	completion.WithOptions(
		cmd,
		completion.WithFirmwareVersion("firmwareId", "firmwareId", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("firmwareId", "firmwareId", false, "additionParents.references.0.managedObject.id", "id"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreatePatchCmd) RunE(cmd *cobra.Command, args []string) error {
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

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
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
		flags.WithStringValue("version", "c8y_Firmware.version"),
		flags.WithStringValue("url", "c8y_Firmware.url"),
		flags.WithStringValue("dependencyVersion", "c8y_Patch.dependency"),
		flags.WithDefaultTemplateString(`
{type: 'c8y_FirmwareBinary', c8y_Global:{}}`),
		cmdutil.WithTemplateValue(cfg),
		flags.WithTemplateVariablesValue(),
		flags.WithRequiredProperties("type", "c8y_Patch.dependency"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects/{firmwareId}/childAdditions")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithFirmwareVersionByNameFirstMatch(client, args, "firmwareId", "firmwareId"),
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "POST",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
