package cmdparser

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ybinary"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ydata"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/repodetails"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/pagination"
	"github.com/spf13/cobra"
)

// flagValue returns the first value of a flag, or "" if it is not set
func flagValue(cmd *cobra.Command, name string) string {
	if name == "" {
		return ""
	}
	if v, err := flags.GetFlagStringValues(cmd, name); err == nil && len(v) > 0 {
		return v[0]
	}
	return ""
}

// withSoftwareDetails sets the details (name, version, url and the software
// type) of a software version. The details of a version which is not given
// with a url are looked up in the software repository.
func (r *references) withSoftwareDetails(flagSoftware, flagVersion, flagURL, flagSoftwareType string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, _ *flags.RequestInputIterators) (string, interface{}, error) {
		if r.client == nil {
			return "", nil, fmt.Errorf("reference lookup: the client is not configured")
		}
		_, dst, _ := flags.UnpackGetterOptions("", opts...)

		details, err := repodetails.LookupSoftware(r.context(), r.client, repodetails.Software{
			Name:    flagValue(cmd, flagSoftware),
			Version: flagValue(cmd, flagVersion),
			URL:     flagValue(cmd, flagURL),
			Type:    flagValue(cmd, flagSoftwareType),
		})
		if err != nil {
			return "", nil, noMatches(err, flagVersion)
		}

		output := map[string]string{}
		if details.Type != "" {
			output["softwareType"] = details.Type
		}
		output["name"] = details.Name
		output["version"] = details.Version
		output["url"] = details.URL
		return dst, output, nil
	}
}

// withFirmwareDetails sets the details (name, version and url) of a firmware
// version. The details of a version which is not given with a url are looked up
// in the firmware repository.
func (r *references) withFirmwareDetails(flagFirmware, flagVersion, flagURL string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, _ *flags.RequestInputIterators) (string, interface{}, error) {
		if r.client == nil {
			return "", nil, fmt.Errorf("reference lookup: the client is not configured")
		}
		_, dst, _ := flags.UnpackGetterOptions("", opts...)

		details, err := repodetails.LookupFirmware(r.context(), r.client, repodetails.Firmware{
			Name:    flagValue(cmd, flagFirmware),
			Version: flagValue(cmd, flagVersion),
			URL:     flagValue(cmd, flagURL),
		})
		if err != nil {
			return "", nil, noMatches(err, flagVersion)
		}

		return dst, map[string]string{
			"name":    details.Name,
			"version": details.Version,
			"url":     details.URL,
		}, nil
	}
}

// withConfigurationDetails sets the details (type, url and name) of a
// configuration file item, looked up by id or name
func (r *references) withConfigurationDetails(flagConfiguration, flagConfigurationType, flagURL string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, _ *flags.RequestInputIterators) (string, interface{}, error) {
		if r.client == nil {
			return "", nil, fmt.Errorf("reference lookup: the client is not configured")
		}
		_, dst, _ := flags.UnpackGetterOptions("", opts...)

		details, err := repodetails.LookupConfiguration(r.context(), r.client, repodetails.Configuration{
			Name: flagValue(cmd, flagConfiguration),
			Type: flagValue(cmd, flagConfigurationType),
			URL:  flagValue(cmd, flagURL),
		})
		if err != nil {
			return "", nil, noMatches(err, flagConfiguration)
		}

		return dst, map[string]string{
			"type": details.Type,
			"url":  details.URL,
			"name": details.Name,
		}, nil
	}
}

// noMatches maps the error of a lookup of an item which does not exist to the
// error of a reference which was not found
func noMatches(err error, flagName string) error {
	if apiv2.IsNotFound(err) || errors.Is(err, repodetails.ErrNoMatch) {
		return cmderrors.NewNoMatchesFoundError(flagName)
	}
	return err
}

// resolveDeviceService resolves a service of the device given by the
// deviceFlag. An id is used as is, otherwise the first service with the name
// (of the device if it is found) is used.
func resolveDeviceService(deviceFlag string) resolveFunc {
	return func(ctx context.Context, r *references, cmd *cobra.Command, value string) (string, error) {
		if isID(value) {
			return value, nil
		}

		device := dependentValue(cmd, deviceFlag)
		if device != "" && !isID(device) {
			id, err := r.client.ManagedObjects.ResolveID(ctx, c8ystream.NameOrID(device), nil)
			switch {
			case err == nil:
				device = id
			case apiv2.IsNotFound(err):
				// look for the service of any device
				device = ""
			default:
				return "", err
			}
		}

		query := fmt.Sprintf("(type eq 'c8y_Service') and name eq '%s'", value)
		if isID(device) {
			query = fmt.Sprintf("(type eq 'c8y_Service') and name eq '%s' and (bygroupid(%s))", value, device)
		}
		items, err := r.client.ManagedObjects.List(ctx, managedobjects.ListOptions{
			Query:             query,
			PaginationOptions: pagination.PaginationOptions{PageSize: 5},
		}).Collect()
		if err != nil {
			return "", err
		}
		if len(items) == 0 {
			return "", nil
		}
		return items[0].ID(), nil
	}
}

