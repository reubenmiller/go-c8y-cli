package set

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/shlex"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

type SetOptions struct {
	factory *cmdutil.Factory
	IO      *iostreams.IOStreams

	Name      string
	Expansion string
	IsShell   bool
	RootCmd   *cobra.Command
}

func NewCmdSet(f *cmdutil.Factory, runF func(*SetOptions) error) *cobra.Command {
	opts := &SetOptions{
		IO:      f.IOStreams,
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "set <alias> <expansion>",
		Short: "Create a shortcut for a c8y command",
		Long: heredoc.Doc(`
			Declare a word as a command alias that will expand to the specified command(s).
			The expansion may specify additional arguments and flags. If the expansion
			includes positional placeholders such as '$1', '$2', etc., any extra arguments
			that follow the invocation of an alias will be inserted appropriately.
			If '--shell' is specified, the alias will be run through a shell interpreter (sh). This allows you
			to compose commands with "|" or redirect with ">". Note that extra arguments following the alias
			will not be automatically passed to the expanded expression. To have a shell alias receive
			arguments, you must explicitly accept them using "$1", "$2", etc., or "$@" to accept all of them.
			Platform note: on Windows, shell aliases are executed via "sh" as installed by Git For Windows. If
			you have installed git on Windows in some other way, shell aliases may not work for you.
			Quotes must always be used when defining a command as in the examples.
		`),
		Example: heredoc.Doc(`
			$ c8y alias set createTestDevice 'devices create --template test.device.json'
			$ c8y createTestDevice
			#=> c8y devices create --template test.device.json

			$ c8y alias set listAlarmsByType 'alarms list --type="$1"'
			$ c8y listAlarmsByType myType
			#=> c8y alarms list --type="myType"

			$ c8y alias set recentEvents 'operations list --dateFrom="-1h"'
			$ c8y recentEvents
			#=> c8y operations list --dateFrom="-1h"

			$ c8y alias set --shell findInAudit 'c8y auditrecords list --dateFrom="-30d" --view auto --includeAll | grep $1'
			$ c8y findInAudit deleted
			#=> c8y auditrecords list --dateFrom="-30d" --output json --view auto --includeAll | grep deleted
		`),
		Args: flags.ExactArgsOrExample(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.RootCmd = cmd.Root()

			opts.Name = args[0]
			opts.Expansion = args[1]

			if runF != nil {
				return runF(opts)
			}
			return setRun(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.IsShell, "shell", "s", false, "Declare an alias to be passed through a shell interpreter")

	return cmd
}

func setRun(opts *SetOptions) error {
	cfg, err := opts.factory.Config()
	if err != nil {
		return err
	}
	cs := opts.IO.ColorScheme()
	aliasCfg := cfg.Aliases()

	isTerminal := opts.IO.IsStdoutTTY()
	if isTerminal {
		fmt.Fprintf(opts.IO.ErrOut, "- Adding alias for %s: %s\n", cs.Bold(opts.Name), cs.Bold(opts.Expansion))
	}

	expansion := opts.Expansion
	isShell := opts.IsShell
	if isShell && !strings.HasPrefix(expansion, "!") {
		expansion = "!" + expansion
	}
	isShell = strings.HasPrefix(expansion, "!")

	if validCommand(opts.RootCmd, opts.Name) {
		return fmt.Errorf("could not create alias: %q is already a c8y command", opts.Name)
	}

	if !isShell && !validCommand(opts.RootCmd, expansion) {
		return fmt.Errorf("could not create alias: %s does not correspond to a c8y command", expansion)
	}

	successMsg := fmt.Sprintf("%s Added alias.", cs.SuccessIcon())
	if oldExpansion, ok := aliasCfg[opts.Name]; ok {
		successMsg = fmt.Sprintf("%s Changed alias %s from %s to %s",
			cs.SuccessIcon(),
			cs.Bold(opts.Name),
			cs.Bold(oldExpansion),
			cs.Bold(expansion),
		)
	}

	aliasCfg[opts.Name] = expansion
	cfg.SetAliases(aliasCfg)
	if err := cfg.WritePersistentConfig(); err != nil {
		return err
	}

	if isTerminal {
		fmt.Fprintln(opts.IO.ErrOut, successMsg)
	}

	return nil
}

func validCommand(rootCmd *cobra.Command, expansion string) bool {
	split, err := shlex.Split(expansion)
	if err != nil {
		return false
	}

	cmd, _, err := rootCmd.Traverse(split)
	return err == nil && cmd != rootCmd
}
