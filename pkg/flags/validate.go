package flags

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// ExactArgsOrExample returns an error if there are not exactly n args or --exapmles is not used
func ExactArgsOrExample(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("examples") {
			return nil
		}
		return cobra.ExactArgs(n)(cmd, args)
	}
}

var ErrInvalidQueryParameter error = errors.New("invalid query parameters")

func WithQueryParameterOneRequired(name ...string) c8y.RequestValidator {
	return func(r *http.Request) error {
		var err error
		matches := 0
		for _, k := range name {
			if r.URL.Query().Has(k) {
				matches += 1
				break
			}
		}

		if matches == 0 {
			msg := fmt.Sprintf("should have one of the following set [%s]", strings.Join(name, ","))
			err = cmderrors.NewUserErrorWithExitCode(cmderrors.ExitUserError, msg)
		}
		return err
	}
}

func WithQueryParameterMutuallyExclusive(name ...string) c8y.RequestValidator {
	return func(r *http.Request) error {
		var err error
		matches := 0
		for _, k := range name {
			if r.URL.Query().Has(k) {
				matches += 1
			}
		}
		if matches > 1 {
			msg := fmt.Sprintf("should have only one of the following set [%s]", strings.Join(name, ","))
			err = cmderrors.NewUserErrorWithExitCode(cmderrors.ExitUserError, msg)
		}
		return err
	}
}

func WithQueryParameterRequiredTogether(name ...string) c8y.RequestValidator {
	return func(r *http.Request) error {
		var err error
		matches := 0
		for _, k := range name {
			if r.URL.Query().Has(k) {
				matches += 1
			}
		}
		if matches == len(name) {
			// err = fmt.Errorf("%w. should have only one of the following set [%s]", ErrInvalidQueryParameter, strings.Join(name, ","))
			msg := fmt.Sprintf("should have all the following set [%s]", strings.Join(name, ","))
			err = cmderrors.NewUserErrorWithExitCode(cmderrors.ExitUserError, msg)
		}
		return err
	}
}
