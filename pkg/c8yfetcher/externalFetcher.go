package c8yfetcher

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extexec"
)

type ExternalFetcher struct {
	externalCommand []string
	*DefaultFetcher
}

func NewExternalFetcher(externalCommand []string) *ExternalFetcher {
	return &ExternalFetcher{
		externalCommand: externalCommand,
	}
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
