package root

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// annotationOutput marks extension commands whose output is processed by c8y (value "json")
const annotationOutput = "c8y.extension.output"

// annotationDeclared marks binary commands which are declared in the extension manifest.
// c8y handles the global flags of these commands, so they are not passed to the binary
const annotationDeclared = "c8y.extension.declared"

// passthroughGlobalFlags are global flags which are still passed to declared binary commands,
// so the binary can show its own help and examples
var passthroughGlobalFlags = map[string]bool{
	"help":     true,
	"examples": true,
}

// lookupGlobalFlag returns the c8y global flag matching the argument, and whether the argument
// includes the value (--name=value, -n=value or -nvalue). Only single shorthand flags are matched
// (e.g. -o csv), not combined shorthands (e.g. -vc)
func lookupGlobalFlag(flags *pflag.FlagSet, arg string) (*pflag.Flag, bool) {
	var flag *pflag.Flag
	inline := false
	switch {
	case strings.HasPrefix(arg, "--") && len(arg) > 2:
		name, _, hasValue := strings.Cut(arg[2:], "=")
		flag = flags.Lookup(name)
		inline = hasValue
	case strings.HasPrefix(arg, "-") && len(arg) >= 2:
		flag = flags.ShorthandLookup(arg[1:2])
		if flag != nil && len(arg) > 2 {
			// Bool flags only accept -v=true, others also accept -ovalue
			if arg[2] != '=' && flag.NoOptDefVal != "" {
				return nil, false
			}
			inline = true
		}
	}
	if flag == nil || passthroughGlobalFlags[flag.Name] {
		return nil, false
	}
	return flag, inline
}

// takesValue returns true if the flag needs a value as a separate argument
func takesValue(flag *pflag.Flag, inline bool) bool {
	return !inline && flag.NoOptDefVal == ""
}

// extractGlobalFlags splits args into the c8y global flags (with their values) and the
// remaining arguments. Arguments after "--" are not checked.
func extractGlobalFlags(flags *pflag.FlagSet, args []string) (globalFlags []string, rest []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			rest = append(rest, args[i:]...)
			break
		}
		flag, inline := lookupGlobalFlag(flags, arg)
		if flag == nil {
			rest = append(rest, arg)
			continue
		}
		globalFlags = append(globalFlags, arg)
		if takesValue(flag, inline) && i+1 < len(args) {
			globalFlags = append(globalFlags, args[i+1])
			i++
		}
	}
	return globalFlags, rest
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
//   - for binary commands declared in the extension manifest, the c8y global flags after the
//     command (e.g. myext list --session prod.json --output csv) are also parsed by c8y and
//     removed, the same as for built-in commands. Other arguments are passed to the binary.
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
	if target.Annotations[annotationDeclared] != "" {
		var globalFlags []string
		globalFlags, rest = extractGlobalFlags(rootCmd.PersistentFlags(), rest)
		c8yFlags = append(c8yFlags, globalFlags...)
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

// completeGlobalFlagValue completes the value of a c8y global flag (e.g. "--session <TAB>" or
// "--output=<TAB>") using the completion registered on the root command. ok is false if the
// argument being completed is not the value of a global flag
func completeGlobalFlagValue(cmd *cobra.Command, args []string, toComplete string) (values []cobra.Completion, directive cobra.ShellCompDirective, ok bool) {
	flags := cmd.Root().PersistentFlags()
	var flag *pflag.Flag
	prefix := ""
	value := toComplete

	if strings.HasPrefix(toComplete, "-") && strings.Contains(toComplete, "=") {
		// --name=<value>
		name, v, _ := strings.Cut(toComplete, "=")
		if f, _ := lookupGlobalFlag(flags, name); f != nil {
			flag = f
			prefix = name + "="
			value = v
		}
	} else if len(args) > 0 {
		// --name <value>
		if f, inline := lookupGlobalFlag(flags, args[len(args)-1]); f != nil && takesValue(f, inline) {
			flag = f
		}
	}
	if flag == nil {
		return nil, cobra.ShellCompDirectiveDefault, false
	}

	completeFunc, exists := cmd.Root().GetFlagCompletionFunc(flag.Name)
	if !exists {
		return nil, cobra.ShellCompDirectiveNoFileComp, true
	}
	values, directive = completeFunc(cmd, args, value)
	if prefix != "" {
		for i := range values {
			values[i] = prefix + values[i]
		}
	}
	return values, directive, true
}

// declaredCommandCompletion returns the completion of a declared binary command. The values of
// c8y global flags are completed by c8y, all other arguments by binaryComplete (if not nil),
// without the c8y global flags
func declaredCommandCompletion(binaryComplete cobra.CompletionFunc) cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if values, directive, ok := completeGlobalFlagValue(cmd, args, toComplete); ok {
			return values, directive
		}
		if binaryComplete == nil {
			return nil, cobra.ShellCompDirectiveDefault
		}
		_, binaryArgs := extractGlobalFlags(cmd.Root().PersistentFlags(), args)
		return binaryComplete(cmd, binaryArgs, toComplete)
	}
}

// hasHelpFlag returns true if the arguments request help (-h or --help, before any "--")
func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "--":
			return false
		}
		if isHelpArg(arg) && arg != "--help=false" {
			return true
		}
	}
	return false
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
		// Not an error, e.g. a command can print a message instead of json
		if log, err := f.Logger(); err == nil {
			log.Infof("Extension command output is not json, so it is written without formatting")
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
