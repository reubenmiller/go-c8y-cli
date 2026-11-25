package cmdparser

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extexec"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/firmware/firmwarepatches"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/firmware/firmwareversions"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/software/softwareversions"
	"github.com/spf13/cobra"
)

// references resolves the reference parameters of a runtime command (a
// device name, an application name, ...) with the go-c8y v2 client. The
// client is set when the command runs; the flag getters are built earlier
// (when the command is created), so they read it from here.
type references struct {
	client *apiv2.Client
	tenant string

	// ctx is the resolution context of the item being built, so lookups
	// reach the API even in a dry run.
	ctx context.Context
}

func (r *references) context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return apiv2.ResolutionContext(context.Background())
}

// resolveFunc resolves one reference value to the value the request uses
// (usually an id).
type resolveFunc func(ctx context.Context, r *references, cmd *cobra.Command, value string) (string, error)

// withReference returns a flag getter which resolves the flag's value with
// resolve. Only the first value is used (as the v1 "FirstMatch" getters did).
// For the pipeline flag, every input item is resolved when the request of the
// item is built; for any other flag the value is resolved once.
func (r *references) withReference(resolve resolveFunc, args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {
		src, dst, format := flags.UnpackGetterOptions("", opts...)

		if inputIterators != nil && inputIterators.PipeOptions != nil && inputIterators.PipeOptions.Name == src {
			pipeOpts := *inputIterators.PipeOptions
			// the format is applied after the lookup
			pipeOpts.Format = ""
			pipeIter, err := flags.NewFlagWithPipeIterator(cmd, &pipeOpts, true)
			if errors.Is(err, iterator.ErrEmptyPipeInput) && !pipeOpts.EmptyPipe {
				return pipeOpts.Property, nil, err
			}
			if err != nil || pipeIter == nil {
				return "", nil, err
			}
			return pipeOpts.Property, &referenceIterator{
				refs:     r,
				cmd:      cmd,
				values:   pipeIter,
				resolve:  resolve,
				format:   format,
				required: pipeOpts.Required,
			}, nil
		}

		value := firstReferenceValue(cmd, src, args)
		if value == "" {
			return "", nil, nil
		}
		resolved, err := r.resolve(resolve, cmd, value)
		if err != nil {
			return dst, nil, err
		}
		return dst, applyFormat(format, resolved), nil
	}
}

func (r *references) resolve(resolve resolveFunc, cmd *cobra.Command, value string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("reference lookup: the client is not configured")
	}
	resolved, err := resolve(r.context(), r, cmd, value)
	if err != nil {
		return "", err
	}
	if resolved == "" {
		return "", cmderrors.NewNoMatchesFoundError(value)
	}
	return resolved, nil
}

// referenceIterator resolves the values of the pipeline flag, one per input
// item.
type referenceIterator struct {
	refs     *references
	cmd      *cobra.Command
	values   iterator.Iterator
	resolve  resolveFunc
	format   string
	required bool
}

func (i *referenceIterator) IsBound() bool { return true }

func (i *referenceIterator) MarshalJSON() ([]byte, error) { return iterator.MarshalJSON(i) }

func (i *referenceIterator) GetNext() ([]byte, interface{}, error) {
	value, input, err := i.values.GetNext()
	if err != nil {
		return nil, input, err
	}
	ref := firstValue(string(value))
	if ref == "" {
		if i.required {
			return nil, input, cmderrors.NewNoMatchesFoundError(string(value))
		}
		return nil, input, nil
	}
	resolved, err := i.refs.resolve(i.resolve, i.cmd, ref)
	if err != nil {
		return nil, input, err
	}
	return []byte(applyFormat(i.format, resolved)), input, nil
}

// firstReferenceValue returns the first value of a reference flag (or the
// first positional argument), ignoring pipeline references ("-" and "-.path")
// which the pipeline handles.
func firstReferenceValue(cmd *cobra.Command, name string, args []string) string {
	values, err := flags.GetFlagStringValues(cmd, name)
	if err != nil {
		return ""
	}
	if i := cmd.ArgsLenAtDash(); i > -1 {
		args = args[:i]
	}
	for _, value := range append(values, args...) {
		if value == flags.FlagReadFromPipeText || strings.HasPrefix(value, flags.FlagReadFromPipeJSON) {
			continue
		}
		if v := firstValue(value); v != "" {
			return v
		}
	}
	return ""
}

// firstValue returns the first part of a comma separated value.
func firstValue(value string) string {
	for part := range strings.SplitSeq(value, ",") {
		if part = strings.TrimSpace(part); part != "" {
			return part
		}
	}
	return ""
}

func applyFormat(format, value string) string {
	if format == "" || !strings.Contains(format, "%") {
		return value
	}
	return fmt.Sprintf(format, value)
}

// dependentValue returns the first value of a flag which scopes a reference
// (e.g. the software of a software version).
func dependentValue(cmd *cobra.Command, name string) string {
	return firstReferenceValue(cmd, name, nil)
}

//
// Resolvers of the parameter types
//

func resolveManagedObject(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.ManagedObjects.ResolveID(ctx, c8ystream.NameOrID(value), nil)
}

func resolveAgent(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.Agents.ResolveID(ctx, c8ystream.NameOrID(value), nil)
}

func resolveDeviceGroup(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.DeviceGroups.ResolveID(ctx, c8ystream.NameOrID(value), nil)
}

func resolveSmartGroup(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.SmartGroups.ResolveID(ctx, c8ystream.NameOrID(value), nil)
}

