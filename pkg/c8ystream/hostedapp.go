package c8ystream

import (
	"context"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/applications"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
)

// ResolveHostedApplication resolves the name of a hosted application to its id
// (an id is used as is). As in v2.59.0 only hosted applications are matched,
// the ones owned by the tenant first, and the name must match exactly. It
// returns "" when no hosted application has the name. The error of a failed
// request is a core.LookupError of a "hosted application".
func ResolveHostedApplication(ctx context.Context, client *apiv2.Client, tenant, value string) (string, error) {
	if c8ydata.IsID(value) {
		return value, nil
	}
	for _, owner := range []string{tenant, ""} {
		res := client.Applications.List(ctx, applications.ListOptions{
			Type:              applications.TypeHosted,
			Owner:             owner,
			PaginationOptions: pagination.PaginationOptions{PageSize: 2000},
		})
		if res.Err != nil {
			return "", core.WrapLookup("hosted application", value, res.Err)
		}
		for app, err := range res.Items() {
			if err != nil {
				return "", core.WrapLookup("hosted application", value, err)
			}
			if app.Name() == value {
				return app.ID(), nil
			}
		}
		// the applications of other tenants (e.g. cockpit) are only used when
		// the tenant has none with the name
	}
	return "", nil
}

// HostedApplicationRef resolves the application flag of a command which
// changes a hosted application (ui applications plugins ...): an id or the
// name of a hosted application. An empty value stays empty. A name which no
// hosted application has is an unresolved name (UnresolvedNameError).
func HostedApplicationRef(ctx context.Context, client *apiv2.Client, tenant, value string) (string, error) {
	if value == "" {
		return "", nil
	}
	id, err := ResolveHostedApplication(ctx, client, tenant, value)
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", UnresolvedNameError(value)
	}
	return id, nil
}
