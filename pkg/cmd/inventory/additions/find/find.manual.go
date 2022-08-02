package find

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8yfetcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/url"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdFind struct {
	*subcommand.SubCommand

	queryTemplate string
	orderBy       string
	factory       *cmdutil.Factory
}

func NewCmdFind(f *cmdutil.Factory) *CmdFind {
	ccmd := &CmdFind{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find child additions",
		Long:  `Retrieve all child additions of a specific managed object by a given ID, or a subset based on queries.`,
		Example: heredoc.Doc(`
			$ c8y inventory additions find --id 12345 --query "name eq 'roomUpperFloor_*'"
			Find child additions matching a specific name

			$ c8y inventory list | c8y inventory additions find --query "name eq 'roomUpperFloor_*'"
			Pipe a list of inventory items and find any child additions matching some criteria
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "ManagedObject id (required) (accepts pipeline)")
	cmd.Flags().String("query", "", "Use query language to perform operations and/or filter the results. Details about the properties and supported operations can be found in Query language")
	cmd.Flags().StringVar(&ccmd.queryTemplate, "queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
	cmd.Flags().StringVar(&ccmd.orderBy, "orderBy", "", "Order the results by the given parameter. i.e. 'id asc' or 'name desc'")
	cmd.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned when fetching the managed object.")
	cmd.Flags().Bool("withChildrenCount", false, "When set to true, the returned result will contain the total number of children in the respective objects (childAdditions, childAssets and childDevices)")
	cmd.Flags().Bool("withTotalElements", false, "When set to true, the returned result will contain in the statistics object the total number of elements. Only applicable on range queries")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("references.#.managedObject"),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdFind) RunE(cmd *cobra.Command, args []string) error {
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
		flags.WithCustomStringValue(func(b []byte) []byte {
			if n.queryTemplate != "" {
				b = []byte(fmt.Sprintf(n.queryTemplate, b))
			}

			// Encode special characters
			b = url.EscapeQuery(b)

			if n.orderBy != "" {
				if !bytes.Contains(b, []byte("$orderby=")) {
					b = append(b, []byte(" $orderby="+n.orderBy)...)
				}
			}

			if len(b) != 0 && !bytes.HasPrefix(b, []byte("$filter")) {
				b = append([]byte("$filter="), b...)
			}
			return b
		}, func() string {
			return "query"
		}, "query"),
		flags.WithBoolValue("withChildren", "withChildren"),
		flags.WithBoolValue("withChildrenCount", "withChildrenCount"),
		flags.WithBoolValue("withTotalElements", "withTotalElements"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
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

	// body
	body := mapbuilder.NewInitializedMapBuilder(false)

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects/{id}/childAdditions")
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
