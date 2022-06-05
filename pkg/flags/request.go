package flags

import (
	"io"
	"net/http"

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
	inputIterators := &RequestInputIterators{}

	// headers
	headers := http.Header{}
	err = WithHeaders(
		cmd,
		headers,
		inputIterators,
		builderOpts.HeaderOptions...,
	)

	if err != nil {
		return err
	}

	//
	// path parameters
	pathParameters := NewStringTemplate(req.Path)
	err = WithPathParameters(
		cmd,
		pathParameters,
		inputIterators,
		builderOpts.PathOptions...,
	)
	if err != nil {
		return err
	}

	//
	// query parameters
	queryValue := ""
	query := NewQueryTemplate()

	err = WithQueryParameters(
		cmd,
		query,
		inputIterators,
		builderOpts.QueryOptions...,
	)
	if err != nil {
		return err
	}

	queryValue, err = query.GetQueryUnescape(true)

	if err != nil {
		return err
	}

	//
	// form data
	formData := make(map[string]io.Reader)
	err = WithFormDataOptions(
		cmd,
		formData,
		inputIterators,
		builderOpts.FormDataOptions...,
	)
	if err != nil {
		return err
	}

	//
	// body
	body := mapbuilder.NewInitializedMapBuilder(true)
	err = WithBody(
		cmd,
		body,
		inputIterators,
		builderOpts.BodyOptions...,
	)
	if err != nil {
		return err
	}

	// set request values
	req.Header = headers
	req.Path = pathParameters.GetTemplate()
	req.Query = queryValue
	req.Body = body
	req.FormData = formData

	return nil
}

/*
	req := c8y.RequestOptions{
		Method:       "POST",
		IgnoreAccept: cliConfig.IgnoreAcceptHeader(),
        DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
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
		IgnoreAccept: cliConfig.IgnoreAcceptHeader(),
        DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
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

	if err != nil {
		return err
	}
*/
