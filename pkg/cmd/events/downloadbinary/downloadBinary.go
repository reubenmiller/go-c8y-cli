// v2-based event binary download: the id flag drives iteration (pipe or --id);
// each event's binary is fetched via Events.Binaries.Get and its content
// streamed to the --outputFileRaw path (with {id}/{filename} placeholders) or,
// when that is unset, to stdout. The command yields no document — the binary IS
// the output — so the runner's --outputFileRaw stage doesn't overwrite the file
// we wrote.
package downloadbinary

import (
	"context"
	"io"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// DownloadBinaryCmd command
type DownloadBinaryCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewDownloadBinaryCmd creates a command to Get event binary
func NewDownloadBinaryCmd(f *cmdutil.Factory) *DownloadBinaryCmd {
	ccmd := &DownloadBinaryCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "downloadBinary",
		Short: "Get event binary",
		Long: heredoc.Doc(`Get the binary associated with an event

When downloading a binary it is useful to use the --outputFileRaw global parameter and to use one of the following placeholders:

* {filename} - Filename found in the Content-Disposition response header
* {id} - The event id
`),
		Example: heredoc.Doc(`
$ c8y events downloadBinary --id 12345 --outputFileRaw ./eventbinary.txt
Download a binary related to an event

$ c8y events list --fragmentType "c8y_IsBinary" | c8y events downloadBinary --outputFileRaw "./output/binary-{id}-{filename}" > /dev/null
Download a list of event binaries and use a template name to save each binary individually
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Event id (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true),
		flags.WithPowershellName("Get-EventBinary"),
		flags.WithOutputType("*/*", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *DownloadBinaryCmd) RunE(cmd *cobra.Command, args []string) error {
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

	// Command-level output config, captured once: where each binary is written.
	outputFileRaw := r.Config.GetOutputFileRaw()
	stdout := cmd.OutOrStdout()

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		id := in.String("id")
		return func(ctx context.Context) output.Seq {
			return func(yield func(jsondoc.JSONDoc, error) bool) {
				res := client.Events.Binaries.Get(ctx, id)
				if res.Err != nil {
					yield(jsondoc.Empty(), res.Err)
					return
				}
				// Under --dry the request has already been rendered; nothing to write.
				if apiv2.IsDryRun(ctx) {
					return
				}
				resp := res.Data
				defer resp.Close()
				if outputFileRaw != "" {
					fields := map[string]string{"id": id, "filename": resp.FileName()}
					if _, err := fileutilities.WriteToFile(resp.Reader(), outputFileRaw, false, false, fields); err != nil {
						yield(jsondoc.Empty(), err)
					}
					return
				}
				if _, err := io.Copy(stdout, resp.Reader()); err != nil {
					yield(jsondoc.Empty(), err)
				}
			}
		}, nil
	})
}
