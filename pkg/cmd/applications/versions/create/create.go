// v2-based application version create: uploads a ZIP (multipart) as a new version
// of an application via ApplicationVersions.CreateFromFile (which accepts a local
// path or an http(s) URL). The application flag drives iteration (pipe or
// --application); the application reference (id or name) is resolved via
// Applications.ResolveID.
package create

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
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

// NewCreateCmd creates a command to Create application version
func NewCreateCmd(f *cmdutil.Factory) *CreateCmd {
	ccmd := &CreateCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create application version",
		Long:  `Uploaded version and tags can only contain upper and lower case letters, integers and ., +, -. Other characters are prohibited.`,
		Example: heredoc.Doc(`
$ c8y applications versions create --application 1234 --file "./testdata/myapp.zip" --version "2.0.0" --tags latest
Create a new application version
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("application", "", "Application (accepts pipeline)")
	cmd.Flags().String("file", "", "The ZIP file to be uploaded (a local path or an http(s) URL)")
	cmd.Flags().String("version", "", "The version name (required)")
	cmd.Flags().StringSlice("tags", []string{""}, "Tags assigned to the version. Version tags must be unique across all versions and version fields of application versions (required)")

	completion.WithOptions(
		cmd,
		completion.WithApplicationWithVersions("application", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithExtendedPipelineSupport("application", "application", false, "id", "name"),
		flags.WithCollectionProperty("-"),
		flags.WithPowershellName("New-ApplicationVersion"),
		flags.WithOutputType("application/json", ""),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("version")
	_ = cmd.MarkFlagRequired("tags")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CreateCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("application"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		appID, err := client.Applications.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("application")), nil)
		if err != nil {
			return nil, err
		}
		file := in.String("file")
		if file == "" {
			return nil, fmt.Errorf("--file is required")
		}
		version := in.String("version")
		tags := in.StringSlice("tags")
		return func(ctx context.Context) output.Seq {
			// SubmitUpload (not Submit): the multipart body can't be prepared for
			// the confirmation prompt without blocking. CreateFromFile opens the
			// local path / downloads the URL when the call runs.
			return c8ystream.SubmitUpload(ctx, http.MethodPost, "", func(ctx context.Context) op.Result[jsonmodels.ApplicationVersion] {
				return client.ApplicationVersions.CreateFromFile(ctx, appID, file, version, tags)
			})
		}, nil
	})
}
