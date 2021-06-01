package cmderrors

import (
	"fmt"

	"github.com/pkg/errors"
)

type NoMatchesFoundError struct {
	Name string
	Err  error
}

func NewNoMatchesFoundError(name string) *NoMatchesFoundError {
	return &NoMatchesFoundError{
		Name: name,
		Err:  ErrNoMatchesFound,
	}
}

func (e *NoMatchesFoundError) Error() string {
	e.Err = ErrNoMatchesFound
	return fmt.Sprintf("%s. name=%s", e.Err, e.Name)
}
func (e *NoMatchesFoundError) Unwrap() error { return e.Err }

var ErrNoMatchesFound = errors.New("referenceByName: no matching items found")
var ErrMoreThanOneFound = errors.New("referenceByName: more than 1 found")
