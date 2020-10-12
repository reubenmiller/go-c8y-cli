// Code generated from specification version 1.0.0: DO NOT EDIT
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type getManagedObjectChildAssetReferenceCmd struct {
	*baseCmd
}

func newGetManagedObjectChildAssetReferenceCmd() *getManagedObjectChildAssetReferenceCmd {
	ccmd := &getManagedObjectChildAssetReferenceCmd{}

	cmd := &cobra.Command{
		Use:   "getChildAsset",
		Short: "Get managed object child asset reference",
		Long:  ``,
		Example: `
$ c8y inventoryReferences getChildAsset --asset 12345 --reference 12345
Get an existing child asset reference
        `,
		PreRunE: nil,
		RunE:    ccmd.getManagedObjectChildAssetReference,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("asset", []string{""}, "Asset id (required)")
	cmd.Flags().StringSlice("reference", []string{""}, "Asset reference id (required)")

	// Required flags
	cmd.MarkFlagRequired("asset")
	cmd.MarkFlagRequired("reference")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getManagedObjectChildAssetReferenceCmd) getManagedObjectChildAssetReference(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
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
	body := mapbuilder.NewMapBuilder()

	// path parameters
	pathParameters := make(map[string]string)
	if cmd.Flags().Changed("asset") {
		assetInputValues, assetValue, err := getFormattedDeviceSlice(cmd, args, "asset")

		if err != nil {
			return newUserError("no matching devices found", assetInputValues, err)
		}

		if len(assetValue) == 0 {
			return newUserError("no matching devices found", assetInputValues)
		}

		for _, item := range assetValue {
			if item != "" {
				pathParameters["asset"] = newIDValue(item).GetID()
			}
		}
	}
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

	return processRequestAndResponse([]c8y.RequestOptions{req}, commonOptions)
}
