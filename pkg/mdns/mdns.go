package mdns

import (
	"context"
	"time"

	"github.com/grandcat/zeroconf"
)

func Discover(service string, domain string, wait time.Duration, entries chan *zeroconf.ServiceEntry) error {
	if domain == "" {
		domain = "local"
	}

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	err = resolver.Browse(ctx, service, domain, entries)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
