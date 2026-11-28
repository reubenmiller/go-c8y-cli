// Package testclient builds go-c8y clients for tests.
package testclient

import apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"

// Must is apiv2.NewClient for tests: it panics when the options are invalid.
func Must(opts apiv2.ClientOptions) *apiv2.Client {
	client, err := apiv2.NewClient(opts)
	if err != nil {
		panic(err)
	}
	return client
}
