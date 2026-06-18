// v2-based (deprecated) get child device reference: fetches a single childDevice
// reference of a parent device via ManagedObjects.ChildDevices.Get and returns
// the referenced managed object. The device flag drives iteration. Hidden and
// deprecated in favour of `c8y devices children get --childType device`.
package getchild

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/cobra"
)

// GetChildCmd command
type GetChildCmd struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

// NewGetChildCmd creates a command to Get child device reference
func NewGetChildCmd(f *cmdutil.Factory) *GetChildCmd {
	ccmd := &GetChildCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:    "getChild",
		Short:  "Get child device reference",
		Long:   `Get managed object child device reference`,
		Hidden: true,
		Example: heredoc.Doc(`
$ c8y devices getChild --device 12345 --reference 12345
Get an existing child device reference
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("device", []string{""}, "ManagedObject id (required) (accepts pipeline)")
	cmd.Flags().StringSlice("reference", []string{""}, "Device reference id (required)")

	completion.WithOptions(
		cmd,
		completion.WithDevice("device", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithDevice("reference", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
	)

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("device", "device", true, "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("device", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithPipelineAliases("reference", "deviceId", "source.id", "managedObject.id", "id"),
		flags.WithCollectionProperty("managedObject"),
		flags.WithPowershellName("Get-ChildDeviceReference"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedObjectReference+json", ""),
		flags.WithDeprecationNotice("please use 'c8y devices children get --childType device' instead"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("reference")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetChildCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}

	if err := r.InputFlag("device"); err != nil {
		return err
	}

	client, err := r.Client()
	if err != nil {
		return err
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		parentID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(in.String("device")), nil)
		if err != nil {
			return nil, err
		}
		childID, err := client.ManagedObjects.ResolveID(in.ResolveContext(), c8ystream.NameOrID(firstValue(in.StringSlice("reference"))), nil)
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context) output.Seq {
			return c8ystream.FromResult(client.ManagedObjects.ChildDevices.Get(ctx, parentID, childID))
		}, nil
	})
}

// firstValue returns the first non-empty entry of a slice flag (the path
// parameter takes a single value even though the flag is a slice).
func firstValue(values []string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
