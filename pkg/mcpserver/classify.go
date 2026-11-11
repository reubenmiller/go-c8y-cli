package mcpserver

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Segment a single c8y command within a pipeline
type Segment struct {
	// Args arguments passed to c8y (without the leading "c8y")
	Args []string

	// CommandPath resolved command path, e.g. "devices list"
	CommandPath string

	// Read true if the command only reads data
	Read bool

	// DeviceShell true if the command runs a shell command on a device. It has no
	// dry run mode, so no preview can be shown
	DeviceShell bool
}

// MarkDeviceCommands marks the segments which send a c8y_Command operation (a shell
// command run by the device's agent) as device shell commands. The data can come from
// the arguments or the input piped to the first command. This is a best effort check:
// a template can assemble the operation in other ways, so deny "operations create" to
// prevent device operations completely.
func (p *Plan) MarkDeviceCommands(stdin string) {
	for i, seg := range p.Segments {
		if seg.DeviceShell || seg.Read {
			continue
		}
		if seg.CommandPath != "operations create" && seg.CommandPath != "api" {
			continue
		}
		text := strings.Join(seg.Args, " ")
		if i == 0 {
			text += " " + stdin
		}
		if strings.Contains(strings.ToLower(text), "c8y_command") {
			p.Segments[i].DeviceShell = true
		}
	}
}

// HasDeviceShell returns true if any of the segments runs a shell command on a device
func (p *Plan) HasDeviceShell() bool {
	for _, s := range p.Segments {
		if s.DeviceShell {
			return true
		}
	}
	return false
}

// Plan the parsed and validated command(s) requested by the agent
type Plan struct {
	Segments []Segment
}

// ReadOnly returns true if all of the segments are read-only
func (p *Plan) ReadOnly() bool {
	for _, s := range p.Segments {
		if !s.Read {
			return false
		}
	}
	return true
}

// String returns the command line which will be executed (without injected flags)
func (p *Plan) String() string {
	parts := make([]string, 0, len(p.Segments))
	for _, s := range p.Segments {
		parts = append(parts, "c8y "+shellquote.Join(s.Args...))
	}
	return strings.Join(parts, " | ")
}

// SplitPipeline splits a command line into the individual c8y commands.
// The command is not passed to a shell, so only " | " (pipe surrounded by
// whitespace) between c8y commands is supported.
func SplitPipeline(commandLine string) ([][]string, error) {
	tokens, err := shellquote.Split(commandLine)
	if err != nil {
		return nil, fmt.Errorf("could not parse command. %w", err)
	}
	segments := [][]string{}
	current := []string{}
	for _, token := range tokens {
		if token == "|" {
			segments = append(segments, current)
			current = []string{}
			continue
		}
		current = append(current, token)
	}
	segments = append(segments, current)

	for i, seg := range segments {
		if len(seg) > 0 && (seg[0] == "c8y" || seg[0] == "c8y.exe") {
			seg = seg[1:]
			segments[i] = seg
		}
		if len(seg) == 0 {
			return nil, fmt.Errorf("empty command in pipeline (segment %d)", i+1)
		}
		if err := checkShellSyntax(seg); err != nil {
			return nil, err
		}
	}
	return segments, nil
}

func checkShellSyntax(args []string) error {
	for _, arg := range args {
		switch arg {
		case "&&", "||", ";", ">", ">>", "<", "&", "2>", "2>&1":
			return fmt.Errorf("shell operator %q is not supported. Only c8y commands joined with \" | \" are supported. Use --select, --filter or --output to shape the output instead", arg)
		}
	}
	return nil
}

// Classifier resolves commands against the c8y command tree and applies the policy
type Classifier struct {
	Root   *cobra.Command
	Policy Policy
}

// Plan parses and validates a command line
func (c *Classifier) Plan(commandLine string) (*Plan, error) {
	segments, err := SplitPipeline(commandLine)
	if err != nil {
		return nil, err
	}
	plan := &Plan{}
	for _, args := range segments {
		seg, err := c.classify(args)
		if err != nil {
			return nil, err
		}
		plan.Segments = append(plan.Segments, *seg)
	}
	return plan, nil
}

func (c *Classifier) classify(args []string) (*Segment, error) {
	if strings.HasPrefix(args[0], "-") {
		return nil, fmt.Errorf("command must start with a c8y command name, e.g. \"devices list\", got %q. Non-c8y commands (jq, grep etc.) are not supported", args[0])
	}

	cmd, _, err := c.Root.Find(args)
	if err != nil || cmd == nil || cmd == c.Root {
		return nil, fmt.Errorf("unknown c8y command %q. Use \"c8y <command> --help\" to discover commands", strings.Join(args, " "))
	}
	commandPath := strings.TrimSpace(strings.TrimPrefix(cmd.CommandPath(), c.Root.Name()))

	if err := c.Policy.CheckCommand(commandPath); err != nil {
		return nil, err
	}
	if err := CheckFlags(args); err != nil {
		return nil, err
	}

	seg := &Segment{
		Args:        args,
		CommandPath: commandPath,
	}

	_, _, hasTerminator := SplitAtTerminator(args)
	deviceShell := IsDeviceShellCommand(commandPath)
	if hasTerminator && !deviceShell {
		return nil, fmt.Errorf("the argument terminator \"--\" is only supported by %s", strings.Join(DeviceShellCommands, ", "))
	}

	switch {
	case IsHelp(args):
		seg.Read = true
	case deviceShell:
		// Always a write: the shell command can change anything on the device, and
		// the command doesn't support the dry run mode
		if err := CheckDeviceShellArgs(commandPath, args); err != nil {
			return nil, err
		}
		seg.DeviceShell = true
		seg.Read = false
	case HasDryRun(args):
		// dry run mode does not send any write requests
		seg.Read = true
	case commandPath == "api":
		read, err := classifyAPI(cmd, args)
		if err != nil {
			return nil, err
		}
		seg.Read = read
	default:
		seg.Read = c.Policy.IsReadCommand(commandPath)
	}
	return seg, nil
}

// classifyAPI checks the method used by the generic "c8y api" command
func classifyAPI(cmd *cobra.Command, args []string) (bool, error) {
	method := http.MethodGet
	positional := []string{}

	flagSet := pflag.NewFlagSet("api", pflag.ContinueOnError)
	flagSet.AddFlagSet(cmd.Flags())
	flagSet.AddFlagSet(cmd.InheritedFlags())

	// skip the command name
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			name, value, hasValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
			var flag *pflag.Flag
			if strings.HasPrefix(arg, "--") {
				flag = flagSet.Lookup(name)
			} else if len(name) == 1 {
				flag = flagSet.ShorthandLookup(name)
			}
			if flag == nil || flag.Value.Type() == "bool" {
				continue
			}
			if !hasValue && i+1 < len(args) {
				i++
				value = args[i]
			}
			switch flag.Name {
			case "method":
				method = value
			case "url":
				if strings.Contains(value, "://") {
					return false, fmt.Errorf("absolute urls are not allowed. Use a path relative to the session's host, e.g. /inventory/managedObjects")
				}
			}
			continue
		}
		positional = append(positional, arg)
	}

	for i, v := range positional {
		if strings.Contains(v, "://") {
			return false, fmt.Errorf("absolute urls are not allowed. Use a path relative to the session's host, e.g. /inventory/managedObjects")
		}
		if i == 0 && len(positional) > 1 {
			method = v
		} else if i == 0 && isHTTPMethod(v) {
			method = v
		}
	}

	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true, nil
	}
	return false, nil
}

func isHTTPMethod(v string) bool {
	switch strings.ToUpper(v) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		return true
	}
	return false
}
