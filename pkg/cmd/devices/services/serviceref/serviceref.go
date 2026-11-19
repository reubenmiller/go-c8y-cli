// Package serviceref resolves a device-service reference (the `--id` of the
// `c8y devices services get|update|delete` commands) to a managed-object id.
// A plain numeric id (the common piped/`--id` case) passes straight through; a
// name is looked up among the parent device's c8y_Service child additions,
// matching v1's device-scoped service-name resolution. It lives in its own leaf
// package so the get/update/delete commands can share it without importing the
// `services` parent (which would form an import cycle).
package serviceref

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects/childadditions"
)

// ServiceType is the managed-object type used to model a device service.
const ServiceType = "c8y_Service"

// First returns the first non-empty entry of a slice flag — the device path
// reference takes a single value even though the flag is declared as a slice.
func First(values []string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// ResolveID resolves a device-service reference to its managed-object id. A plain
// numeric id passes through unchanged (no lookup, so it works under --dry without
// a server round-trip); otherwise the name is resolved among the parent device's
// c8y_Service child additions. deviceRef is only needed for the name case.
func ResolveID(ctx context.Context, client *apiv2.Client, deviceRef, idRef string) (string, error) {
	if idRef == "" || isPlainID(idRef) {
		return idRef, nil
	}
	deviceID, err := client.ManagedObjects.ResolveID(ctx, c8ystream.NameOrID(deviceRef), nil)
	if err != nil {
		return "", err
	}
	opt := childadditions.ListOptions{
		Query: fmt.Sprintf("(type eq '%s') and (name eq '%s')", ServiceType, idRef),
	}
	opt.PageSize = 1
	for mo, err := range client.ManagedObjects.ChildAdditions.ListAll(ctx, deviceID, opt).Items() {
		if err != nil {
			return "", err
		}
		return mo.ID(), nil
	}
	return "", fmt.Errorf("no device service found with name %q", idRef)
}

// isPlainID reports whether value is a non-empty all-digit id.
func isPlainID(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
