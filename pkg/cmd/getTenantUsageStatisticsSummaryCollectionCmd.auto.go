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

type getTenantUsageStatisticsSummaryCollectionCmd struct {
	*baseCmd
}

func newGetTenantUsageStatisticsSummaryCollectionCmd() *getTenantUsageStatisticsSummaryCollectionCmd {
	ccmd := &getTenantUsageStatisticsSummaryCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "listSummaryForTenant",
		Short: "Get collection of tenant usage statistics summary",
		Long:  `Get summary of requests and database usage from the start of this month until now`,
		Example: `
$ c8y tenantStatistics listSummaryForTenant
Get tenant summary statistics for the current tenant

$ c8y tenantStatistics listSummaryForTenant --dateFrom "-30d"
Get tenant summary statistics collection for the last 30 days

$ c8y tenantStatistics listSummaryForTenant --dateFrom "-10d" --dateTo "-9d"
Get tenant summary statistics collection for the last 10 days, only return until the last 9 days
		`,
		RunE: ccmd.getTenantUsageStatisticsSummaryCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("dateFrom", "", "Start date or date and time of the statistics.")
	cmd.Flags().String("dateTo", "", "End date or date and time of the statistics.")

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getTenantUsageStatisticsSummaryCollectionCmd) getTenantUsageStatisticsSummaryCollection(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return newUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if flagVal, err := cmd.Flags().GetString("dateFrom"); err == nil && flagVal != "" {
		if v, err := tryGetTimestampFlag(cmd, "dateFrom"); err == nil && v != "" {
			query.Add("dateFrom", v)
		} else {
			return newUserError("invalid date format", err)
		}
	}
	if flagVal, err := cmd.Flags().GetString("dateTo"); err == nil && flagVal != "" {
		if v, err := tryGetTimestampFlag(cmd, "dateTo"); err == nil && v != "" {
			query.Add("dateTo", v)
		} else {
			return newUserError("invalid date format", err)
		}
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
	body := mapbuilder.NewMapBuilder()

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("/tenant/statistics/summary", pathParameters)

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
