package cmd

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type getGenericRestCmd struct {
	*baseCmd

	flagHost string
}

func newGetGenericRestCmd() *getGenericRestCmd {
	ccmd := &getGenericRestCmd{}

	cmd := &cobra.Command{
		Use:   "rest",
		Short: "Send generic REST request",
		Long:  `Send generic REST request`,
		Example: `
Get a list of managed objects
c8y rest GET /alarm/alarms

c8y rest GET "/alarm/alarms?pageSize=10&status=ACTIVE"

// Create a new alarm
c8y rest POST "alarm/alarms" --data "text=one,severity=MAJOR,type=test_Type,time=2019-01-01,source={'id': '12345'}"
		`,
		RunE: ccmd.getGenericRest,
	}

	addDataFlag(cmd)
	cmd.Flags().String("file", "", "File to be uploaded as a binary")
	cmd.Flags().StringSliceP("header", "H", nil, "headers. i.e. --header \"Accept: value\"")
	cmd.Flags().String("accept", "", "accept (header)")
	cmd.Flags().String("contentType", "", "content type (header)")
	cmd.Flags().StringVar(&ccmd.flagHost, "host", "", "host to use for the rest request. If empty, then the session's host will be used")
	cmd.Flags().Bool("ignoreAcceptHeader", false, "Without the accept header")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *getGenericRestCmd) getGenericRest(cmd *cobra.Command, args []string) error {
	method := "get"

	header := http.Header{}

	ignoreAcceptHeader := false
	if cmd.Flags().Changed("ignoreAcceptHeader") {
		if v, err := cmd.Flags().GetBool("ignoreAcceptHeader"); err == nil {
			ignoreAcceptHeader = v
		}
	}

	if v, err := cmd.Flags().GetString("accept"); err == nil && v != "" {
		if !ignoreAcceptHeader {
			header.Set("Accept", v)
		}
	}

	if v, err := cmd.Flags().GetString("contentType"); err == nil && v != "" {
		header.Set("Content-Type", v)
	}

	if values, err := cmd.Flags().GetStringSlice("header"); err == nil && len(values) > 0 {
		for _, v := range values {
			parts := strings.SplitN(v, ":", 2)
			if len(parts) != 2 {
				Logger.Warningf("Invalid header. %s", v)
				continue
			}
			Logger.Debugf("Setting header: name=%s, value=%s", parts[0], parts[1])
			header.Add(parts[0], strings.TrimSpace(parts[1]))
		}
	}

	var uri string
	if len(args) == 1 {
		uri = args[0]
	} else if len(args) > 1 {
		method = args[0]
		uri = args[1]
	}

	method = strings.ToUpper(method)

	if !(method == "GET" || method == "POST" || method == "PUT" || method == "DELETE") {
		return newUserError("Invalid method. Only GET, PUT, POST and DELETE are accepted")
	}

	if method == "PUT" {
		if err := validateUpdateMode(cmd, args); err != nil {
			return err
		}
	}
	if method == "POST" {
		if err := validateCreateMode(cmd, args); err != nil {
			return err
		}
	}
	if method == "DELETE" {
		if err := validateDeleteMode(cmd, args); err != nil {
			return err
		}
	}

	baseURL, _ := url.Parse(uri)

	var host string
	if n.flagHost != "" {
		host = n.flagHost
	}

	req := c8y.RequestOptions{
		Method:       method,
		Host:         host,
		Path:         baseURL.Path,
		Query:        baseURL.RawQuery,
		Header:       header,
		DryRun:       globalFlagDryRun,
		IgnoreAccept: ignoreAcceptHeader,
		ResponseData: nil,
	}

	commonOptions, err := getCommonOptions(cmd)
	if err != nil {
		return err
	}

	if method == "PUT" || method == "POST" {
		req.Body = getDataFlag(cmd)

		//if !cmd.Flags().Changed(FlagDataName) && !cmd.Flags().Changed("file") {
		//	return newUserError("Missing required arguments. Either --data or --file are required")
		//}

		// get file info
		if cmd.Flags().Changed("file") {
			req.FormData = make(map[string]io.Reader)
			getFileFlag(cmd, "file", req.FormData)
		}
	}

	// Hide usage for system errors
	cmd.SilenceUsage = true

	return processRequestAndResponse([]c8y.RequestOptions{req}, commonOptions)
}
