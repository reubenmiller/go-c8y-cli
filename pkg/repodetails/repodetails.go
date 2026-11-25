// Package repodetails looks up the details of the software, firmware and
// configuration items of the Cumulocity repositories (name, version, url,
// ...) with the go-c8y v2 client. It is shared by the commands which install
// them on a device and by the software/firmware/configuration parameter types
// of the extension (runtime) commands, so both build the same fragments.
package repodetails

import (
	"context"
	"errors"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/configuration"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/firmware/firmwareversions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/software/softwareitems"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/software/softwareversions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/source"
)

// ErrNoMatch is returned when no item of the repository matches
var ErrNoMatch = errors.New("no matching item found")

// Software is the details of a software version (the c8y_Software fragment of
// a software version, and the entry of a software update operation)
type Software struct {
	Name    string
	Version string
	URL     string
	// Type is the type of the software (e.g. apt)
	Type string
}

// LookupSoftware completes the details of a software version from the
// software repository.
//
// The software type is taken from the software when it is not given (a
// software which does not exist is not an error). The version, url and name
// are taken from the software version when the version is not given with a
// url (the version can be a managed object id, or a version of the software).
// A version which does not exist is an error (see apiv2.IsNotFound).
func LookupSoftware(ctx context.Context, client *apiv2.Client, in Software) (Software, error) {
	out := in

	if out.Type == "" && out.Name != "" {
		res := client.Repository.Software.Get(ctx, source.ParseRef(out.Name), softwareitems.GetOptions{})
		if res.Err != nil {
			if !apiv2.IsNotFound(res.Err) {
				return out, res.Err
			}
		} else if t := res.Data.SoftwareType(); t != "" {
			out.Type = t
		}
	}

	if needsVersionLookup(in.Name, in.Version, in.URL) {
		ref := SoftwareVersionRef(in.Version, in.Name)
		res := client.Repository.Software.Versions.Get(ctx, ref, softwareversions.GetOptions{WithParents: true})
		if res.Err != nil {
			return out, res.Err
		}
		if v := res.Data.Version(); v != "" {
			out.Version = v
		}
		if u := res.Data.URL(); u != "" {
			out.URL = u
		}
		if name := res.Data.SoftwareName(); name != "" {
			out.Name = name
		}
	}
	return out, nil
}

// Firmware is the details of a firmware version (the c8y_Firmware fragment of
// a firmware version, and the entry of a firmware update operation)
type Firmware struct {
	Name    string
	Version string
	URL     string
}

// LookupFirmware completes the details of a firmware version from the firmware
// repository: the version, url and name are taken from the firmware version
// when the version is not given with a url (the version can be a managed object
// id, or a version of the firmware). A version which does not exist is an error
// (see apiv2.IsNotFound).
func LookupFirmware(ctx context.Context, client *apiv2.Client, in Firmware) (Firmware, error) {
	out := in
	if !needsVersionLookup(in.Name, in.Version, in.URL) {
		return out, nil
	}
	ref := FirmwareVersionRef(in.Version, in.Name)
	res := client.Repository.Firmware.Versions.Get(ctx, ref, firmwareversions.GetOptions{WithParents: true})
	if res.Err != nil {
		return out, res.Err
	}
	if v := res.Data.Version(); v != "" {
		out.Version = v
	}
	if u := res.Data.URL(); u != "" {
		out.URL = u
	}
	if name := res.Data.FirmwareName(); name != "" {
		out.Name = name
	}
	return out, nil
}

// needsVersionLookup reports whether the details of a version must be looked
// up: not when there is no version, nor when everything is already given.
func needsVersionLookup(name, version, url string) bool {
	return !(version == "" || (name != "" && url != ""))
}

// SoftwareVersionRef returns the reference of a software version: a managed
// object id (the version itself) is used as is, otherwise the version is scoped
// to its software (an id or a name).
func SoftwareVersionRef(version, software string) string {
	if version == "" || software == "" || c8ydata.IsID(version) {
		return version
	}
	if c8ydata.IsID(software) {
		return softwareversions.NewRef().ByVersion(version, software)
	}
	return softwareversions.NewRef().ByVersionAndName(version, software)
}

// FirmwareVersionRef returns the reference of a firmware version: a managed
// object id (the version itself) is used as is, otherwise the version is scoped
// to its firmware (an id or a name).
func FirmwareVersionRef(version, firmware string) string {
	if version == "" || firmware == "" || c8ydata.IsID(version) {
		return version
	}
	if c8ydata.IsID(firmware) {
		return firmwareversions.NewRef().ByVersion(version, firmware)
	}
	return firmwareversions.NewRef().ByVersionAndName(version, firmware)
}

// Configuration is the details of a configuration file item
type Configuration struct {
	Name string
	Type string
	URL  string
}

// LookupConfiguration looks up a configuration file item by id or by name
// (optionally of the given configuration type). When no name is given but the
// type and url are, they are used as they are. It returns ErrNoMatch when no
// configuration matches.
func LookupConfiguration(ctx context.Context, client *apiv2.Client, in Configuration) (Configuration, error) {
	if in.Name == "" && in.Type != "" && in.URL != "" {
		return in, nil
	}

	// An explicit managed object id
	if isID(in.Name) {
		res := client.Repository.Configuration.Get(ctx, in.Name, configuration.GetOptions{})
		if res.Err != nil {
			return in, res.Err
		}
		return Configuration{
			Name: res.Data.Name(),
			Type: res.Data.ConfigurationType(),
			URL:  res.Data.URL(),
		}, nil
	}

	items, err := client.Repository.Configuration.List(ctx, configuration.ListOptions{
		Name:              in.Name,
		ConfigurationType: in.Type,
		PaginationOptions: pagination.PaginationOptions{PageSize: 5},
	}).Collect()
	if err != nil {
		return in, err
	}
	if len(items) == 0 {
		return in, fmt.Errorf("configuration %q: %w", in.Name, ErrNoMatch)
	}
	return Configuration{
		Name: items[0].Name(),
		Type: items[0].ConfigurationType(),
		URL:  items[0].URL(),
	}, nil
}

// isID reports whether the value is a managed object id (digits only)
func isID(v string) bool {
	return v != "" && c8ydata.IsID(v)
}
