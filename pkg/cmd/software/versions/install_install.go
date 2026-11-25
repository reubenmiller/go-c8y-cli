// v2-based software version install: builds a device operation carrying a
// c8y_SoftwareUpdate entry (action "install"), resolves the target device
// (name -> id), and creates the operation via Operations.CreateRaw. The software
// type is looked up from the software repository when not given, and the version
// url/name are looked up from the software/version when no explicit url is
// provided (matching v1). The device flag is the iterating input, so the same
// software can be installed on many devices from a pipe.

package versions

import (
	"context"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/repodetails"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// InstallCmd command
type InstallCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewInstallCmd creates a command to Install software version on a device
func NewInstallCmd(f *cmdutil.Factory) *InstallCmd {
	ccmd := &InstallCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install software version on a device",
		Long:  `Install software version on a device`,
		Example: heredoc.Doc(`
$ c8y software versions install --device 1234 --software go-c8y-cli --version 1.0.0
Install a software package version
If the software/version exists in the software repository, then it will add the url automatically


$ c8y software versions install --device 1234 --software go-c8y-cli --version 1.0.0 --url "https://mybloblstore/go-c8y-cli.deb"
Install a software package version with an explicit url
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return f.CreateModeEnabled(cmd)
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "Device or agent where the software should be installed (accepts pipeline)")
	cmd.Flags().String("software", "", "Software name (required)")
	cmd.Flags().String("version", "", "Software version id or name")
	cmd.Flags().String("url", "", "Software url. Leave blank to automatically set it if a matching software/version is found in the c8y software repository")
	cmd.Flags().String("softwareType", "", "Software type. Leave blank to automatically set it if a matching software/version is found in the c8y software repository")
	cmd.Flags().String("description", "Install software package", "Operation description")
	cmd.Flags().String("action", "install", "Software action")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", ccmd.factory.ClientV2),
		completion.WithSoftware("software", ccmd.factory.ClientV2),
		completion.WithSoftwareVersion("version", "software", ccmd.factory.ClientV2),
		completion.WithValidateSet("action", "install"),
	)

	flags.WithOptions(
		cmd,
		flags.WithProcessingMode(),
		flags.WithData(),
		f.WithTemplateFlag(cmd),
		flags.WithExtendedPipelineSupport("device", "deviceId", false, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Install-SoftwareVersion"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.operation+json", ""),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("software")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *InstallCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("device"); err != nil {
		return err
	}

	err = r.Body(
		flags.WithDataFlagValue(),
		flags.WithStringValue("software", "c8y_SoftwareUpdate.0.name"),
		// version/url use WithAnyStringValue so an empty value is still written
		// (c8y_SoftwareUpdate.0.version: ""), matching v1 which always emitted
		// all keys via the computed softwareDetails type. WithStringValue skips
		// empties.
		flags.WithAnyStringValue("version", "c8y_SoftwareUpdate.0.version"),
		flags.WithAnyStringValue("url", "c8y_SoftwareUpdate.0.url"),
		flags.WithStringValue("softwareType", "c8y_SoftwareUpdate.0.softwareType"),
		flags.WithStringValue("description", "description"),
		flags.WithStringValue("action", "c8y_SoftwareUpdate.0.action"),
		cmdutil.WithTemplateValue(n.factory),
		flags.WithTemplateVariablesValue(),
	)
	if err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		device := in.String("device")
		if device == "" {
			return nil, cmderrors.NewUserError("Body is missing required properties: deviceId")
		}
		body, err := in.Body()
		if err != nil {
			return nil, err
		}
		if body, err = sjson.SetBytes(body, "deviceId", device); err != nil {
			return nil, err
		}

		// Look up the software type (when not provided) and the version url and
		// software name (when no explicit url is provided and a version was
		// given) from the repository, matching v1. A missing software is not an
		// error: the type stays unset.
		details, err := repodetails.LookupSoftware(in.ResolveContext(), client, repodetails.Software{
			Name:    in.String("software"),
			Version: in.String("version"),
			URL:     in.String("url"),
			Type:    in.String("softwareType"),
		})
		if err != nil {
			return nil, err
		}
		// v1 always emitted c8y_SoftwareUpdate.0.version/url (defaulting to "") via
		// the computed softwareDetails type, so they are always set.
		for _, kv := range [][2]string{
			{"c8y_SoftwareUpdate.0.name", details.Name},
			{"c8y_SoftwareUpdate.0.version", details.Version},
			{"c8y_SoftwareUpdate.0.url", details.URL},
		} {
			if body, err = sjson.SetBytes(body, kv[0], kv[1]); err != nil {
				return nil, err
			}
		}
		if details.Type != "" {
			if body, err = sjson.SetBytes(body, "c8y_SoftwareUpdate.0.softwareType", details.Type); err != nil {
				return nil, err
			}
		}

		// version is intentionally NOT required: an empty version is valid (v1
		// emitted c8y_SoftwareUpdate.0.version: "" and only checked key existence).
		if err := installRequireBodyKeys(body, "deviceId", "c8y_SoftwareUpdate.0.name", "c8y_SoftwareUpdate.0.action"); err != nil {
			return nil, err
		}

		return func(ctx context.Context) any {
			return client.Operations.CreateRaw(ctx, body)
		}, nil
	})
}

// requireBodyKeys returns a user error listing any of the given gjson paths that
// are missing or empty in the body (mirrors v1's bodyRequiredKeys check).
func installRequireBodyKeys(body []byte, keys ...string) error {
	var missing []string
	for _, k := range keys {
		if v := gjson.GetBytes(body, k); !v.Exists() || v.String() == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return cmderrors.NewUserError("Body is missing required properties: " + strings.Join(missing, ", "))
	}
	return nil
}
