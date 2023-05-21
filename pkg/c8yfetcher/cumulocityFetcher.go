package c8yfetcher

import (
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type CumulocityFetcher struct {
	factory *cmdutil.Factory
	*DefaultFetcher
}

func (f *CumulocityFetcher) Client() *c8y.Client {
	client, err := f.factory.Client()
	if err != nil {
		panic(err)
	}
	return client
}
