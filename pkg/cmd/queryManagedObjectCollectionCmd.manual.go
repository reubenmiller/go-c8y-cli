package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type QueryManagedObjectCollectionCmd struct {
	*subcommand.SubCommand
}

func NewQueryManagedObjectCollectionCmd() *QueryManagedObjectCollectionCmd {
	ccmd := &QueryManagedObjectCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find managed object collection",
		Long:  `Get a collection of managedObjects based on Cumulocity query language`,
		Example: heredoc.Doc(`
$ c8y inventory find --query "name eq 'roomUpperFloor_*'"
Get a list of managed objects
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("query", "", "ManagedObject query. (required)")
	cmd.Flags().String("orderBy", "", "Order the results by the given parameter. i.e. 'id asc'")
	cmd.Flags().Bool("withParents", false, "include a flat list of all parents and grandparents of the given object")

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport(""),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("query")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *QueryManagedObjectCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
	inputIterators := &flags.RequestInputIterators{}

	// query parameters
	query := flags.NewQueryTemplate()

	commonOptions, err := cliConfig.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}

	commonOptions.ResultProperty = "managedObjects"
	commonOptions.AddQueryParameters(query)

	orderBy := ""
	if v, err := cmd.Flags().GetString("orderBy"); err == nil {
		if v != "" {
			orderBy = v
		}
	}
	if v, err := cmd.Flags().GetString("query"); err == nil {
		if v != "" {
			c8yQuery := fmt.Sprintf("$filter=%s", url.QueryEscape(v))

			if orderBy != "" {
				c8yQuery = c8yQuery + fmt.Sprintf("+$orderby=%s", url.QueryEscape(orderBy))
			}

			query.SetVariable("query", c8yQuery)
		}
	} else {
		return cmderrors.NewUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "query", err))
	}
	if cmd.Flags().Changed("withParents") {
		if v, err := cmd.Flags().GetBool("withParents"); err == nil {
			query.SetVariable("withParents", fmt.Sprintf("%v", v))
		} else {
			return cmderrors.NewUserError("Flag does not exist")
		}
	}

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}

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
		IgnoreAccept: cliConfig.IgnoreAcceptHeader(),
		DryRun:       cliConfig.DryRun(),
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
