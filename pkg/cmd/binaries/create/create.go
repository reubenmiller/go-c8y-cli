// v2-based binary create: uploads a file (multipart) to the inventory via
// Binaries.Create. --name/--type set the binary metadata and --data/--template
// add custom properties to the binary's managed object. Runs once (no pipeline).
package create

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/binaries"
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

// NewCreateCmd creates a command to Create binary
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create binary",
		Long:  `Create/upload a new binary to Cumulocity`,
		Example: heredoc.Doc(`
$ c8y binaries create --file ./myfile.log
Upload a log file

$ c8y binaries create --file "myConfig.json" --type c8y_upload --data "c8y_Global={}"
Upload a config file and make it globally accessible for all users
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("file", "", "File to be uploaded as a binary (required)")
	cmd.Flags().String("name", "", "Set the name of the binary file. This will be the name of the file when it is downloaded in the UI")
	cmd.Flags().String("type", "", "Custom type. If left blank, the MIME type will be detected from the file extension")
	_ = cmd.MarkFlagRequired("file")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("", "", false),
		flags.WithPowershellName("New-Binary"),
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

	// Binaries create has no pipeline target (a single file is uploaded), so it
	// runs exactly once and must NOT read stdin — calling r.Input() here would
	// block waiting for piped input that never comes.

	// --data/--template build the custom object properties (not a request body
	// here — they are merged into the multipart object metadata).
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
		file := in.String("file")
		if file == "" {
			return nil, fmt.Errorf("--file is required")
		}
		var properties map[string]any
		if body, err := in.Body(); err == nil && len(body) > 0 {
			_ = json.Unmarshal(body, &properties)
		}
		opt := binaries.UploadFileOptions{
			FilePath:    file,
			Name:        in.String("name"),
			ContentType: in.String("type"),
			Properties:  properties,
		}
		return func(ctx context.Context) output.Seq {
			// SubmitUpload (not Submit): the multipart body can't be prepared for
			// the confirmation prompt without blocking.
			return c8ystream.SubmitUpload(ctx, http.MethodPost, "", func(ctx context.Context) op.Result[jsonmodels.Binary] {
				return client.Binaries.Create(ctx, opt)
			})
		}, nil
	})
}
