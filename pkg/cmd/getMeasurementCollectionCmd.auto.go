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

type getMeasurementCollectionCmd struct {
	*baseCmd
}

func newGetMeasurementCollectionCmd() *getMeasurementCollectionCmd {
	ccmd := &getMeasurementCollectionCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a collection of measurements based on filter parameters",
		Long:  `Get a collection of measurements based on filter parameters`,
		Example: `
$ c8y measurements list
Get a list of measurements
		`,
		RunE: ccmd.getMeasurementCollection,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device ID")
	cmd.Flags().String("type", "", "Measurement type.")
	cmd.Flags().String("valueFragmentType", "", "value fragment type")
	cmd.Flags().String("valueFragmentSeries", "", "value fragment series")
	cmd.Flags().String("fragmentType", "", "Fragment name from measurement (deprecated).")
	cmd.Flags().String("dateFrom", "", "Start date or date and time of measurement occurrence.")
	cmd.Flags().String("dateTo", "", "End date or date and time of measurement occurrence.")
	cmd.Flags().Bool("revert", false, "Return the newest instead of the oldest measurements. Must be used with dateFrom and dateTo parameters")
	cmd.Flags().Bool("csv", false, "Results will be displayed in csv format")
	cmd.Flags().Bool("excel", false, "Results will be displayed in Excel format")
	cmd.Flags().String("unit", "", "Every measurement fragment which contains 'unit' property will be transformed to use required system of units.")

	// Required flags

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getMeasurementCollectionCmd) getMeasurementCollection(cmd *cobra.Command, args []string) error {

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}
	if cmd.Flags().Changed("device") {
		deviceInputValues, deviceValue, err := getFormattedDeviceSlice(cmd, args, "device")

		if err != nil {
			return newUserError("no matching devices found", deviceInputValues, err)
		}

		if len(deviceValue) == 0 {
			return newUserError("no matching devices found", deviceInputValues)
		}

		for _, item := range deviceValue {
			if item != "" {
				query.Add("source", newIDValue(item).GetID())
			}
		}
	}
	if v, err := cmd.Flags().GetString("type"); err == nil {
		if v != "" {
			query.Add("type", url.QueryEscape(v))
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "type", err))
	}
	if v, err := cmd.Flags().GetString("valueFragmentType"); err == nil {
		if v != "" {
			query.Add("valueFragmentType", url.QueryEscape(v))
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "valueFragmentType", err))
	}
	if v, err := cmd.Flags().GetString("valueFragmentSeries"); err == nil {
		if v != "" {
			query.Add("valueFragmentSeries", url.QueryEscape(v))
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "valueFragmentSeries", err))
	}
	if v, err := cmd.Flags().GetString("fragmentType"); err == nil {
		if v != "" {
			query.Add("fragmentType", url.QueryEscape(v))
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "fragmentType", err))
	}
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
	if cmd.Flags().Changed("revert") {
		if v, err := cmd.Flags().GetBool("revert"); err == nil {
			query.Add("revert", fmt.Sprintf("%v", v))
		} else {
			return newUserError("Flag does not exist")
		}
	}
	commonOptions.AddQueryParameters(&query)
	if cmd.Flags().Changed("pageSize") || globalUseNonDefaultPageSize {
		if v, err := cmd.Flags().GetInt("pageSize"); err == nil && v > 0 {
			query.Add("pageSize", fmt.Sprintf("%d", v))
		}
	}
	queryValue, err = url.QueryUnescape(query.Encode())

	if err != nil {
		return newSystemError("Invalid query parameter")
	}

	// headers
	headers := http.Header{}
	if cmd.Flags().Changed("csv") {
		if _, err := cmd.Flags().GetBool("csv"); err == nil {
			headers.Add("Accept", "text/csv")
		} else {
			return newUserError("Flag does not exist")
		}
	}
	if cmd.Flags().Changed("excel") {
		if _, err := cmd.Flags().GetBool("excel"); err == nil {
			headers.Add("Accept", "application/vnd.ms-excel")
		} else {
			return newUserError("Flag does not exist")
		}
	}
	if v, err := cmd.Flags().GetString("unit"); err == nil {
		if v != "" {
			headers.Add("X-Cumulocity-System-Of-Units", v)
		}
	} else {
		return newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", "unit", err))
	}

	// form data
	formData := make(map[string]io.Reader)

	// body
	body := mapbuilder.NewMapBuilder()

	// path parameters
	pathParameters := make(map[string]string)

	path := replacePathParameters("measurement/measurements", pathParameters)

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
