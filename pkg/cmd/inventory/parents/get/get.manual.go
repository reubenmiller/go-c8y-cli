package get

import (
	"fmt"
	"io"
	"net/http"
	"strings"

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

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	Type    string
	Level   int64
	All     bool
	Reverse bool

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get addition parent
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get addition parent",
		Long:  `Get addition parent`,
		Example: heredoc.Doc(`
			$ c8y inventory parents get --id 12345 --type addition
			Get parent of the child addition

			$ c8y inventory parents get --id 12345 --type asset
			Get parent of the child asset (e.g. usually the group)

			$ c8y devices list | c8y inventory parents get --type device --level 2
			Get the grandparent (parent of the parent)

			$ c8y devices list | c8y inventory parents get --type device --level -1
			Get the root parent device of a list of devices

			$ c8y devices get --id 12345 | c8y inventory parents get --type device -all
			Get all parents (parent, grandparents, grandparents etc.) of a single device returning the root parent last

			$ c8y devices get --id 12345 | c8y inventory parents get --type device -all --reverse
			Get all parents (parent, grandparents, grandparents etc.) of a single device returning the root parent first
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "ManagedObject id (required) (accepts pipeline)")
	cmd.Flags().BoolVar(&ccmd.All, "all", false, "Return all parents in the chain")
	cmd.Flags().BoolVar(&ccmd.Reverse, "reverse", false, "Return all parents in order from root to parent")
	cmd.Flags().StringVar(&ccmd.Type, "type", "", "Type of relationship, e.g. addition, asset, device")
	cmd.Flags().Int64Var(&ccmd.Level, "level", 1, "Number of parent jumps to do. 0 = current item, 1 = parent, 2 = grandparent etc. Defaults to 1. Use -1 for last parent")
	cmd.Flags().Bool("withParents", true, "include a flat list of all parents and grandparents of the given object")

	// Hide withParents as it is fixed
	cmd.Flags().MarkHidden("withParents")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("type", "addition", "asset", "device"),
	)

	flags.WithOptions(
		cmd,

		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetCmd) RunE(cmd *cobra.Command, args []string) error {
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

	if n.Type == "" {
		return &flags.ParameterError{
			Name: "type",
			Err:  flags.ErrParameterMissing,
		}
	}

	// Change collection based on options

	property := ""

	if n.All {
		if n.Reverse {
			property = fmt.Sprintf("%sParents.references|@reverse|#.managedObject", strings.ToLower(n.Type))
		} else {
			property = fmt.Sprintf("%sParents.references.#.managedObject", strings.ToLower(n.Type))
		}
	} else {
		if n.Level == 0 {
			property = ""
		} else if n.Level > 0 {
			property = fmt.Sprintf("%sParents.references.%d.managedObject", strings.ToLower(n.Type), n.Level-1)
		} else {
			// Support python style indexing (-1 = last index, -2 = second last)
			property = fmt.Sprintf("%sParents.references|@reverse|%d.managedObject", strings.ToLower(n.Type), (-1*n.Level)-1)
		}
	}

	flags.WithOptions(
		cmd,
		flags.WithCollectionProperty(property),
	)

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
		flags.WithDefaultBoolValue("withParents", "withParents", ""),
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
	path := flags.NewStringTemplate("inventory/managedObjects/{id}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		c8yfetcher.WithIDSlice(args, "id", "id"),
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
