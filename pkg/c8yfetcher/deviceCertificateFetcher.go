package c8yfetcher

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type DeviceCertificateFetcher struct {
	client *c8y.Client
	*IDNameFetcher
}

func NewDeviceCertificateFetcher(client *c8y.Client) *DeviceCertificateFetcher {
	return &DeviceCertificateFetcher{
		client: client,
	}
}

func (f *DeviceCertificateFetcher) IsID(id string) bool {
	isFingerprint := true
	for _, c := range id {
		if !strings.Contains("0123456789abcdef", string(c)) {
			isFingerprint = false
			break
		}
	}

	return isFingerprint && len(id) > 30
}

func (f *DeviceCertificateFetcher) getByID(id string) ([]fetcherResultSet, error) {
	cert, resp, err := f.client.DeviceCertificate.GetCertificate(
		WithDisabledDryRunContext(f.client),
		id,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by id")
	}

	results := make([]fetcherResultSet, 1)
	results[0] = fetcherResultSet{
		ID:    cert.Fingerprint,
		Name:  cert.Name,
		Self:  cert.Self,
		Value: resp.JSON(),
	}
	return results, nil
}

func (f *DeviceCertificateFetcher) getByName(name string) ([]fetcherResultSet, error) {
	// check if already resolved, so we can save a lookup
	col, _, err := f.client.DeviceCertificate.GetCertificates(
		WithDisabledDryRunContext(f.client),
		&c8y.DeviceCertificateCollectionOptions{
			PaginationOptions: *c8y.NewPaginationOptions(100),
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch by name")
	}

	results := make([]fetcherResultSet, 0)

	for _, cert := range col.Certificates {
		nameMatch, _ := matcher.MatchWithWildcards(cert.Name, name)
		fingerprintMatch, _ := matcher.MatchWithWildcards(cert.Fingerprint, name)

		if !nameMatch && !fingerprintMatch {
			continue
		}
		results = append(results, fetcherResultSet{
			ID:    cert.Fingerprint,
			Name:  cert.Name,
			Self:  cert.Self,
			Value: cert,
		})
	}

	return results, nil
}
