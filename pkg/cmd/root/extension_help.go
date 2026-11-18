package root

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/spf13/cobra"
)

// extensionHelpRequest is a request for the help or the examples of a declared binary command.
// The binary is a separate program which does not know the c8y --examples flag or the help
// levels (i.e. --help local), so c8y handles them, and only asks the binary for its plain help
type extensionHelpRequest struct {
	help     bool
	level    string
	examples bool

	// args are the other arguments (without the help and examples flags)
	args []string
}

func (r extensionHelpRequest) requested() bool {
	return r.help || r.examples
}

// parseExtensionHelpArgs finds the help (-h, --help, --help=<level>, --help <level>) and
// examples (--examples) flags in the arguments of a declared binary command. Arguments after
// "--" are not checked
func parseExtensionHelpArgs(args []string) (extensionHelpRequest, error) {
	req := extensionHelpRequest{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			req.args = append(req.args, args[i:]...)
			return req, nil
		case arg == "-h" || arg == "--help":
			req.help = true
			if i+1 < len(args) && isHelpLevel(args[i+1]) {
				req.level = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--help=") || strings.HasPrefix(arg, "-h="):
			_, value, _ := strings.Cut(arg, "=")
			if v, err := strconv.ParseBool(value); err == nil {
				req.help = v
				req.level = ""
				continue
			}
			if !isHelpLevel(value) {
				return req, cmderrors.NewUserError(fmt.Sprintf("invalid help level %q. Valid values: %s", value, strings.Join(HelpLevelNames(), ", ")))
			}
			req.help = true
			req.level = value
		case arg == "--examples":
			req.examples = true
		case strings.HasPrefix(arg, "--examples="):
			v, err := strconv.ParseBool(strings.TrimPrefix(arg, "--examples="))
			if err != nil {
				return req, cmderrors.NewUserError(fmt.Sprintf("invalid argument %q for \"--examples\" flag: %s", arg, err))
			}
			req.examples = v
		default:
			req.args = append(req.args, arg)
		}
	}
	return req, nil
}

// runDeclaredCommandHelp shows the help or the examples of a declared binary command.
// binArgs are the arguments which select the binary's command (without the help and examples flags)
func runDeclaredCommandHelp(f *cmdutil.Factory, cmd *cobra.Command, ext extensions.Extension, exe string, binArgs []string, req extensionHelpRequest) error {
	out := f.IOStreams.Out

	sel := helpSelection{}
	if req.help {
		level := req.level
		if level == "" && f != nil {
			if cfg, err := f.Config(); err == nil && cfg != nil {
				level = cfg.GetDefaultHelpLevel()
			}
		}
		sel = parseHelpLevel(level)
		if sel.listCategories && !sel.local && !sel.allGlobal && len(sel.categories) == 0 {
			writeHelpCategories(out, cmd)
			return nil
		}
	}

	// Only ask the binary for its help when it is needed
	var binaryHelp []byte
	var runErr error
	if req.examples || sel.local {
		var stdout bytes.Buffer
		runErr = runExtension(f, ext, exe, append(slices.Clone(binArgs), "--help"), true, &stdout)
		binaryHelp = stdout.Bytes()
	}

	if !req.help {
		if examples := extractHelpSection(binaryHelp, "Examples"); examples != "" {
			fmt.Fprintln(out, examples)
		} else if log, err := f.Logger(); err == nil {
			log.Infof("The extension command does not provide any examples")
		}
		return runErr
	}

	writeDeclaredCommandUsage(out, cmd, binaryHelp, sel)
	return runErr
}

// writeDeclaredCommandUsage writes the help of a declared binary command: the help of the binary
// (which includes the flags of the command), followed by the c8y global flags selected by the help level
func writeDeclaredCommandUsage(w io.Writer, cmd *cobra.Command, binaryHelp []byte, sel helpSelection) {
	if sel.local {
		fmt.Fprint(w, strings.TrimRight(string(binaryHelp), " \t\r\n"))
	} else {
		fmt.Fprintf(w, "Usage:\n  %s", cmd.UseLine())
	}

	hiddenGlobal := writeGlobalFlags(w, cmd.InheritedFlags(), sel)

	if sel.listCategories {
		fmt.Fprint(w, "\n")
		writeHelpCategories(w, cmd)
	}
	if hiddenGlobal || !sel.local {
		fmt.Fprintf(w, "\n\nUse \"%s --help all\" to show all flags, or \"%s --help category\" to list the flag categories.", cmd.CommandPath(), cmd.CommandPath())
	}
	fmt.Fprintln(w)
}

// extractHelpSection returns the content of a section (i.e. "Examples") of a cobra help text.
// The section ends at the next heading, which is a line without indentation ending with ":"
func extractHelpSection(help []byte, name string) string {
	heading := name + ":"
	lines := []string{}
	inSection := false
	scanner := bufio.NewScanner(bytes.NewReader(help))
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t\r")
		if !inSection {
			inSection = line == heading
			continue
		}
		if line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && strings.HasSuffix(line, ":") {
			break
		}
		lines = append(lines, line)
	}
	return strings.Trim(strings.Join(lines, "\n"), "\n")
}
