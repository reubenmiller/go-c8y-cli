package find

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdFind struct {
	*subcommand.SubCommand

	onlyDevices   bool
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
		Short: "Find managed object collection",
		Long:  `Get a collection of managedObjects based on Cumulocity query language`,
		Example: heredoc.Doc(`
			$ c8y inventory find --query "name eq 'roomUpperFloor_*'"
			Get a list of managed objects

			$ echo "myname" | c8y inventory find --queryTemplate 'name eq '*%s*' --onlyDevices
			Find devices which include myname in their names. query=$filter=name eq '*myname*'

			$ echo "name eq 'name'" | c8y inventory find --queryTemplate 'not(%s)'
			Invert a given query received via piped input (stdin) by using a template
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("query", "", "ManagedObject query. (required) (accepts pipeline)")
	cmd.Flags().StringVar(&ccmd.queryTemplate, "queryTemplate", "", "String template to be used when applying the given query. Use %s to reference the query/pipeline input")
	cmd.Flags().StringVar(&ccmd.orderBy, "orderBy", "", "Order the results by the given parameter. i.e. 'id asc' or 'name desc'")
	cmd.Flags().Bool("withParents", false, "include a flat list of all parents and grandparents of the given object")
	cmd.Flags().BoolVar(&ccmd.onlyDevices, "onlyDevices", false, "Only include devices in the query")

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("query", "query", true, "c8y_DeviceQueryString"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("query")

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
	inputIterators, err := flags.NewRequestInputIterators(cmd)
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
			if n.onlyDevices {
				return "q"
			}
			return "query"
		}, "query"),
		flags.WithBoolValue("withParents", "withParents"),
	)
	if err != nil {
		return cmderrors.NewUserError(err)
	}

	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	commonOptions.ResultProperty = "managedObjects"
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
	body := mapbuilder.NewInitializedMapBuilder()

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects")

	req := c8y.RequestOptions{
		Method:       "GET",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: cfg.IgnoreAcceptHeader(),
		DryRun:       cfg.DryRun(),
	}

	return n.factory.RunWithWorkers(client, cmd, &req, inputIterators)
}
