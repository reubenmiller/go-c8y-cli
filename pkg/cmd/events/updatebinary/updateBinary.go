// v2-based event binary update: the id flag drives iteration (pipe or --id);
// the event's binary content is replaced from --file via Events.Binaries.Update.
package updatebinary

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

// UpdateBinaryCmd command
type UpdateBinaryCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewUpdateBinaryCmd creates a command to Update event binary
func NewUpdateBinaryCmd(f *cmdutil.Factory) *UpdateBinaryCmd {
	ccmd := &UpdateBinaryCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "updateBinary",
		Short: "Update event binary",
		Long:  `Update an existing event binary`,
		Example: heredoc.Doc(`
$ c8y events updateBinary --id 12345 --file ./myfile.log
Update a binary related to an event
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.UpdateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Event id (required) (accepts pipeline)")
	cmd.Flags().String("file", "", "File to be uploaded as a binary (required)")
	_ = cmd.MarkFlagRequired("file")

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Update-EventBinary"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.event+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *UpdateBinaryCmd) RunE(cmd *cobra.Command, args []string) error {
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
		opt := eventbinaries.UploadFileOptions{FilePath: file}
		return func(ctx context.Context) output.Seq {
			// SubmitUpload (not Submit): the upload body can't be prepared for the
			// confirmation prompt without blocking.
			return c8ystream.SubmitUpload(ctx, http.MethodPut, id, func(ctx context.Context) op.Result[jsonmodels.EventBinary] {
				return client.Events.Binaries.Update(ctx, id, opt)
			})
		}, nil
	})
}
