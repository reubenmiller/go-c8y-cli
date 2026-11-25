package parents

import (
	"context"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/inventory/managedobjects"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsonmodels"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/spf13/cobra"
)

// GetCmd command
type GetCmd struct {
	*subcommand.SubCommand

	Type    string
	Level   int64
	All     bool
	Reverse bool

	factory *cmdutil.Factory
}

// NewGetCmd creates a command to Get addition parent
func NewGetCmd(f *cmdutil.Factory) *GetCmd {
	ccmd := &GetCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get managed object parent",
		Long:  `Get managed object parent for relationship types (e.g. addition, asset, device)`,
		Example: heredoc.Doc(`
			$ c8y inventory parents get --id 12345 --type addition
			Get parent of the child addition

			$ c8y inventory parents get --id 12345 --type asset
			Get parent of the child asset (e.g. usually the group)

			$ c8y devices list | c8y inventory parents get --type device --level 2
			Get the grandparent (parent of the parent)

			$ c8y devices list | c8y inventory parents get --type device --level -1
			Get the root parent device of a list of devices

			$ c8y devices get --id 12345 | c8y inventory parents get --type device -all
			Get all parents (parent, grandparents, grandparents etc.) of a single device returning the root parent last

			$ c8y devices get --id 12345 | c8y inventory parents get --type device -all --reverse
			Get all parents (parent, grandparents, grandparents etc.) of a single device returning the root parent first
        `),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringSlice("id", []string{""}, "ManagedObject id (required) (accepts pipeline)")
	cmd.Flags().BoolVar(&ccmd.All, "all", false, "Return all parents in the chain")
	cmd.Flags().BoolVar(&ccmd.Reverse, "reverse", false, "Return all parents in order from root to parent")
	cmd.Flags().StringVar(&ccmd.Type, "type", "", "Type of relationship, e.g. addition, asset, device")
	cmd.Flags().Int64Var(&ccmd.Level, "level", 1, "Number of parent jumps to do. 0 = current item, 1 = parent, 2 = grandparent etc. Defaults to 1. Use -1 for last parent")
	cmd.Flags().Bool("withParents", true, "include a flat list of all parents and grandparents of the given object")

	// Hide withParents as it is fixed
	_ = cmd.Flags().MarkHidden("withParents")

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("type", "addition", "asset", "device"),
	)

	flags.WithOptions(
		cmd,

		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id", "managedObject.id"),
	)

	// Required flags

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *GetCmd) RunE(cmd *cobra.Command, args []string) error {
	if n.Type == "" {
		return &flags.ParameterError{
			Name: "type",
			Err:  flags.ErrParameterMissing,
		}
	}
	path := parentsPath(n.Type, n.All, n.Reverse, n.Level)

	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	if err := r.InputFlag("id"); err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		ref := c8ystream.NameOrID(in.String("id"))
		return func(ctx context.Context) any {
			res := client.ManagedObjects.Get(ctx, ref, managedobjects.GetOptions{WithParents: true})
			return parents(res, path)
		}, nil
	})
}

// parentsPath is the path of the wanted parents in a managed object fetched
// with withParents, e.g. "assetParents.references.0.managedObject" for the
// first asset parent. An empty path is the managed object itself.
func parentsPath(typ string, all, reverse bool, level int64) string {
	refs := strings.ToLower(typ) + "Parents.references"
	switch {
	case all && reverse:
		return refs + "|@reverse|#.managedObject"
	case all:
		return refs + ".#.managedObject"
	case level == 0:
		return ""
	case level > 0:
		return fmt.Sprintf("%s.%d.managedObject", refs, level-1)
	default:
		// Python style indexing: -1 is the last parent, -2 the one before.
		return fmt.Sprintf("%s|@reverse|%d.managedObject", refs, -level-1)
	}
}

// parents yields the managed objects at path in the result (one per element
// when it is a list). A dry run yields nothing.
func parents(res op.Result[jsonmodels.ManagedObject], path string) func(yield func(jsondoc.JSONDoc, error) bool) {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		if res.Err != nil {
			yield(jsondoc.Empty(), res.Err)
			return
		}
		if res.Request != nil || res.IsDeferred() {
			return
		}
		value := res.Data.GetJSONDoc().Get(path)
		if !value.Exists() {
			return
		}
		if value.IsArray() {
			for _, item := range value.Array() {
				if !yield(jsondoc.New([]byte(item.Raw)), nil) {
					return
				}
			}
			return
		}
		yield(jsondoc.New([]byte(value.Raw)), nil)
	}
}
