// Code generated from specification version 1.0.0: DO NOT EDIT
package assigngroup

import (
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// AssignGroupCmd command
type AssignGroupCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewAssignGroupCmd creates a command to Assign child group
func NewAssignGroupCmd(f *cmdutil.Factory) *AssignGroupCmd {
	ccmd := &AssignGroupCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "assignGroup",
		Short:  "Assign child group",
		Long:   `Assigns a group to a group. The group will be a childAsset of the group`,
		Hidden: true,

		Example: heredoc.Doc(`
$ c8y devicegroups assignGroup --group 12345 --newChildGroup 43234
Add a group to a group

$ c8y devicegroups assignGroup --group 12345 --newChildGroup 43234,99292,12222
Add multiple groups to a group
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled()
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("group", []string{""}, "Group (required)")
	cmd.Flags().StringSlice("newChildGroup", []string{""}, "New child group to be added to the group as an child asset (required) (accepts pipeline)")

	completion.WithOptions(
		cmd,
		completion.WithDeviceGroup("group", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDeviceGroup("newChildGroup", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),

		flags.WithExtendedPipelineSupport("newChildGroup", "managedObject.id", true, "id"),
		flags.WithPipelineAliases("group", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("newChildGroup", "source.id", "managedObject.id", "id"),

		flags.WithCollectionProperty("managedObject"),
		flags.WithDeprecationNotice("please use 'c8y devicegroups children unassign --childType asset' instead"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("group")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *AssignGroupCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	// Runtime flag options
	flags.WithOptions(
		cmd,
		flags.WithRuntimePipelineProperty(),
	)
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
		flags.WithDataFlagValue(),
		c8yfetcher.WithDeviceGroupByNameFirstMatch(n.factory, args, "newChildGroup", "managedObject.id"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects/{id}/childAssets")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithDeviceGroupByNameFirstMatch(n.factory, args, "group", "id"),
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
