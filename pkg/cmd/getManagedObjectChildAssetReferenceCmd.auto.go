// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type GetManagedObjectChildAssetReferenceCmd struct {
	*baseCmd
}

func NewGetManagedObjectChildAssetReferenceCmd() *GetManagedObjectChildAssetReferenceCmd {
	ccmd := &GetManagedObjectChildAssetReferenceCmd{}
	cmd := &cobra.Command{
		Use:   "getChildAsset",
		Short: "Get managed object child asset reference",
		Long:  ``,
		Example: `
$ c8y inventoryReferences getChildAsset --asset 12345 --reference 12345
Get an existing child asset reference
        `,
		PreRunE: nil,
		RunE:    ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("asset", []string{""}, "Asset id (required) (accepts pipeline)")
	cmd.Flags().StringSlice("reference", []string{""}, "Asset reference id (required)")

	flags.WithOptions(
		cmd,
		flags.WithPipelineSupport("asset"),
	)

	// Required flags
	cmd.MarkFlagRequired("reference")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *GetManagedObjectChildAssetReferenceCmd) RunE(cmd *cobra.Command, args []string) error {
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err := flags.WithQueryOptions(
		cmd,
		query,
	)
	if err != nil {
		return newUserError(err)
	}
	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}
	commonOptions.AddQueryParameters(&query)

	queryValue, err = url.QueryUnescape(query.Encode())

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
	pathParameters := make(map[string]string)
	if cmd.Flags().Changed("reference") {
		referenceInputValues, referenceValue, err := getFormattedDeviceSlice(cmd, args, "reference")

		if err != nil {
			return newUserError("no matching devices found", referenceInputValues, err)
		}

		if len(referenceValue) == 0 {
			return newUserError("no matching devices found", referenceInputValues)
		}

		for _, item := range referenceValue {
			if item != "" {
				pathParameters["reference"] = newIDValue(item).GetID()
			}
		}
	}

	path := replacePathParameters("inventory/managedObjects/{asset}/childAssets/{reference}", pathParameters)

	req := c8y.RequestOptions{
		Method:       "GET",
		Path:         path,
		Query:        queryValue,
		Body:         body.GetMap(),
		FormData:     formData,
		Header:       headers,
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	return processRequestAndResponseWithWorkers(cmd, &req, "asset")
}
