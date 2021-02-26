package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type GetAgentCollectionCmd struct {
	*baseCmd
}

func NewGetAgentCollectionCmd() *GetAgentCollectionCmd {
	ccmd := &GetAgentCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a collection of agents based on filter parameters",
		Long:  `Get a collection of agents based on filter parameters`,
		Example: `
		c8y agents list --name "sensor*" --type myType

		Get a collection of agents of type "myType", and their names start with "sensor"
		`,
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("name", "", "Agent name.")
	cmd.Flags().String("type", "", "Agent type.")
	cmd.Flags().String("fragmentType", "", "Agent fragment type.")
	cmd.Flags().String("owner", "", "Agent owner.")
	cmd.Flags().String("query", "", "Additional query filter")
	cmd.Flags().Bool("withParents", false, "include a flat list of all parents and grandparents of the given object")

	// Required flags
	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *GetAgentCollectionCmd) RunE(cmd *cobra.Command, args []string) error {
	var err error
	inputIterators, err := flags.NewRequestInputIterators(cmd)
	if err != nil {
		return err
	}

	// query parameters
	query := flags.NewQueryTemplate()

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	commonOptions.ResultProperty = "managedObjects"
	commonOptions.AddQueryParameters(query)

	c8yQueryParts, err := flags.WithC8YQueryOptions(
		cmd,
		flags.WithC8YQueryFixedString("(has(com_cumulocity_model_Agent))"),
		flags.WithC8YQueryFormat("name", "(name eq '%s')"),
		flags.WithC8YQueryFormat("type", "(type eq '%s')"),
		flags.WithC8YQueryFormat("fragmentType", "has(%s)"),
		flags.WithC8YQueryFormat("owner", "(owner eq '%s')"),
		flags.WithC8YQueryFormat("query", "%s"),
	)

	if err != nil {
		return err
	}

	// Compile query
	// replace all spaces with "+" due to url encoding
	filter := url.QueryEscape(strings.Join(c8yQueryParts, " and "))
	orderBy := url.QueryEscape("name")
	query.SetVariable("query", fmt.Sprintf("$filter=%s+$orderby=%s", filter, orderBy))

	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithBoolValue("withParents", "withParents"),
	)

	if err != nil {
		return nil
	}

	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return cmderrors.NewSystemError("Invalid query parameter")
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
		IgnoreAccept: globalFlagIgnoreAccept,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
