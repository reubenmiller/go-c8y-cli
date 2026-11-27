package repodetails

import (
	"context"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/core"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/model"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
)

// VersionKind is a kind of version stored in a repository
type VersionKind struct {
	// Type is the managed object type, e.g. c8y_SoftwareBinary
	Type string
	// Fragment is the fragment which holds the version, e.g. c8y_Software
	Fragment string
	// Patch is true for the patches of a firmware, false for its versions
	Patch bool
}

var (
	// SoftwareVersionKind is a version of a software package
	SoftwareVersionKind = VersionKind{Type: "c8y_SoftwareBinary", Fragment: "c8y_Software"}
	// FirmwareVersionKind is a version of a firmware
	FirmwareVersionKind = VersionKind{Type: "c8y_FirmwareBinary", Fragment: "c8y_Firmware"}
	// FirmwarePatchKind is a patch version of a firmware
	FirmwarePatchKind = VersionKind{Type: "c8y_FirmwareBinary", Fragment: "c8y_Firmware", Patch: true}
)

// VersionWithoutParent returns the id of a version which is given without the
// software or firmware it belongs to (a version number, "1.0.0", and no id).
// As in v2.59.0 the first version with that number of any software or firmware
// is used. A version which is an id, or an empty version, is returned as is.
// An unknown version is an unresolved name (c8ystream.UnresolvedNameError).
func VersionWithoutParent(ctx context.Context, client *apiv2.Client, kind VersionKind, version string) (string, error) {
	if version == "" || c8ydata.IsID(version) {
		return version, nil
	}
	query := model.NewInventoryQuery().AddFilterEqStr("type", kind.Type)
	if kind.Patch {
		query = query.HasFragment("c8y_Patch")
	} else {
		query = query.AddFilterPart("not(has(c8y_Patch))")
	}
	res := client.ManagedObjects.List(apiv2.ResolutionContext(ctx), managedobjects.ListOptions{
		Query:             query.AddFilterEqStr(kind.Fragment+".version", version).Build(),
		PaginationOptions: pagination.PaginationOptions{PageSize: 5},
	})
	if res.Err != nil {
		return "", core.WrapLookup("version", version, res.Err)
	}
	for item := range res.Seq() {
		return item.ID(), nil
	}
	return "", c8ystream.UnresolvedNameError(version)
}
