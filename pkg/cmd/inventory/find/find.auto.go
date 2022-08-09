// Code generated from specification version 1.0.0: DO NOT EDIT
package find

import (
	"fmt"
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// FindCmd command
type FindCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewFindCmd creates a command to Find managed object collection
func NewFindCmd(f *cmdutil.Factory) *FindCmd {
	ccmd := &FindCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find managed object collection",
		Long:  `Get a collection of managedObjects based on the Cumulocity query language`,
		Example: heredoc.Doc(`
$ c8y inventory find --query "name eq 'roomUpperFloor_*'"
Get a list of managed objects

$ echo "myname" | c8y inventory find --queryTemplate "name eq '*%s*'"
Find managed objects which include myname in their names. query=$filter=name eq '*myname*'

$ echo "name eq 'name'" | c8y inventory find --queryTemplate 'not(%s)'
Invert a given query received via piped input (stdin) by using a template
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("query", "", "ManagedObject query (required) (accepts pipeline)")
	cmd.Flags().String("queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
	cmd.Flags().String("orderBy", "", "Order by. e.g. _id asc or name asc or creationTime.date desc")
	cmd.Flags().Bool("onlyDevices", false, "Only include devices (deprecated)")
	cmd.Flags().Bool("withParents", false, "include a flat list of all parents and grandparents of the given object")

	completion.WithOptions(
		cmd,
	)

	flags.WithOptions(
		cmd,

		flags.WithExtendedPipelineSupport("query", "query", true, "c8y_DeviceQueryString"),
		flags.WithCollectionProperty("managedObjects"),
	)

	// Required flags

	_ = cmd.Flags().MarkHidden("onlyDevices")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *FindCmd) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithBoolValue("withParents", "withParents", ""),

		flags.WithCumulocityQuery(
			[]flags.GetOption{
				flags.WithStringValue("query", "query", "(%s)"),
				flags.WithDefaultBoolValue("onlyDevices", "onlyDevices", "has(c8y_IsDevice)"),
			},
			"query",
		),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(query)

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
	body := mapbuilder.NewInitializedMapBuilder(false)
	err = flags.WithBody(
		cmd,
		body,
		inputIterators,
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		Method:       "GET",
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
