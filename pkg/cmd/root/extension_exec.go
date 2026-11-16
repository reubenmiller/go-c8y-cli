package root

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions/cmdtree"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// ExtensionContextVersion is the version of the runtime context (environment variables)
	// which is passed to extension commands. Increase it when variables are added or changed
	ExtensionContextVersion = "1"

	// Environment variables passed to extension commands
	EnvExtensionName           = "C8Y_EXTENSION_NAME"
	EnvExtensionDir            = "C8Y_EXTENSION_DIR"
	EnvExtensionContextVersion = "C8Y_EXTENSION_CONTEXT_VERSION"
	EnvBin                     = "C8Y_BIN"
	EnvSettingsDryRun          = "C8Y_SETTINGS_DEFAULTS_DRY"
	EnvSettingsMode            = "C8Y_SETTINGS_MODE"
	EnvSettingsInsecure        = "C8Y_SETTINGS_DEFAULTS_INSECURE"

	// annotationExtension marks the root command of an extension (value is the extension name)
	annotationExtension = "c8y.extension"

	// manifestOrigin is the origin used in log messages for commands declared in the manifest
	manifestOrigin = "extension.yaml"
)

// IsExtensionCommand returns true if the command is provided by an extension
func IsExtensionCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations[annotationExtension] != "" {
			return true
		}
	}
	return false
}

// countLeadingFlags returns the number of arguments at the start of args which are flags
// (and their values), using the same rules as cobra's Command.Traverse
func countLeadingFlags(flags *pflag.FlagSet, args []string) int {
	inFlag := false
	for i, arg := range args {
		switch {
		case arg == "--":
			return i
		case strings.HasPrefix(arg, "--") && !strings.Contains(arg, "="):
			flag := flags.Lookup(arg[2:])
			inFlag = flag == nil || flag.NoOptDefVal == ""
		case strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=") && len(arg) == 2:
			flag := flags.ShorthandLookup(arg[1:])
			inFlag = flag == nil || flag.NoOptDefVal == ""
		case inFlag:
			inFlag = false
		case strings.HasPrefix(arg, "-") && len(arg) > 1:
			// flag with an = separated value
		default:
			return i
		}
	}
	return len(args)
}

// StripLeadingGlobalFlags removes the c8y flags which are placed before an extension
// command, e.g. "c8y --session prod.json myext list" (they have already been parsed by c8y
// and are passed to the extension via the runtime context). The arguments are returned
// unchanged for any other command.
func StripLeadingGlobalFlags(rootCmd *cobra.Command, args []string) []string {
	cmd, _, err := rootCmd.Traverse(args)
	if err != nil || cmd == rootCmd || !IsExtensionCommand(cmd) {
		return args
	}
	n := countLeadingFlags(rootCmd.PersistentFlags(), args)
	if n == 0 || n >= len(args) {
		return args
	}
	return args[n:]
}

// extensionEnv builds the environment for an extension command: the current environment,
// the variables of the active session (without the password), and the extension variables
func extensionEnv(f *cmdutil.Factory, ext extensions.Extension) []string {
	set := map[string]string{}
	unset := map[string]bool{}

	if cfg, err := f.Config(); err == nil {
		if client, err := f.Client(); err == nil && client != nil && client.BaseURL != nil && client.BaseURL.Host != "" {
			session := &c8ysession.CumulocitySession{
				Path:     cfg.GetSessionFile(),
				Host:     strings.TrimSuffix(c8y.FormatBaseURL(client.BaseURL.String()), "/"),
				Token:    client.Token,
				Tenant:   cfg.GetTenant(),
				Version:  cfg.GetCumulocityVersion(),
				Username: client.Username,
			}
			for key, value := range c8ysession.GetVariablesFromSession(session, cfg, client, false) {
				// Remove values from the parent environment which belong to another session
				unset[key] = true
				if v := fmt.Sprintf("%v", value); v != "" {
					set[key] = v
				}
			}
			// The password is never passed on. The extension can use the session file,
			// token or authorization header instead. A value set by the user is kept
			delete(set, "C8Y_PASSWORD")
			delete(unset, "C8Y_PASSWORD")
		}

		mode := cfg.SessionMode().String()
		set[config.EnvSessionMode] = mode
		set[EnvSettingsMode] = mode
		if cfg.DryRun() {
			set[EnvSettingsDryRun] = "true"
		}
		if cfg.SkipSSLVerify() {
			set[EnvSettingsInsecure] = "true"
		}
	}

	set[EnvExtensionName] = ext.Name()
	set[EnvExtensionDir] = ext.Path()
	set[EnvExtensionContextVersion] = ExtensionContextVersion
	if exe, err := os.Executable(); err == nil {
		set[EnvBin] = exe
	} else if f.Executable != "" {
		set[EnvBin] = f.Executable
	}

	return mergeEnv(os.Environ(), set, unset)
}

// mergeEnv returns base without the variables in set and unset, followed by the variables in set
func mergeEnv(base []string, set map[string]string, unset map[string]bool) []string {
	sameKey := func(a, b string) bool {
		if runtime.GOOS == "windows" {
			return strings.EqualFold(a, b)
		}
		return a == b
	}
	isReplaced := func(name string) bool {
		for key := range set {
			if sameKey(key, name) {
				return true
			}
		}
		for key := range unset {
			if sameKey(key, name) {
				return true
			}
		}
		return false
	}

	env := make([]string, 0, len(base)+len(set))
	for _, item := range base {
		name, _, _ := strings.Cut(item, "=")
		if isReplaced(name) {
			continue
		}
		env = append(env, item)
	}
	for key, value := range set {
		env = append(env, key+"="+value)
	}
	return env
}

