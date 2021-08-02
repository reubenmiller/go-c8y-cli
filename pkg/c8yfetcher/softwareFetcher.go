package c8yfetcher

import (
	"fmt"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type SoftwareFetcher struct {
	*ManagedObjectFetcher
}

func NewSoftwareFetcher(client *c8y.Client) *SoftwareFetcher {
	return &SoftwareFetcher{
		ManagedObjectFetcher: &ManagedObjectFetcher{
			client: client,
			Query: func(s string) string {
				return fmt.Sprintf("(type eq 'c8y_Software') and name eq '%s'", s)
			},
		},
	}
}
