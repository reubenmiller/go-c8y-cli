package config

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonfilter"
)

// CommonCommandOptions control the handling of the response which are available for all commands
// which interact with the server
type CommonCommandOptions struct {
	ConfirmText       string
	OutputFile        string
	OutputFileRaw     string
	Filters           *jsonfilter.JSONFilters
	ResultProperty    string
	IncludeAll        bool
	WithTotalPages    bool
	WithTotalElements bool
	PageSize          int
	CurrentPage       int64
	TotalPages        int64
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

	if options.WithTotalElements {
		query.SetVariable(flags.FlagWithTotalElements, "true")
	}
}

func (options CommonCommandOptions) AddQueryParametersWithMapping(query *flags.QueryTemplate, aliases map[string]string) {
	if query == nil {
		return
	}

	if options.CurrentPage > 0 {
		if alias, ok := aliases[flags.FlagCurrentPage]; ok {
			query.SetVariable(alias, options.CurrentPage)
		} else {
			query.SetVariable(flags.FlagCurrentPage, options.CurrentPage)
		}
	}

	if options.PageSize > 0 {
		if alias, ok := aliases[flags.FlagPageSize]; ok {
			query.SetVariable(alias, options.PageSize)
		} else {
			query.SetVariable(flags.FlagPageSize, options.PageSize)
		}
	}

	if options.WithTotalPages {
		if alias, ok := aliases[flags.FlagWithTotalPages]; ok {
			query.SetVariable(alias, "true")
		} else {
			query.SetVariable(flags.FlagWithTotalPages, "true")
		}
	}

	if options.WithTotalElements {
		if alias, ok := aliases[flags.FlagWithTotalElements]; ok {
			query.SetVariable(alias, "true")
		} else {
			query.SetVariable(flags.FlagWithTotalElements, "true")
		}
	}
}
