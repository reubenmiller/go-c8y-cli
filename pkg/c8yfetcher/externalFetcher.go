package c8yfetcher

import (
	"bytes"
	"regexp"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extexec"
)

type ExternalFetcher struct {
	externalCommand []string
	IDPattern       *regexp.Regexp
	*DefaultFetcher
}

func NewExternalFetcher(externalCommand []string, idPattern string) *ExternalFetcher {
	opt := ExternalFetcher{
		externalCommand: externalCommand,
	}

	if idPattern != "" {
		if p, err := regexp.Compile(idPattern); err != nil {
			opt.IDPattern = p
		}
	}
	return &opt
}

func (f *ExternalFetcher) IsID(v string) bool {
	if f.IDPattern != nil {
		return f.IDPattern.MatchString(v)
	}
	return false
}

func (f *ExternalFetcher) getByID(id string) ([]fetcherResultSet, error) {
	results := make([]fetcherResultSet, 1)
	results[0] = fetcherResultSet{
		ID: id,
	}
	return results, nil
}

func (f *ExternalFetcher) getByName(name string) ([]fetcherResultSet, error) {
	output, err := extexec.ExecuteExternalCommand(name, f.externalCommand)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by name")
	}

	results := make([]fetcherResultSet, 0)
	for _, row := range bytes.Split(output, []byte("\n")) {
		if len(row) > 0 {
			results = append(results, fetcherResultSet{
				ID: string(row),
			})
		}
	}

	return results, nil
}
