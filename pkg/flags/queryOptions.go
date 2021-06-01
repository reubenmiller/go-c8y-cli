package flags

import (
	"errors"
)

var ErrFlagDoesNotExist = errors.New("Flag does not exist")

func WithCurrentPage() GetOption {
	return WithStringValue(FlagCurrentPage, FlagCurrentPage)
}

func WithPageSize() GetOption {
	return WithStringValue(FlagPageSize, FlagPageSize)
}

func WithTotalPages() GetOption {
	return WithStringValue(FlagWithTotalPages, FlagWithTotalPages)
}
