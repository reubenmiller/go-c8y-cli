package list

import (
	"fmt"
	"io"
	"sort"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	IO      *iostreams.IOStreams
	factory *cmdutil.Factory
}

func NewCmdList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		IO:      f.IOStreams,
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your aliases",
		Long: heredoc.Doc(`
			This command prints out all of the aliases gh is configured to use.
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
		},
	}

	cmdutil.DisableAuthCheck(cmd)

	return cmd
}

func listRun(opts *ListOptions) error {

	cfg, err := opts.factory.Config()
	if err != nil {
		return err
	}

	aliasCfg := cfg.Aliases()
	commonAliasCfg := cfg.CommonAliases()

	if len(aliasCfg) == 0 && len(commonAliasCfg) == 0 {
		if opts.IO.IsStdoutTTY() {
			fmt.Fprintf(opts.IO.ErrOut, "no aliases configured\n")
		}
		return nil
	}

	w := opts.IO.Out

	err = printAliases(w, opts.IO.ColorScheme(), "session aliases", aliasCfg)
	if err != nil {
		return err
	}

	err = printAliases(w, opts.IO.ColorScheme(), "common aliases", commonAliasCfg)
	if err != nil {
		return err
	}

	return nil
}

func printAliases(w io.Writer, cs *iostreams.ColorScheme, title string, aliases map[string]string) error {
	if len(aliases) == 0 {
		return nil
	}
	keys := []string{}
	for alias := range aliases {
		keys = append(keys, alias)
	}
	sort.Strings(keys)

	fmt.Fprintf(w, "\n%s\n", cs.Bold(cs.Magenta(title)))

	// TODO: Change to json writer
	for _, alias := range keys {
		_, err := fmt.Fprintf(w, "%s: %s\n", cs.CyanBold(alias), aliases[alias])
		if err != nil {
			return err
		}
	}
	return nil
}
