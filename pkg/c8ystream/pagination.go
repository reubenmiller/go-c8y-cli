package c8ystream

import (
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
)

// Pagination returns the pagination options of a list command, built once
// from the common flags (--pageSize, --withTotalPages, --withTotalElements,
// --currentPage, --includeAll/--totalPages via MaxItems, and
// --paginationStrategy). Every list command passes it to its service's ListAll
// unchanged, so paging behaves the same everywhere; --raw is handled by the
// Runner when it renders the iterator.
//
// A command whose strategy depends on its own flags (e.g. an inventory query
// that switches to the id keyset for --includeAll) overrides Strategy on the
// returned value.
func (r *Runner) Pagination() (pagination.PaginationOptions, error) {
	common, err := r.Config.GetOutputCommonOptions(r.Cmd)
	if err != nil {
		return pagination.PaginationOptions{}, err
	}
	return pagination.PaginationOptions{
		PageSize:          common.PageSize,
		WithTotalPages:    common.WithTotalPages,
		WithTotalElements: common.WithTotalElements,
		CurrentPage:       int(common.CurrentPage),
		MaxItems:          r.Config.MaxItems(),
		Strategy:          pagination.StrategyKind(r.Config.PaginationStrategy()),
		// Render the full-walk first request in dry runs of --includeAll.
		FullWalk: r.Config.IncludeAll(),
		// A command that returns one page sends the request as given (v2.59.0:
		// e.g. "?pageSize=1"), not the paging strategy's decorated first
		// request (currentPage=1, the id keyset query). The real request then
		// matches the --dry one, and --currentPage is kept.
		SinglePage: !r.Config.IncludeAll() && r.Config.GetTotalPages() <= 1,
	}, nil
}
