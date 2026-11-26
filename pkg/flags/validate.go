package flags

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/spf13/cobra"
)

// ExactArgsOrExample returns an error if there are not exactly n args or --examples is not used
func ExactArgsOrExample(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("examples") {
			return nil
		}
		return cobra.ExactArgs(n)(cmd, args)
	}
}

// QueryValidator validates the query parameters of a request
type QueryValidator func(query url.Values) error

// ValidateQuery validates the query parameters of a request with all the validators, and returns the
// first error
func ValidateQuery(query url.Values, validators ...QueryValidator) error {
	for _, validate := range validators {
		if err := validate(query); err != nil {
			return err
		}
	}
	return nil
}

var ErrInvalidQueryParameter error = errors.New("invalid query parameters")

func WithQueryParameterOneRequired(name ...string) QueryValidator {
	return func(query url.Values) error {
		var err error
		matches := 0
		for _, k := range name {
			if query.Has(k) {
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

func WithQueryParameterMutuallyExclusive(name ...string) QueryValidator {
	return func(query url.Values) error {
		var err error
		matches := 0
		for _, k := range name {
			if query.Has(k) {
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

func WithQueryParameterRequiredTogether(name ...string) QueryValidator {
	return func(query url.Values) error {
		var err error
		matches := 0
		for _, k := range name {
			if query.Has(k) {
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
