package c8yquerycmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

func NewInventoryQueryRunner(cmd *cobra.Command, args []string, f *cmdutil.Factory, opts ...flags.GetOption) func() error {
	return func() error {
		cfg, err := f.Config()
		if err != nil {
			return err
		}
		client, err := f.Client()
		if err != nil {
			return err
		}
		inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
		if err != nil {
			return err
		}

		// query parameters
		query := flags.NewQueryTemplate()

		c8yQueryParts, err := flags.WithC8YQueryOptions(
			cmd,
			inputIterators,
			opts...,
		)

		if err != nil {
			return err
		}

		// Compile query
		// replace all spaces with "+" due to url encoding
		filter := url.QueryEscape(strings.Join(c8yQueryParts, " and "))
		orderBy := "name"

		if v, err := cmd.Flags().GetString("orderBy"); err == nil {
			if v != "" {
				orderBy = url.QueryEscape(v)
			}
		}

		// q will automatically add a fragmentType=c8y_IsDevice to the query
		query.SetVariable("query", fmt.Sprintf("$filter=%s+$orderby=%s", filter, orderBy))

		err = flags.WithQueryParameters(
			cmd,
			query,
			inputIterators,
			flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetQueryParameters(), nil }, "custom"),
			flags.WithCustomStringValue(
				flags.BuildCumulocityQuery(cmd, c8yQueryParts, orderBy),
				func() string { return "query" },
				"query",
			),
		)
		if err != nil {
			return cmderrors.NewUserError(err)
		}
		commonOptions, err := cfg.GetOutputCommonOptions(cmd)
		if err != nil {
			return cmderrors.NewUserError(fmt.Sprintf("Failed to get common options. err=%s", err))
		}
		commonOptions.AddQueryParameters(query)

		queryValue, err := query.GetQueryUnescape(true)

		if err != nil {
			return cmderrors.NewSystemError("Invalid query parameter")
		}

		// headers
		headers := http.Header{}
		err = flags.WithHeaders(
			cmd,
			headers,
			inputIterators,
			flags.WithCustomStringSlice(func() ([]string, error) { return cfg.GetHeader(), nil }, "header"),
		)
		if err != nil {
			return cmderrors.NewUserError(err)
		}

		// form data
		formData := make(map[string]io.Reader)
		err = flags.WithFormDataOptions(
			cmd,
			formData,
			inputIterators,
		)
		if err != nil {
			return cmderrors.NewUserError(err)
		}

		// body
		body := mapbuilder.NewInitializedMapBuilder(true)
		err = flags.WithBody(
			cmd,
			body,
			inputIterators,
		)
		if err != nil {
			return cmderrors.NewUserError(err)
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
			Body:         body,
			FormData:     formData,
			Header:       headers,
			IgnoreAccept: cfg.IgnoreAcceptHeader(),
			DryRun:       cfg.ShouldUseDryRun(cmd.CommandPath()),
		}

		return f.RunWithWorkers(client, cmd, &req, inputIterators)
	}
}
