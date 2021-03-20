package config

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonfilter"
)

// CommonCommandOptions control the handling of the response which are available for all commands
// which interact with the server
type CommonCommandOptions struct {
	ConfirmText    string
	OutputFile     string
	OutputFileRaw  string
	Filters        *jsonfilter.JSONFilters
	ResultProperty string
	IncludeAll     bool
	WithTotalPages bool
	PageSize       int
	CurrentPage    int64
	TotalPages     int64
}

// AddQueryParameters adds the common query parameters to the given query values
func (options CommonCommandOptions) AddQueryParameters(query *flags.QueryTemplate) {
	if query == nil {
		return
	}

	if options.CurrentPage > 0 {
		query.SetVariable(flags.FlagCurrentPage, options.CurrentPage)
	}

	if options.PageSize > 0 {
		query.SetVariable(flags.FlagPageSize, options.PageSize)
	}

	if options.WithTotalPages {
		query.SetVariable(flags.FlagWithTotalPages, "true")
	}
}
