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
	pageSize, maxItems := common.PageSize, r.Config.MaxItems()
	if r.noPageSizeActive() {
		// the platform's default page size, which the requests do not send
		// (see noPageSize); --totalPages N reads N pages of it
		pageSize = noPageSize
		maxItems = int64(noPageSize) * max(r.Config.GetTotalPages(), 1)
		r.dropPageSize.Store(true)
	}
	return pagination.PaginationOptions{
		PageSize:          pageSize,
		WithTotalPages:    common.WithTotalPages,
		WithTotalElements: common.WithTotalElements,
		CurrentPage:       int(common.CurrentPage),
		MaxItems:          maxItems,
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

// PaginationByPage returns the pagination options of a list which --totalPages
// N (N > 1, without --includeAll) reads by page number, as in v2.59.0:
// "--pageSize 1 --totalPages 3" returns 3 items with the requests
// "?pageSize=1", "?currentPage=2&pageSize=1" and "?currentPage=3&pageSize=1".
// It is used by the time-series lists (events, alarms, measurements,
// operations), whose default time keyset requests the items of the boundary
// timestamp again on every page (so N requests return fewer than N pages of
// items), and by "inventory list". An explicit --paginationStrategy is kept.
func (r *Runner) PaginationByPage() (pagination.PaginationOptions, error) {
	paging, err := r.Pagination()
	if err != nil {
		return paging, err
	}
	if r.Config.GetTotalPages() > 1 && !r.Config.IncludeAll() && (paging.Strategy == "" || paging.Strategy == pagination.StrategyAuto) {
		paging.Strategy = pagination.StrategyOffset
	}
	return paging, nil
}

// noPageSizeActive reports whether a list sends no page size: the page size
// (setting or --pageSize) is blank, 0 or 5, and there is no --includeAll. A dry
// run shows the request without the SDK's default page size already.
func (r *Runner) noPageSizeActive() bool {
	cfg := r.Config
	common, err := cfg.GetOutputCommonOptions(r.Cmd)
	if err != nil || common.PageSize != 0 || cfg.IncludeAll() {
		return false
	}
	return !cfg.ShouldUseDryRun(r.Cmd.CommandPath())
}
