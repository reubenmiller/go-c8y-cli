package mcpserver

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

// Approval controls how write commands are approved before they are executed
type Approval string

const (
	// ApprovalAuto asks the user via MCP elicitation when the client supports it,
	// otherwise it relies on the MCP client's own tool permission prompt
	ApprovalAuto Approval = "auto"

	// ApprovalElicit always asks the user via MCP elicitation. If the client does
	// not support elicitation, then write commands are rejected
	ApprovalElicit Approval = "elicit"

	// ApprovalClient relies only on the MCP client's tool permission prompt
	// (the write tool is annotated as destructive)
	ApprovalClient Approval = "client"

	// ApprovalNone does not ask for any approval. Only use for trusted automation
	ApprovalNone Approval = "none"
)

func ParseApproval(v string) (Approval, error) {
	switch a := Approval(strings.ToLower(strings.TrimSpace(v))); a {
	case "":
		return ApprovalAuto, nil
	case ApprovalAuto, ApprovalElicit, ApprovalClient, ApprovalNone:
		return a, nil
	}
	return "", fmt.Errorf("invalid approval value %q. Accepted values: auto, elicit, client, none", v)
}

// DefaultDenyCommands commands which are not available to the agent by default.
// They either expose secrets, change the local configuration, or open local
// network services.
var DefaultDenyCommands = []string{
	"mcp",
	"sessions",
	"settings",
	"cli",
	"alias",
	"completion",
	"extensions",
	"microservices getBootstrapUser",
	"deviceregistration getCredentials",
	"remoteaccess connect run",
	"remoteaccess server",
}

// DeviceShellCommands commands which run a shell command on a device through a
// Cumulocity remote access (ssh) tunnel. The agent can only use them
// non-interactively, with the command to run on the device after "--", e.g.
// "remoteaccess connect ssh --device 1234 -- systemctl status". They are always
// treated as write commands, as the shell command can change anything on the device,
// and their dry run mode would still run the command.
var DeviceShellCommands = []string{
	"remoteaccess connect ssh",
}

// deviceShellDeniedFlags flags of the device shell commands which would open local
// network listeners
var deviceShellDeniedFlags = []string{"listen", "port-forward"}
var deviceShellDeniedShortFlags = []rune{'L'}

// DefaultReadCommands commands which never write to Cumulocity, in addition to
// any command whose name starts with a read verb (see readVerbPrefixes)
var DefaultReadCommands = []string{
	"version",
	"template",
	"util",
	"* subscribe",
	"* subscribeAll",
	"operations wait",
	"realtime subscribe",
	"realtime subscribeAll",
}

// readVerbPrefixes command names starting with one of these values are read-only
var readVerbPrefixes = []string{"get", "list", "find", "count", "exists", "show"}

// DeniedFlags global flags the agent is not allowed to set, as they would
// allow the agent to escape the policy (e.g. switching session or mode),
// write local files, or disable the audit trail
var DeniedFlags = []string{
	"session",
	"sessionUsername",
	"sessionPassword",
	"sessionMode",
	"host",
	"proxy",
	"noProxy",
	"insecure",
	"outputFile",
	"outputFileRaw",
	"noLog",
	"logMessage",
	"force",
	"confirm",
	"confirmText",
	"debug",
}

// DeniedShortFlags shorthand versions of DeniedFlags
var DeniedShortFlags = []rune{'U', 'P', 'k', 'f', 'l'}

// Policy controls what the agent can and can't do
type Policy struct {
	// Mode maximum session mode the agent may use. It can never exceed the
	// mode of the session itself. SessionModeUnset uses the session's mode
	Mode config.SessionMode

	// Allow if non-empty, only commands matching one of these patterns are allowed
	Allow []string

	// Deny commands matching any of these patterns are rejected (wins over Allow)
	Deny []string

	// ReadCommands additional commands which should be treated as read-only
	ReadCommands []string

	// Approval how write commands are approved
	Approval Approval

	// Timeout default timeout per tool call
	Timeout time.Duration

	// MaxTimeout upper limit for the timeout requested by the agent
	MaxTimeout time.Duration

	// MaxOutputBytes maximum number of bytes returned to the agent per call
	MaxOutputBytes int

	// DenyPaths additional local files and folders the commands can't read, e.g. with
	// --template, --data or --file. The session home folder and the active session
	// file are always denied
	DenyPaths []string

	// AllowDeviceShell allows commands which run a shell command on a device (see
	// DeviceShellCommands and c8y_Command operations). They still need a mode which
	// allows creating data, and are approved like other write commands
	AllowDeviceShell bool
}

func DefaultPolicy() Policy {
	return Policy{
		Mode:           config.SessionModeUnset,
		Deny:           append([]string{}, DefaultDenyCommands...),
		ReadCommands:   append([]string{}, DefaultReadCommands...),
		Approval:       ApprovalAuto,
		Timeout:        2 * time.Minute,
		MaxTimeout:     10 * time.Minute,
		MaxOutputBytes: 100_000,
	}
}

// ParseMode parses a mode name (prod, qual, dev). CI mode is not accepted
func ParseMode(name string) (config.SessionMode, error) {
	switch m := config.SessionModeUnset.FromString(name, false); m {
	case config.SessionModeProduction, config.SessionModeQual, config.SessionModeDev:
		return m, nil
	}
	return config.SessionModeUnset, fmt.Errorf("invalid mode %q. Accepted values: prod, qual, dev", name)
}

