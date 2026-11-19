// v2-based inventory count: the type flag is the iterating input; each item
// builds a managed-object filter and returns the total count via
// ManagedObjects.Count (GET /inventory/managedObjects/count).
package count

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// CountCmd command
type CountCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewCountCmd creates a command to Get managed object count
func NewCountCmd(f *cmdutil.Factory) *CountCmd {
	ccmd := &CountCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Get managed object count",
		Long:  `Retrieve the total number of managed objects (e.g. devices, assets, etc.) registered in your tenant, or a subset based on queries.`,
		Example: heredoc.Doc(`
$ c8y inventory count
Get count of managed objects

$ c8y inventory count --text myname
Get count of managed objects matching text (using Cumulocity text search algorithm)

$ c8y inventory count --type "c8y_Sensor"
Get count of managed objects with a specific type value

$ c8y inventory count --type "c8y_Sensor" --owner "device_mylinuxbox01"
Get count of managed objects with a specific type value and owner

$ c8y inventory count --fragmentType "c8y_IsDevice"
Get total number of devices
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("ids", []string{""}, "List of ids.")
	cmd.Flags().String("type", "", "ManagedObject type. (accepts pipeline)")
	cmd.Flags().String("fragmentType", "", "ManagedObject fragment type.")
	cmd.Flags().String("owner", "", "List of managed objects that are owned by the given username.")
	cmd.Flags().String("text", "", "Search for managed objects where a property value is equal to the given one. The following properties are examined: id, type, name, owner, externalIds.")
	cmd.Flags().String("childAdditionId", "", "Search for a specific child addition and list all the groups to which it belongs.")
	cmd.Flags().String("childAssetId", "", "Search for a specific child asset and list all the groups to which it belongs.")
	cmd.Flags().StringSlice("childDeviceId", []string{""}, "Search for a specific child device and list all the groups to which it belongs.")

	completion.WithOptions(
		cmd,
		completion.WithDevice("childDeviceId", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("type", "type", false, "type"),
		flags.WithPipelineAliases("childDeviceId", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPowershellName("Get-ManagedObjectCount"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedobjectuser+json", ""),
	)

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *CountCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("type"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		childDeviceID := firstNonEmpty(in.StringSlice("childDeviceId"))
		if childDeviceID != "" {
			childDeviceID, err = client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(childDeviceID), nil)
			if err != nil {
				return nil, err
			}
		}

		opt := managedobjects.ListOptions{
			Ids:             nonEmpty(in.StringSlice("ids")),
			Type:            in.String("type"),
			FragmentType:    in.String("fragmentType"),
			Owner:           in.String("owner"),
			Text:            in.String("text"),
			ChildAdditionID: in.String("childAdditionId"),
			ChildAssetID:    in.String("childAssetId"),
			ChildDeviceID:   childDeviceID,
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromJSONValue(client.ManagedObjects.Count(ctx, opt), apiv2.IsDryRun(ctx))
		}, nil
	})
}

// firstNonEmpty returns the first non-empty entry of a slice flag.
func firstNonEmpty(values []string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// nonEmpty drops empty entries from a slice flag (the StringSlice default is
// `[""]`, which would otherwise serialise as an empty ids filter).
func nonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
