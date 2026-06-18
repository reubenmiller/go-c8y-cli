// v2-based event binary create: the id flag drives iteration (pipe or --id);
// each event gets a binary uploaded (multipart) from --file via
// Events.Binaries.Create. --name/--type set the attachment's file name and MIME
// type in the multipart object metadata.
package createbinary

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/events/eventbinaries"
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

// NewCreateBinaryCmd creates a command to Create event binary
func NewCreateBinaryCmd(f *cmdutil.Factory) *CreateBinaryCmd {
	ccmd := &CreateBinaryCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "createBinary",
		Short: "Create event binary",
		Long:  `Upload a new binary file to an event`,
		Example: heredoc.Doc(`
$ c8y events createBinary --id 12345 --file ./myfile.log
Add a binary to an event

$ c8y events createBinary --id 12345 --file ./myfile.log --name "myfile-2022-03-31.txt"
Add a binary to an event using a custom name and use an auto-detected mime-type

$ c8y events createBinary --id 12345 --file ./myfile.log --name "example.bin" --type "application/octet-stream"
Add a binary to an event using a custom name and use an explicit mime-type
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Event id (required) (accepts pipeline)")
	cmd.Flags().String("file", "", "File to be uploaded as a binary (required)")
	cmd.Flags().String("name", "", "Set the name of the binary file. This will be the name of the file when it is downloaded in the UI")
	cmd.Flags().String("type", "", "Set the MIME type of the binary file, e.g. text/plain. If left blank, the type will be detected from the file extension or its contents")
	_ = cmd.MarkFlagRequired("file")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("New-EventBinary"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.event+json", ""),
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

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		id := in.String("id")
		file := in.String("file")
		if file == "" {
			return nil, fmt.Errorf("--file is required")
		}
		opt := eventbinaries.UploadFileOptions{
			FilePath:    file,
			Name:        in.String("name"),
			ContentType: in.String("type"),
		}
		return func(ctx context.Context) output.Seq {
			// SubmitUpload (not Submit): the multipart body can't be prepared for
			// the confirmation prompt without blocking.
			return c8ystream.SubmitUpload(ctx, http.MethodPost, id, func(ctx context.Context) op.Result[jsonmodels.EventBinary] {
				return client.Events.Binaries.Create(ctx, id, opt)
			})
		}, nil
	})
}
