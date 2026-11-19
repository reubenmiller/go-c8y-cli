// Package childref dispatches the device-group children commands' --childType
// flag (addition|asset|device) to the matching go-c8y v2 child-reference service
// (ChildAdditions / ChildAssets / ChildDevices). The three SDK services share an
// identical method set, so a single interface backs the get/create/assign/unassign
// commands; list is dispatched separately because each service's ListOptions is a
// distinct (but structurally identical) named type.
package childref

import (
	"context"
	"fmt"

	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects/child"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects/childadditions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects/childassets"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects/childdevices"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
)

// Child relationship type flag values.
const (
	TypeAddition = "addition"
	TypeAsset    = "asset"
	TypeDevice   = "device"
)

// Service is the subset of a child-reference service used by the children
// get/create/assign/unassign commands. *childadditions.Service,
// *childassets.Service and *childdevices.Service all satisfy it.
type Service interface {
	Get(ctx context.Context, parentID, childID string) op.Result[jsonmodels.ManagedObject]
	Create(ctx context.Context, parentID string, body any) op.Result[jsonmodels.ManagedObject]
	Assign(ctx context.Context, parentID string, child any) op.Result[core.NoContent]
	Unassign(ctx context.Context, parentID, childID string) op.Result[core.NoContent]
}

// First returns the first non-empty entry of a slice flag — the parent-id path
// parameter takes a single value even when the flag is declared as a slice.
func First(values []string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// For returns the child-reference service for the given childType
// (addition|asset|device), or an error for any other value.
func For(client *apiv2.Client, childType string) (Service, error) {
	switch childType {
	case TypeAddition:
		return client.ManagedObjects.ChildAdditions, nil
	case TypeAsset:
		return client.ManagedObjects.ChildAssets, nil
	case TypeDevice:
		return client.ManagedObjects.ChildDevices, nil
	default:
		return nil, fmt.Errorf("invalid child type %q (expected addition, asset or device)", childType)
	}
}

// ListAll returns the paginating list function for childType bound to parentID,
// shaped for c8ystream.ListCall. An invalid childType yields an error iterator.
func ListAll(client *apiv2.Client, childType, parentID string) func(context.Context, child.ListOptions) *pagination.Iterator[jsonmodels.ManagedObject] {
	return func(ctx context.Context, o child.ListOptions) *pagination.Iterator[jsonmodels.ManagedObject] {
		switch childType {
		case TypeAddition:
			return client.ManagedObjects.ChildAdditions.ListAll(ctx, parentID, childadditions.ListOptions(o))
		case TypeAsset:
			return client.ManagedObjects.ChildAssets.ListAll(ctx, parentID, childassets.ListOptions(o))
		case TypeDevice:
			return client.ManagedObjects.ChildDevices.ListAll(ctx, parentID, childdevices.ListOptions(o))
		default:
			return pagination.NewErrorIterator[jsonmodels.ManagedObject](fmt.Errorf("invalid child type %q (expected addition, asset or device)", childType))
		}
	}
}
