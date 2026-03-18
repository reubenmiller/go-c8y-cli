package c8ywaiter

import (
	"context"
	"net/http"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

// TenantExistence tenant existence checker
type TenantExistence struct {
	ID     string
	Client *c8y.Client
	Negate bool
}

type tenantResponse struct {
	Tenant   *c8y.Tenant
	Response *c8y.Response
}

// Check check if tenant exists or not
func (s *TenantExistence) Check(m interface{}) (done bool, err error) {
	if result, ok := m.(*tenantResponse); ok {
		var exists, notFound bool
		tenantID := s.ID

		if result.Response != nil {
			exists = result.Response.StatusCode() >= 200 && result.Response.StatusCode() <= 399
			notFound = result.Response.StatusCode() == http.StatusNotFound
		}

		if s.Negate {
			done = notFound
			if !done {
				err = cmderrors.NewAssertionError(&cmderrors.AssertionError{
					Type:    cmderrors.Tenant,
					Wanted:  "NotFound",
					Got:     "Found",
					Context: struct{ ID string }{ID: tenantID},
				})
			}
		} else {
			done = exists
			if !done {
				err = cmderrors.NewAssertionError(&cmderrors.AssertionError{
					Type:    cmderrors.Tenant,
					Wanted:  "Found",
					Got:     "NotFound",
					Context: struct{ ID string }{ID: tenantID},
				})
			}
		}

		if done {
			return done, nil
		}
	}
	return
}

func (s *TenantExistence) SetValue(v interface{}) error {
	if id, ok := v.(string); ok {
		s.ID = id
	}
	return nil
}

// Get get current tenant state
func (s *TenantExistence) Get() (interface{}, error) {
	// Corner case: check if it is the current tenant as the getTenant by id will fail with 403 error
	// however since they already have access to the current tenant so it should work fine
	if currentTenant := s.Client.GetTenantName(context.Background()); currentTenant == s.ID {
		tenant, resp, err := s.Client.Tenant.GetCurrentTenant(context.Background())
		// resp.JSON("id").String()
		return &tenantResponse{&c8y.Tenant{
			ID:     tenant.Name,
			Domain: tenant.DomainName,
		}, resp}, err
	}

	tenant, resp, err := s.Client.Tenant.GetTenant(
		context.Background(),
		s.ID,
	)

	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		// ignore not found errors, these are processed in the Check func
		err = nil
	}
	return &tenantResponse{tenant, resp}, err
}
