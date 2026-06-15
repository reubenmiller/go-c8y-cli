// v2-based application binary create: uploads a file (multipart) to an
// application via Applications.Upload. The id flag drives iteration (pipe or
// --id); the application reference (id or name) is resolved internally by the
// SDK. --data/--template add custom properties to the multipart metadata.
package createbinary

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/applications"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// CreateBinaryCmd command
type CreateBinaryCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCreateBinaryCmd creates a command to Create application binary
func NewCreateBinaryCmd(f *cmdutil.Factory) *CreateBinaryCmd {
	ccmd := &CreateBinaryCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "createBinary",
		Short: "Create application binary",
		Long:  `Upload an application binary (e.g. a ZIP) for a registered hosted application`,
		Example: heredoc.Doc(`
$ c8y applications createBinary --id 12345 --file ./myapp.zip
Upload an application binary
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("id", "", "Application id (required) (accepts pipeline)")
	cmd.Flags().String("file", "", "File to be uploaded as a binary (required)")
	_ = cmd.MarkFlagRequired("file")

	completion.WithOptions(
		cmd,
		completion.WithApplication("id", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPipelineAliases("id", "id"),
		flags.WithPowershellName("New-ApplicationBinary"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObject+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateBinaryCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("id"); err != nil {
		return err
	}

	// --data/--template build the custom multipart object metadata.
	err = r.Body(
		flags.WithDataFlagValue(),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		ref := c8ystream.NameOrID(in.String("id"))
		file := in.String("file")
		if file == "" {
			return nil, fmt.Errorf("--file is required")
		}
		var properties map[string]any
		if body, err := in.Body(); err == nil && len(body) > 0 {
			_ = json.Unmarshal(body, &properties)
		}
		opt := applications.UploadFileOptions{FilePath: file, Properties: properties}
		return func(ctx context.Context) output.Seq {
			return c8ystream.SubmitUpload(ctx, http.MethodPost, "", func(ctx context.Context) op.Result[jsonmodels.Application] {
				return client.Applications.Upload(ctx, ref, opt)
			})
		}, nil
	})
}
