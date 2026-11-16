package root

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/spf13/cobra"
)

// annotationOutput marks extension commands whose output is processed by c8y (value "json")
const annotationOutput = "c8y.extension.output"

// outputFlag is a c8y output flag which is handled by c8y for extension commands with json output
type outputFlag struct {
	name      string
	shorthand string
	hasValue  bool
}

// outputFlags are not passed to extension commands which produce json, as c8y applies them
var outputFlags = []outputFlag{
	{name: "output", shorthand: "o", hasValue: true},
	{name: "select", hasValue: true},
	{name: "filter", hasValue: true},
	{name: "view", hasValue: true},
	{name: "compact", shorthand: "c"},
	{name: "raw", shorthand: "r"},
	{name: "noColor", shorthand: "M"},
}

// matchOutputFlag returns the flag matching the argument, and whether the value is part of the
// argument (--name=value or -ovalue)
func matchOutputFlag(arg string) (*outputFlag, bool) {
	for i, flag := range outputFlags {
		if strings.HasPrefix(arg, "--") {
			name, _, inline := strings.Cut(arg[2:], "=")
			if name == flag.name {
				return &outputFlags[i], inline
			}
			continue
		}
		if flag.shorthand != "" && len(arg) >= 2 && arg[0] == '-' && arg[1:2] == flag.shorthand {
			if len(arg) == 2 {
				return &outputFlags[i], false
			}
			// -ocsv or -o=csv, or -c=false
			return &outputFlags[i], true
		}
	}
	return nil, false
}

// extractOutputFlags splits args into the c8y output flags (with their values) and the
// remaining arguments. Arguments after "--" are not checked.
func extractOutputFlags(args []string) (flags []string, rest []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			rest = append(rest, args[i:]...)
			break
		}
		flag, inline := matchOutputFlag(arg)
		if flag == nil {
			rest = append(rest, arg)
			continue
		}
		flags = append(flags, arg)
		if flag.hasValue && !inline && i+1 < len(args) {
			flags = append(flags, args[i+1])
			i++
		}
	}
	return flags, rest
}

// findChild returns the subcommand matching the name or one of its aliases
func findChild(cmd *cobra.Command, name string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if strings.EqualFold(c.Name(), name) {
			return c
		}
		for _, alias := range c.Aliases {
			if strings.EqualFold(alias, name) {
				return c
			}
		}
	}
	return nil
}

// PrepareExtensionArgs prepares the arguments of an extension command before it is executed.
// Extension commands do not parse c8y flags, so:
//   - c8y flags before the command (e.g. c8y --session prod.json myext list) are parsed by c8y
//     and removed. They are passed to the extension via the runtime context.
//   - for commands with json output, the c8y output flags (e.g. --output, --select) are parsed by c8y
//     and removed, as c8y formats the output.
//
// The arguments are returned unchanged for all other commands.
func PrepareExtensionArgs(rootCmd *cobra.Command, args []string) []string {
	target, _, err := rootCmd.Find(args)
	if err != nil || target == rootCmd || !IsExtensionCommand(target) {
		return args
	}

	n := countLeadingFlags(rootCmd.PersistentFlags(), args)

	// The command names are expected directly after the leading flags
	cmd := rootCmd
	i := n
	for ; i < len(args) && cmd != target; i++ {
		next := findChild(cmd, args[i])
		if next == nil {
			return args
		}
		cmd = next
	}
	if cmd != target {
		return args
	}

	c8yFlags := append([]string{}, args[:n]...)
	rest := args[i:]
	if target.Annotations[annotationOutput] == extensions.OutputJSON {
		var output []string
		output, rest = extractOutputFlags(rest)
		c8yFlags = append(c8yFlags, output...)
	}

	if len(c8yFlags) > 0 {
		if err := rootCmd.ParseFlags(c8yFlags); err != nil {
			return args
		}
	}

	prepared := make([]string, 0, len(args))
	prepared = append(prepared, args[n:i]...)
	prepared = append(prepared, rest...)
	return prepared
}

// splitJSONValues splits the output of a command into json values. The output can be a json
// array, a json object or newline delimited json values. ok is false if the output is not json
func splitJSONValues(output []byte) (values [][]byte, isArray bool, ok bool) {
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		return nil, false, true
	}

	if trimmed[0] == '[' && json.Valid(trimmed) {
		var items []json.RawMessage
		if err := json.Unmarshal(trimmed, &items); err != nil {
			return nil, false, false
		}
		for _, item := range items {
			values = append(values, item)
		}
		return values, true, true
	}

	if json.Valid(trimmed) {
		return [][]byte{trimmed}, false, true
	}

	// Newline delimited json
	scanner := bufio.NewScanner(bytes.NewReader(trimmed))
	scanner.Buffer(make([]byte, 64*1024), 64*1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		if !json.Valid(line) {
			return nil, false, false
		}
		values = append(values, append([]byte{}, line...))
	}
	if scanner.Err() != nil {
		return nil, false, false
	}
	return values, true, true
}

// runJSONExtensionCommand runs an extension command which writes json to stdout, and
// formats the output with c8y's output handling (views, --select, --output etc.)
func runJSONExtensionCommand(f *cmdutil.Factory, cmd *cobra.Command, ext extensions.Extension, exe string, args []string) error {
	var stdout bytes.Buffer
	runErr := runExtension(f, ext, exe, args, true, &stdout)

	if err := writeJSONOutput(f, cmd, stdout.Bytes()); err != nil {
		return err
	}
	return runErr
}

func writeJSONOutput(f *cmdutil.Factory, cmd *cobra.Command, output []byte) error {
	values, isArray, ok := splitJSONValues(output)
	if !ok {
		if log, err := f.Logger(); err == nil {
			log.Warnf("Extension command output is not valid json, so it is written without formatting")
		}
		_, err := f.IOStreams.Out.Write(output)
		return err
	}

	cfg, err := f.Config()
	if err != nil {
		return err
	}
	commonOptions, err := cfg.GetOutputCommonOptions(cmd)
	if err != nil {
		return err
	}
	if isArray {
		// Each item is a result, so don't look for a nested data property
		commonOptions.DisableResultPropertyDetection()
	}

	for _, value := range values {
		if err := f.WriteOutput(value, cmdutil.OutputContext{}, &commonOptions); err != nil {
			return err
		}
	}
	return nil
}