func resolveDeviceProfile(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.DeviceProfiles.ResolveID(ctx, c8ystream.NameOrID(value), nil)
}

func resolveApplication(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.Applications.ResolveID(ctx, c8ystream.NameOrID(value), nil)
}

func resolveMicroservice(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.Microservices.ResolveID(ctx, c8ystream.NameOrID(value), nil)
}

func resolveUIPlugin(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.UIPlugins.ResolveID(ctx, c8ystream.NameOrID(value), nil)
}

func resolveSoftware(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.Repository.Software.ResolveID(ctx, c8ystream.NameOrID(value), nil)
}

func resolveFirmware(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.Repository.Firmware.ResolveID(ctx, c8ystream.NameOrID(value), nil)
}

func resolveConfiguration(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.Repository.Configuration.ResolveID(ctx, c8ystream.NameOrID(value), nil)
}

// resolveSoftwareVersion resolves a software version (an id, or a version of
// the software given by the softwareFlag)
func resolveSoftwareVersion(softwareFlag string) resolveFunc {
	return func(ctx context.Context, r *references, cmd *cobra.Command, value string) (string, error) {
		ref := versionRef(value, dependentValue(cmd, softwareFlag), softwareversions.NewRef().ByVersion, func(version, name string) string {
			return softwareversions.NewRef().ByVersionAndName(version, name)
		})
		return r.client.Repository.Software.Versions.ResolveID(ctx, ref, nil)
	}
}

// resolveFirmwareVersion resolves a firmware version (an id, or a version of
// the firmware given by the firmwareFlag)
func resolveFirmwareVersion(firmwareFlag string) resolveFunc {
	return func(ctx context.Context, r *references, cmd *cobra.Command, value string) (string, error) {
		ref := versionRef(value, dependentValue(cmd, firmwareFlag), firmwareversions.NewRef().ByVersion, firmwareversions.NewRef().ByVersionAndName)
		return r.client.Repository.Firmware.Versions.ResolveID(ctx, ref, nil)
	}
}

// resolveFirmwarePatch resolves a firmware patch (an id, or a patch version of
// the firmware given by the firmwareFlag)
func resolveFirmwarePatch(firmwareFlag string) resolveFunc {
	return func(ctx context.Context, r *references, cmd *cobra.Command, value string) (string, error) {
		ref := versionRef(value, dependentValue(cmd, firmwareFlag), firmwarepatches.NewRef().ByVersion, firmwarepatches.NewRef().ByVersionAndName)
		return r.client.Repository.Firmware.Patches.ResolveID(ctx, ref, nil)
	}
}

// versionRef returns the reference of a version scoped to its parent (a
// software or firmware id or name). An id, or a version without a parent, is
// used as is.
func versionRef(version, parent string, byParentID, byParentName func(version, parent string) string) string {
	version = c8ystream.FirstID(version)
	if version == "" || parent == "" || c8ydata.IsID(version) {
		return version
	}
	if c8ydata.IsID(parent) {
		return byParentID(version, parent)
	}
	return byParentName(version, parent)
}

func resolveUserGroup(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.UserGroups.ResolveID(ctx, r.tenant, c8ystream.NameOrID(value))
}

func resolveCertificate(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.TrustedCertificates.ResolveID(ctx, r.tenant, value)
}

// resolveUser returns the user id (the user name)
func resolveUser(_ context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.Users.UserID(value), nil
}

// resolveUserSelf returns the self link of a user
func resolveUserSelf(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.Users.UserSelfLink(r.tenant, r.client.Users.UserID(value)), nil
}

// resolveRole returns the role id (the role name). A name with wildcards is
// looked up.
func resolveRole(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	return r.client.UserRoles.ResolveID(ctx, value, nil)
}

// resolveRoleSelf returns the self link of a role. A name with wildcards is
// looked up.
func resolveRoleSelf(ctx context.Context, r *references, _ *cobra.Command, value string) (string, error) {
	id, err := r.client.UserRoles.ResolveID(ctx, value, nil)
	if err != nil {
		return "", err
	}
	return r.client.UserRoles.RoleSelfLink(id), nil
}

// resolveRemoteAccessConfiguration resolves a remote access configuration of
// the device given by the deviceFlag
func resolveRemoteAccessConfiguration(deviceFlag string) resolveFunc {
	return func(ctx context.Context, r *references, cmd *cobra.Command, value string) (string, error) {
		device := dependentValue(cmd, deviceFlag)
		if device == "" {
			return "", fmt.Errorf("device not found")
		}
		deviceID, err := r.client.ManagedObjects.ResolveID(ctx, c8ystream.NameOrID(device), nil)
		if err != nil {
			return "", err
		}
		return r.client.RemoteAccess.Configurations.ResolveID(ctx, deviceID, c8ystream.NameOrID(value))
	}
}

// resolveExternal resolves a value with an external command (the first line
// of its output). A value matching idPattern is used as is.
func resolveExternal(command []string, idPattern string) resolveFunc {
	var pattern *regexp.Regexp
	if idPattern != "" {
		pattern, _ = regexp.Compile(idPattern)
	}
	return func(_ context.Context, _ *references, _ *cobra.Command, value string) (string, error) {
		if pattern != nil && pattern.MatchString(value) {
			return value, nil
		}
		output, err := extexec.ExecuteExternalCommand(value, command)
		if err != nil {
			return "", fmt.Errorf("could not fetch by name. %w", err)
		}
		for row := range bytes.SplitSeq(output, []byte("\n")) {
			if len(row) > 0 {
				return string(row), nil
			}
		}
		return "", nil
	}
}
