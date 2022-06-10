package delete

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

type DeleteOptions struct {
	IO      *iostreams.IOStreams
	factory *cmdutil.Factory

	Name string
}

func NewCmdDelete(f *cmdutil.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{
		IO:      f.IOStreams,
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "delete <alias>",
		Short: "Delete an alias",
		Args:  flags.ExactArgsOrExample(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 1 {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}
			cfg, err := opts.factory.Config()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			keys := make([]string, 0)
			for key := range cfg.Aliases() {
				keys = append(keys, key)
			}
			return keys, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]

			if runF != nil {
				return runF(opts)
			}
			return deleteRun(opts)
		},
	}

	return cmd
}

func deleteRun(opts *DeleteOptions) error {
	cfg, err := opts.factory.Config()
	if err != nil {
		return err
	}

	aliasCfg := cfg.Aliases()

	expansion, ok := aliasCfg[opts.Name]
	if !ok {
		return fmt.Errorf("no such alias %s", opts.Name)
	}

	delete(aliasCfg, opts.Name)
	cfg.SetAliases(aliasCfg)
	if err := cfg.WritePersistentConfig(); err != nil {
		return err
	}

	if opts.IO.IsStdoutTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut, "%s Deleted alias %s; was %s\n", cs.SuccessIconWithColor(cs.Red), opts.Name, expansion)
	}

	return nil
}