// extensionExitError converts the error of an extension process. The extension has already
// reported the error, so only the exit code is used
func extensionExitError(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code := exitErr.ExitCode()
		if code < 0 {
			code = int(cmderrors.ExitError)
		}
		return cmderrors.NewSilentErrorWithExitCode(cmderrors.ExitCode(code), err)
	}
	return err
}

// runExtension runs an extension executable with the extension runtime context
func runExtension(f *cmdutil.Factory, ext extensions.Extension, exe string, args []string, isBinary bool, stdout io.Writer) error {
	// Get the logger at runtime, as it is reconfigured after the commands are created
	if log, err := f.Logger(); err == nil {
		log.Infof("Executing extension. name: %s, command: %s, args: %v", ext.Name(), exe, args)
	}
	err := f.ExtensionManager().Execute(extensions.ExecOptions{
		Executable: exe,
		Args:       args,
		IsBinary:   isBinary,
		Env:        extensionEnv(f, ext),
		Stdin:      f.IOStreams.In,
		Stdout:     stdout,
		Stderr:     f.IOStreams.ErrOut,
	})
	return extensionExitError(err)
}

// completeWithBinary returns a completion function which asks the binary for completions
// (using cobra's __complete protocol). prefixArgs are the arguments which select the
// binary's command, e.g. the exec args of a declared command
func completeWithBinary(f *cmdutil.Factory, ext extensions.Extension, exe string, prefixArgs []string) cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		completeArgs := append([]string{cobra.ShellCompRequestCmd}, prefixArgs...)
		completeArgs = append(completeArgs, args...)
		completeArgs = append(completeArgs, toComplete)

		var stdout bytes.Buffer
		err := f.ExtensionManager().Execute(extensions.ExecOptions{
			Executable: exe,
			Args:       completeArgs,
			IsBinary:   true,
			Env:        extensionEnv(f, ext),
			Stdout:     &stdout,
			Stderr:     io.Discard,
		})
		if err != nil {
			cobra.CompDebugln(fmt.Sprintf("extension completion failed. %s", err), true)
			return nil, cobra.ShellCompDirectiveError
		}
		return parseCobraCompletion(&stdout)
	}
}

// parseCobraCompletion parses the output of a cobra __complete command
func parseCobraCompletion(r io.Reader) ([]cobra.Completion, cobra.ShellCompDirective) {
	lines := []string{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	directive := cobra.ShellCompDirectiveDefault
	if len(lines) > 0 {
		last := lines[len(lines)-1]
		if strings.HasPrefix(last, ":") {
			if v, err := strconv.Atoi(last[1:]); err == nil {
				directive = cobra.ShellCompDirective(v)
			}
			lines = lines[:len(lines)-1]
		}
	}
	completions := make([]cobra.Completion, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			completions = append(completions, line)
		}
	}
	return completions, directive
}

// addExtensionBinaryCommands adds the binary commands declared in the extension manifest
// (extension.yaml), and mounts the binary
func addExtensionBinaryCommands(f *cmdutil.Factory, log *logger.Logger, tree *cmdtree.Tree, ext extensions.Extension) error {
	manifest, err := ext.Manifest()
	if err != nil {
		return err
	}
	if manifest.Binary == nil {
		if len(manifest.Commands) > 0 {
			log.Warnf("Ignoring binary commands as the extension does not declare a binary. extension=%s, file=%s", ext.Name(), manifestOrigin)
		}
		return nil
	}

	exe := ext.Executable()
	if _, err := os.Stat(exe); err != nil {
		log.Warnf("Extension binary not found, so its commands are ignored. extension=%s, path=%s", ext.Name(), exe)
		return nil
	}

	binary := manifest.Binary
	for _, item := range manifest.Commands {
		path := item.Path()
		if len(path) == 0 {
			continue
		}
		name := path[len(path)-1]
		tree.AddLeaf(path, cmdtree.Leaf{
			Source: cmdtree.SourceBinary,
			Origin: manifestOrigin,
			Build: func() (*cobra.Command, error) {
				short := item.Description
				if short == "" {
					short = fmt.Sprintf("Run %s command", name)
				}
				cmd := &cobra.Command{
					Use:   name,
					Short: short,
					Args:  cobra.ArbitraryArgs,
					// Pass all arguments (including flags) to the binary
					DisableFlagParsing: true,
					Annotations:        map[string]string{},
					RunE: func(cmd *cobra.Command, args []string) error {
						binArgs := append(slices.Clone(item.ExecArgs()), args...)
						if item.IsJSON() {
							// c8y output flags have already been removed from the args (see PrepareExtensionArgs)
							return runJSONExtensionCommand(f, cmd, ext, exe, binArgs)
						}
						return runExtension(f, ext, exe, binArgs, true, f.IOStreams.Out)
					},
				}
				if item.IsJSON() {
					cmd.Annotations[annotationOutput] = extensions.OutputJSON
				}
				if binary.CompletionEnabled() {
					cmd.ValidArgsFunction = completeWithBinary(f, ext, exe, item.ExecArgs())
				}
				cmdutil.DisableAuthCheckForCommand(cmd)
				return cmd, nil
			},
		})
	}

	tree.AddMount(binary.MountPath(), cmdtree.Mount{
		Origin: manifestOrigin,
		Short:  fmt.Sprintf("Run %s extension binary", ext.Name()),
		Long:   "Arguments which do not match a command are passed to the extension binary.",
		Configure: func(cmd *cobra.Command) error {
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				return runExtension(f, ext, exe, args, true, f.IOStreams.Out)
			}
			if binary.CompletionEnabled() {
				cmd.ValidArgsFunction = completeWithBinary(f, ext, exe, nil)
			}
			cmdutil.DisableAuthCheckForCommand(cmd)
			return nil
		},
	})
	return nil
}
