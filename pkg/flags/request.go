package flags

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type RequestBuilder struct {
	HeaderOptions   []GetOption
	QueryOptions    []GetOption
	BodyOptions     []GetOption
	FormDataOptions []GetOption
	PathOptions     []GetOption
}

func WithRequestOptions(cmd *cobra.Command, args []string, req *c8y.RequestOptions, builderOpts *RequestBuilder) (err error) {

	// headers
	headers := http.Header{}
	err = WithHeaders(
		cmd,
		headers,
		builderOpts.HeaderOptions...,
	)

	if err != nil {
		return err
	}

	//
	// path parameters
	pathParameters := make(map[string]string)
	err = WithPathParameters(
		cmd,
		pathParameters,
		builderOpts.PathOptions...,
	)

	//
	// query parameters
	queryValue := url.QueryEscape("")
	query := url.Values{}

	err = WithQueryParameters(
		cmd,
		query,
		builderOpts.QueryOptions...,
	)
	if err != nil {
		return err
	}

	queryValue, err = url.QueryUnescape(query.Encode())

	if err != nil {
		return err
	}

	//
	// form data
	formData := make(map[string]io.Reader)
	err = WithFormDataOptions(
		cmd,
		formData,
		builderOpts.FormDataOptions...,
	)
	if err != nil {
		return err
	}

	//
	// body
	body := mapbuilder.NewInitializedMapBuilder()
	err = WithBody(
		cmd,
		body,
		builderOpts.BodyOptions...,
	)
	if err != nil {
		return err
	}

	if body != nil {
		if err := body.Validate(); err != nil {
			return fmt.Errorf("body validation error. %w", err)
		}
	}

	// set request values
	req.Header = headers
	req.Path = replacePathParameters(req.Path, pathParameters)
	req.Query = queryValue
	req.Body = body
	req.FormData = formData

	return nil
}

func replacePathParameters(uri string, parameters map[string]string) string {
	if parameters == nil {
		return uri
	}
	for key, value := range parameters {
		uri = strings.ReplaceAll(uri, "{"+key+"}", value)
	}
	return uri
}

/*
	req := c8y.RequestOptions{
		Method:       "POST",
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	err = flags.WithRequestOptions(
		cmd,
		args,
		&req,
		&flags.RequestBuilder{
			Path: "${RESTPath}"
			HeaderOptions: []flags.GetOption{
				flags.WithProcessingModeValue(),
			},

			BodyOptions: []flags.GetOption{
				flags.WithDataValue(FlagDataName, ""),
				WithDeviceByNameFirstMatch(args, "newChildDevice", "managedObject.id"),
			},

			FormDataOptions: []flags.GetOption{},

			PathOptions: []flags.GetOption{
				WithDeviceGroupByNameFirstMatch(args, "group", "id"),
			},

			QueryOptions: []flags.GetOption{},
		},
	)
	if err != nil {
		return err
	}

	req := c8y.RequestOptions{
		$RESTHost
		Path:         "$RESTPath",
		Method:       "$RESTMethod",
		IgnoreAccept: false,
		DryRun:       globalFlagDryRun,
	}

	err = flags.WithRequestOptions(
		cmd,
		args,
		&req,
		&flags.RequestBuilder{
			HeaderOptions: []flags.GetOption{
				$RestHeaderBuilderOptions
			},

			BodyOptions: []flags.GetOption{
				$RESTBodyBuilderOptions
			},

			FormDataOptions: []flags.GetOption{
				$RESTFormDataBuilder
			},

			PathOptions: []flags.GetOption{
				$RESTPathBuilderOptions
			},

			QueryOptions: []flags.GetOption{
				$RESTQueryBuilderWithValues
			},
		},
	)

	// bodyErr := body.MergeJsonnet(`
	//	addIfEmptyString(base, "password", {sendPasswordResetEmail: true})
	//`, false)


	if err != nil {
		return err
	}
*/