// isID reports whether the value is a managed object id (digits only)
func isID(v string) bool {
	return v != "" && c8ydata.IsID(v)
}

// isDryRun reports whether the command runs as a dry run
func isDryRun(factory *cmdutil.Factory, cmd *cobra.Command) bool {
	cfg, err := factory.Config()
	if err != nil {
		return false
	}
	return cfg.ShouldUseDryRun(cmd.CommandPath())
}

// withBinaryUploadURL uploads a binary (the file of the flag) and sets the url
// of the uploaded binary. Nothing is uploaded in a dry run: the url is only
// shown as a placeholder.
func (r *references) withBinaryUploadURL(progress *iostreams.Progress, dryRun func(*cobra.Command) bool, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, _ *flags.RequestInputIterators) (string, interface{}, error) {
		src, dst, _ := flags.UnpackGetterOptions("%s", opts...)

		if !cmd.Flags().Changed(src) {
			return "", "", nil
		}

		filename, err := cmd.Flags().GetString(src)
		if err != nil {
			return dst, filename, err
		}
		if r.client == nil {
			return "", nil, fmt.Errorf("binary upload: the client is not configured")
		}

		if dryRun != nil && dryRun(cmd) {
			return dst, "{host}/inventory/binaries/{id}", nil
		}

		binary, err := c8ybinary.UploadBinary(r.context(), r.client, progress, filename)
		if err != nil {
			return "", nil, err
		}
		return dst, binary.Self(), nil
	}
}

// withIDSlice adds an id slice from the flag and the cli arguments. The values
// can be piped.
func withIDSlice(args []string, opts ...string) flags.GetOption {
	return func(cmd *cobra.Command, inputIterators *flags.RequestInputIterators) (string, interface{}, error) {

		src, dst, _ := flags.UnpackGetterOptions("", opts...)

		// check for arguments which could override the value
		values, err := cmd.Flags().GetStringSlice(src)
		if err != nil {
			singleValue, singleErr := cmd.Flags().GetString(src)
			if singleErr != nil {
				return "", "", singleErr
			}
			err = nil
			values = []string{singleValue}
		}

		values = parseValues(append(values, args...))

		var overrideValue iterator.Iterator
		if len(values) > 0 {
			overrideValue = iterator.NewSliceIterator(values)
		}

		if inputIterators != nil && inputIterators.PipeOptions.Name == src {
			hasPipeSupport := inputIterators.PipeOptions.Name == src
			pipeIter, err := flags.NewFlagWithPipeIterator(cmd, inputIterators.PipeOptions, hasPipeSupport)

			if err == iterator.ErrEmptyPipeInput && !inputIterators.PipeOptions.EmptyPipe {
				return inputIterators.PipeOptions.Property, nil, err
			}

			if err != nil || pipeIter == nil {
				return "", nil, err
			}

			if pipeIter.IsBound() {
				// Use infinite slice iterator so that the stdin can drive the iteration
				// but only if the other pipe iterator is bound, otherwise it would create an infinite loop!
				overrideValue = iterator.NewInfiniteSliceIterator(values)
			}

			iter := iterator.NewOverrideIterator(pipeIter, overrideValue)
			return inputIterators.PipeOptions.Property, iter, nil
		}

		if len(values) == 0 {
			return "", values, nil
		}

		return dst, values, err
	}
}

// parseValues splits the comma separated values, ignoring empty values and the
// pipeline references
func parseValues(values []string) (ids []string) {
	for _, value := range values {
		for part := range strings.SplitSeq(value, ",") {
			// Only add uint looking values, and filter out custom pipeline mapped flags
			if part != "" && part != flags.FlagReadFromPipeText && !strings.HasPrefix(part, flags.FlagReadFromPipeJSON) {
				ids = append(ids, strings.TrimSpace(part))
			}
		}
	}
	return
}
