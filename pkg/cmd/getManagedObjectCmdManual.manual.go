package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type getManagedObjectCmdManual struct {
	*baseCmd
}

func newGetManagedObjectCmdManual() *getManagedObjectCmdManual {
	ccmd := &getManagedObjectCmdManual{}

	cmd := &cobra.Command{
		Use:   "get2",
		Short: "Get managed objects/s",
		Long:  `Get a managed object by id`,
		Example: `
$ c8y inventory get --id 12345
Get a managed object

$ c8y inventory get --id 12345 --withParents
Get a managed object with parent references
        `,
		PreRunE: nil,
		RunE:    ccmd.getManagedObject,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{}, "ManagedObject id (required)")
	cmd.Flags().Bool("withParents", false, "include a flat list of all parents and grandparents of the given object")

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("id"),
	)

	// Required flags
	// cmd.MarkFlagRequired("id")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getManagedObjectCmdManual) getManagedObject(cmd *cobra.Command, args []string) error {
	inputIterators := &flags.RequestInputIterators{}
	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	query := flags.NewQueryTemplate()
	err = flags.WithQueryParameters(
		cmd,
		query,
		inputIterators,
		flags.WithBoolValue("withParents", "withParents"),
	)
	if err != nil {
		return newUserError(err)
	}
	commonOptions.AddQueryParameters(query)
	queryValue, err := query.GetQueryUnescape(true)

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewInitializedMapBuilder()

	// path parameters
	path := flags.NewStringTemplate("inventory/managedObjects/{id}")
	err = flags.WithPathParameters(
		cmd,
		path,
		inputIterators,
		flags.WithStringValue("id", "id"),
	)

	req := c8y.RequestOptions{
		Method:       "GET",
		Path:         path.GetTemplate(),
		Query:        queryValue,
		Body:         body,
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, inputIterators)
}
