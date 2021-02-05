package flags

import (
	"errors"
)

var ErrFlagDoesNotExist = errors.New("Flag does not exist")

func WithCurrentPage() GetOption {
	return WithStringValue("currentPage", "currentPage")
}

func WithPageSize() GetOption {
	return WithStringValue("pageSize", "pageSize")
}

func WithTotalPages() GetOption {
	return WithStringValue("withTotalPages", "withTotalPages")
}
