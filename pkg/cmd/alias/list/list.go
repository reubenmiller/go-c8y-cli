package list

import (
	"fmt"
	"sort"

	"github.com/MakeNowJust/heredoc"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	Config func() (*config.Config, error)
	IO     *iostreams.IOStreams
}

func NewCmdList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		IO:     f.IOStreams,
		Config: f.Config,
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

	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	aliasCfg := cfg.Aliases()

	if len(aliasCfg) == 0 {
		if opts.IO.IsStdoutTTY() {
			fmt.Fprintf(opts.IO.ErrOut, "no aliases configured\n")
		}
		return nil
	}

	keys := []string{}
	for alias := range aliasCfg {
		keys = append(keys, alias)
	}
	sort.Strings(keys)

	w := opts.IO.Out

	fmt.Fprintf(w, "%s: %v\n", "select", cfg.GetJSONSelect())

	// TODO: Change to json writer
	for _, alias := range keys {
		_, err := fmt.Fprintf(w, "%s: %s\n", alias, aliasCfg[alias])
		if err != nil {
			return err
		}
	}

	return nil
}