// CapMode returns the more restrictive of the policy mode and the session mode. An
// unset policy mode uses the session mode, and an unset session mode is prod (read
// only). The ci mode is treated like dev so the agent never gets more than dev.
func CapMode(policyMode, sessionMode config.SessionMode) config.SessionMode {
	if sessionMode == config.SessionModeUnset {
		sessionMode = config.SessionModeProduction
	}
	if sessionMode > config.SessionModeDev {
		sessionMode = config.SessionModeDev
	}
	if policyMode > config.SessionModeDev {
		policyMode = config.SessionModeDev
	}
	if policyMode < sessionMode {
		return policyMode
	}
	return sessionMode
}

// MatchCommand checks if a command path (e.g. "devices list") matches a pattern.
// Each word in the pattern is matched against the corresponding word of the
// command path and supports glob wildcards, e.g. "devices *", "* delete".
// A pattern also matches all sub commands, e.g. "devices" matches "devices list"
func MatchCommand(pattern, commandPath string) bool {
	patternWords := strings.Fields(strings.ToLower(pattern))
	pathWords := strings.Fields(strings.ToLower(commandPath))
	if len(patternWords) == 0 || len(patternWords) > len(pathWords) {
		return false
	}
	for i, p := range patternWords {
		if ok, err := path.Match(p, pathWords[i]); err != nil || !ok {
			return false
		}
	}
	return true
}

func matchAny(patterns []string, commandPath string) (string, bool) {
	for _, p := range patterns {
		if MatchCommand(p, commandPath) {
			return p, true
		}
	}
	return "", false
}

// CheckCommand checks if the command is allowed by the allow/deny lists
func (p Policy) CheckCommand(commandPath string) error {
	if pattern, ok := matchAny(p.Deny, commandPath); ok {
		return fmt.Errorf("command %q is denied by the MCP server policy (deny pattern: %q)", commandPath, pattern)
	}
	if len(p.Allow) > 0 {
		if _, ok := matchAny(p.Allow, commandPath); !ok {
			return fmt.Errorf("command %q is not in the MCP server allow list", commandPath)
		}
	}
	return nil
}

// IsReadCommand checks if the command only reads data
func (p Policy) IsReadCommand(commandPath string) bool {
	if _, ok := matchAny(p.ReadCommands, commandPath); ok {
		return true
	}
	parts := strings.Fields(commandPath)
	if len(parts) == 0 {
		return false
	}
	name := parts[len(parts)-1]
	for _, prefix := range readVerbPrefixes {
		if strings.HasPrefix(name, prefix) {
			// avoid matching names like "listen" style false positives by
			// requiring the next char (if any) to be upper case
			rest := name[len(prefix):]
			if rest == "" || (rest[0] >= 'A' && rest[0] <= 'Z') {
				return true
			}
		}
	}
	return false
}

// SplitAtTerminator splits the arguments at the argument terminator "--". The
// arguments after it are not interpreted as c8y flags.
func SplitAtTerminator(args []string) (before []string, after []string, found bool) {
	for i, arg := range args {
		if arg == "--" {
			return args[:i], args[i+1:], true
		}
	}
	return args, nil, false
}

// IsDeviceShellCommand returns true if the command runs a shell command on a device
func IsDeviceShellCommand(commandPath string) bool {
	_, ok := matchAny(DeviceShellCommands, commandPath)
	return ok
}

// CheckDeviceShellArgs checks that a device shell command is only used non-interactively
// and doesn't open local listeners
func CheckDeviceShellArgs(commandPath string, args []string) error {
	before, after, found := SplitAtTerminator(args)
	if !found || len(after) == 0 {
		return fmt.Errorf("%q can only run a single command on the device. Pass the command after \"--\", e.g. \"%s --device 1234 -- systemctl status\". Interactive sessions are not supported", commandPath, commandPath)
	}
	if err := checkDeniedFlags(before, deviceShellDeniedFlags, deviceShellDeniedShortFlags, "opens a local listener"); err != nil {
		return err
	}
	return nil
}

// CheckFlags rejects any flags the agent is not allowed to use. Arguments after
// the argument terminator "--" are not c8y flags, so they are not checked.
func CheckFlags(args []string) error {
	before, _, _ := SplitAtTerminator(args)
	return checkDeniedFlags(before, DeniedFlags, DeniedShortFlags, "is controlled by the MCP server")
}

func checkDeniedFlags(args []string, deniedFlags []string, deniedShortFlags []rune, reason string) error {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			name := strings.TrimPrefix(arg, "--")
			if i := strings.Index(name, "="); i >= 0 {
				name = name[:i]
			}
			for _, denied := range deniedFlags {
				if strings.EqualFold(name, denied) {
					return fmt.Errorf("flag --%s %s and can't be used by the agent", denied, reason)
				}
			}
			continue
		}
		if len(arg) >= 2 && arg[0] == '-' && isLetter(arg[1]) {
			// shorthand flag or group of shorthand flags, i.e. -k or -kf
			name := arg[1:]
			if i := strings.Index(name, "="); i >= 0 {
				name = name[:i]
			}
			for _, c := range name {
				for _, denied := range deniedShortFlags {
					if c == denied {
						return fmt.Errorf("flag -%c %s and can't be used by the agent", denied, reason)
					}
				}
			}
		}
	}
	return nil
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// HasDryRun checks if the arguments activate the dry run mode
func HasDryRun(args []string) bool {
	dry := false
	args, _, _ = SplitAtTerminator(args)
	for _, arg := range args {
		switch strings.ToLower(arg) {
		case "--dry", "--dry=true":
			dry = true
		case "--dry=false":
			dry = false
		}
	}
	return dry
}

// IsHelp checks if the arguments only request help
func IsHelp(args []string) bool {
	args, _, _ = SplitAtTerminator(args)
	for _, arg := range args {
		switch arg {
		case "--help", "-h", "--examples":
			return true
		}
	}
	return false
}
