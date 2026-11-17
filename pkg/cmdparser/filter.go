package cmdparser

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/integration/models"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

// ApplyCommandFilter adds the client side filters of an extension API command
// to the output filters. The filters are combined with the filters provided by
// the user (via --filter), so the user can only narrow the results further.
//
// The common options are cached by the configuration and the filters are
// referenced by pointer, so the filters are used when the response is processed
func ApplyCommandFilter(options config.CommonCommandOptions, spec models.Command) error {
	if len(spec.Filter) == 0 || options.Filters == nil {
		return nil
	}
	if err := options.Filters.AddRawFilters(spec.Filter); err != nil {
		return fmt.Errorf("invalid filter in command '%s'. %w", spec.Name, err)
	}
	return nil
}
