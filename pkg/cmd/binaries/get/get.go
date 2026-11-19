// v2-based binary download: the id flag drives iteration (pipe or --id); each
// binary is fetched via Binaries.Get and its content streamed to the
// --outputFileRaw path (with {id}/{filename} placeholders) or, when that is
// unset, to stdout. The command yields no document — the binary IS the output —
// so the runner's --outputFileRaw stage doesn't overwrite the file we wrote.
package get

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

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Download binary
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Download binary",
		Long: heredoc.Doc(`Download a binary stored in Cumulocity and display it on the console.

For non text based binaries or if the output should be saved to file, the --outputFileRaw global parameter should be used, which supports the following placeholders:

* {filename} - Filename found in the Content-Disposition response header
* {id} - The binary id
`),
		Example: heredoc.Doc(`
$ c8y binaries get --id 12345
Get a binary and display the contents on the console

$ c8y binaries get --id 12345 --outputFileRaw "./download-binary1.txt"
Get a binary and save it to a file

$ c8y binaries list | c8y binaries get --outputFileRaw "download-{id}-{filename}" > /dev/null
Download a list of binaries and save the files to the current working directory
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "Inventory binary id (required) (accepts pipeline)")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true, "id", "binaryId"),
		flags.WithPowershellName("Get-Binary"),
		flags.WithOutputType("*/*", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetCmd) RunE(cmd *cobra.Command, args []string) error {
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
				res := client.Binaries.Get(ctx, id)
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
