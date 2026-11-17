package extension

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdparser"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions/cmdtree"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// Finding severities
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityInfo    = "info"
)

// Validation checks
const (
	CheckManifest = "manifest"
	CheckBinary   = "binary"
	CheckFlags    = "flags"
	CheckCommands = "commands"
	CheckAPI      = "api"
	CheckScripts  = "scripts"
	CheckName     = "name"
)

// Finding is a problem found when validating an extension
type Finding struct {
	Severity string `json:"severity"`
	Check    string `json:"check"`
	Location string `json:"location"`
	Message  string `json:"message"`
}

// ValidateOptions controls the validation of an extension
type ValidateOptions struct {
	// GlobalFlags are the c8y global flags (persistent flags of the root command). The flag check
	// is skipped if nil
	GlobalFlags *pflag.FlagSet

	// SkipFlagCheck skips the flag check, which runs the extension's binary
	SkipFlagCheck bool

	// GOOS is the operating system to validate for. Defaults to runtime.GOOS
	GOOS string

	// CompleteTimeout is the maximum duration to wait for the binary to list its flags
	CompleteTimeout time.Duration

	// RootCommand is the c8y root command, used to check that the name is not a built-in
	// command. The check is skipped if nil
	RootCommand *cobra.Command

	// Installed is true if the extension is installed (and not a folder being developed), so its
	// installed name can be compared with the declared name
	Installed bool
}

// globalFlagsHandledByBinary are global flags which are still passed to declared binary commands
var globalFlagsHandledByBinary = map[string]bool{
	"help":     true,
	"examples": true,
}

var unknownFieldPattern = regexp.MustCompile(`line (\d+): field (\S+) not found in type (\S+)`)

// manifestSections maps the manifest types to the key path shown to the user
var manifestSections = map[string]string{
	"extension.ExtensionFile":    "",
	"extensions.Manifest":        "",
	"extensions.BinarySpec":      "binary.",
	"extensions.BinaryCommand":   "commands[].",
	"extension.AliasExtension":   "aliases[].",
	"extensions.SessionProvider": "sessionProviders[].",
}

type validator struct {
	ext      extensions.Extension
	opts     ValidateOptions
	findings []Finding
}

func (v *validator) add(severity, check, location, format string, args ...any) {
	v.findings = append(v.findings, Finding{
		Severity: severity,
		Check:    check,
		Location: location,
		Message:  fmt.Sprintf(format, args...),
	})
}

func (v *validator) relative(path string) string {
	if rel, err := filepath.Rel(v.ext.Path(), path); err == nil {
		return filepath.ToSlash(rel)
	}
	return path
}

func (v *validator) commandPath(path ...string) string {
	return strings.Join(append([]string{"c8y", v.ext.Name()}, path...), " ")
}

// ValidateExtension checks an extension for problems and returns the findings, sorted by severity
func ValidateExtension(ext extensions.Extension, opts ValidateOptions) []Finding {
	if opts.GOOS == "" {
		opts.GOOS = runtime.GOOS
	}
	if opts.CompleteTimeout == 0 {
		opts.CompleteTimeout = 5 * time.Second
	}
	v := &validator{ext: ext, opts: opts}

	manifest := v.checkManifest()
	v.checkName(manifest)
	commands := v.checkDeclaredCommands(manifest)
	v.checkSessionProviders(manifest)
	exe := v.checkBinary(manifest)
	v.checkCommandTree(manifest, commands)
	if exe != "" && !opts.SkipFlagCheck && opts.GlobalFlags != nil {
		v.checkFlags(manifest, commands, exe)
	}

	order := map[string]int{SeverityError: 0, SeverityWarning: 1, SeverityInfo: 2}
	sort.SliceStable(v.findings, func(i, j int) bool {
		return order[v.findings[i].Severity] < order[v.findings[j].Severity]
	})
	return v.findings
}

// checkManifest parses extension.yaml, and reports invalid yaml and unknown keys
func (v *validator) checkManifest() *extensions.Manifest {
	path := filepath.Join(v.ext.Path(), fileAlias)
	contents, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			v.add(SeverityError, CheckManifest, fileAlias, "could not read the file. %s", err)
		}
		return &extensions.Manifest{}
	}

	file := &ExtensionFile{}
	if err := yaml.Unmarshal(contents, file); err != nil {
		v.add(SeverityError, CheckManifest, fileAlias, "invalid yaml, so the manifest is ignored. %s", err)
		return &extensions.Manifest{}
	}

	strict := yaml.NewDecoder(bytes.NewReader(contents))
	strict.KnownFields(true)
	if err := strict.Decode(&ExtensionFile{}); err != nil {
		for _, match := range unknownFieldPattern.FindAllStringSubmatch(err.Error(), -1) {
			key := manifestSections[match[3]] + match[2]
			v.add(SeverityWarning, CheckManifest, fileAlias+":"+match[1], "unknown key '%s' is ignored. Check the spelling", key)
		}
	}

	if file.Binary != nil {
		switch strings.ToLower(file.Binary.Completion) {
		case "", extensions.CompletionNone, extensions.CompletionCobra:
		default:
			v.add(SeverityError, CheckManifest, fileAlias, "invalid binary.completion '%s'. Use '%s' or '%s'", file.Binary.Completion, extensions.CompletionCobra, extensions.CompletionNone)
		}
	}
	return &file.Manifest
}

// checkName checks the name declared in the manifest
func (v *validator) checkName(manifest *extensions.Manifest) {
	if manifest.Name == "" {
		return
	}
	name, err := extensions.NormalizeName(manifest.Name)
	if err != nil {
		v.add(SeverityError, CheckName, fileAlias, "invalid name '%s', so the extension can't be installed. Use lowercase letters, numbers, and '-' (%s)", manifest.Name, extensions.NameFormat)
		return
	}
	if manifest.Name != name {
		v.add(SeverityWarning, CheckName, fileAlias, "the extension is installed as '%s'. Use 'name: %s'", name, name)
	}
	if root := v.opts.RootCommand; root != nil {
		if isBuiltInCommand(root, name) {
			v.add(SeverityError, CheckName, fileAlias, "name '%s' is a built-in command, so the extension can't be installed under it", name)
		}
	}
	if v.opts.Installed && v.ext.Name() != name {
		v.add(SeverityInfo, CheckName, fileAlias, "the extension is installed as '%s', but declares the name '%s'", v.ext.Name(), name)
	}
}

// checkDeclaredCommands checks the declared binary commands, and returns the valid ones
func (v *validator) checkDeclaredCommands(manifest *extensions.Manifest) []extensions.BinaryCommand {
	if len(manifest.Commands) == 0 {
		return nil
	}
	if manifest.Binary == nil {
		v.add(SeverityError, CheckManifest, fileAlias, "%d declared command(s) are ignored, as the manifest does not declare a binary", len(manifest.Commands))
		return nil
	}

	valid := []extensions.BinaryCommand{}
	seen := map[string]bool{}
	for i, item := range manifest.Commands {
		location := fmt.Sprintf("%s commands[%d]", fileAlias, i)
		path := strings.Join(item.Path(), " ")
		switch {
		case path == "":
			v.add(SeverityError, CheckManifest, location, "the command name is empty")
			continue
		case seen[path]:
			v.add(SeverityError, CheckManifest, location, "duplicate command '%s'", path)
			continue
		}
		seen[path] = true
		switch strings.ToLower(item.Output) {
		case "", extensions.OutputRaw, extensions.OutputJSON:
		default:
			v.add(SeverityError, CheckManifest, location, "invalid output '%s' for command '%s'. Use '%s' or '%s'", item.Output, path, extensions.OutputJSON, extensions.OutputRaw)
		}
		valid = append(valid, item)
	}
	return valid
}

// checkSessionProviders reports session providers which are ignored
func (v *validator) checkSessionProviders(manifest *extensions.Manifest) {
	_, problems := manifest.ValidSessionProviders()
	for _, problem := range problems {
		v.add(SeverityError, CheckManifest, fileAlias+" sessionProviders", "%s", problem)
	}
}

// checkBinary checks that the declared binary exists and is executable, and returns its path
func (v *validator) checkBinary(manifest *extensions.Manifest) string {
	if manifest.Binary == nil {
		return ""
	}
	exe := v.ext.Executable()
	info, err := os.Stat(exe)
	switch {
	case err != nil:
		v.add(SeverityError, CheckBinary, v.relative(exe), "the binary does not exist, so the binary commands are not available. Build it, or check binary.path")
		return ""
	case !info.Mode().IsRegular():
		v.add(SeverityError, CheckBinary, v.relative(exe), "the binary is not a file")
		return ""
	case v.opts.GOOS != "windows" && info.Mode().Perm()&0111 == 0:
		v.add(SeverityError, CheckBinary, v.relative(exe), "the binary is not executable. Run: chmod +x %s", v.relative(exe))
		return ""
	}
	return exe
}

// checkCommandTree reports commands which are ignored because of conflicts between command sources.
// It uses the same command paths as the command registration (ConvertToCobraCommands)
func (v *validator) checkCommandTree(manifest *extensions.Manifest, commands []extensions.BinaryCommand) {
	tree := cmdtree.New(v.ext.Name(), "")
	tree.OnConflict = func(c cmdtree.Conflict) {
		v.add(SeverityWarning, CheckCommands, "c8y "+c.Path, "%s is ignored, as %s defines the same command (precedence: api, binary, script)", c.Dropped, c.Kept)
	}

	// 1. API specifications
	apiDir := filepath.Join(v.ext.Path(), apiName)
	_ = filepath.WalkDir(apiDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if ext := filepath.Ext(path); ext != ".yaml" && ext != ".yml" {
			return nil
		}
		origin := v.relative(path)
		file, err := os.Open(path)
		if err != nil {
			v.add(SeverityError, CheckAPI, origin, "could not read the file. %s", err)
			return nil
		}
		spec, err := cmdparser.ParseSpec(file)
		file.Close()
		if err != nil {
			v.add(SeverityError, CheckAPI, origin, "invalid api specification, so its commands are not available. %s", err)
			return nil
		}
		if spec.Group.Skip {
			return nil
		}
		groupPath := strings.Fields(spec.Group.Name)
		tree.AddGroup(groupPath, spec.Group.Description, spec.Group.DescriptionLong, origin)
		for _, item := range spec.Commands {
			if item.ShouldIgnore() {
				continue
			}
			tree.AddLeaf(append(slices.Clone(groupPath), item.Name), cmdtree.Leaf{Source: cmdtree.SourceAPI, Origin: origin})
		}
		return nil
	})

	// 2. Binary
	if manifest.Binary != nil {
		for _, item := range commands {
			tree.AddLeaf(item.Path(), cmdtree.Leaf{Source: cmdtree.SourceBinary, Origin: fileAlias})
		}
		tree.AddMount(manifest.Binary.MountPath(), cmdtree.Mount{Origin: fileAlias})
	}

	// 3. Scripts
	scripts, err := v.ext.Commands()
	if err != nil {
		v.add(SeverityError, CheckScripts, commandsName, "could not read the commands folder. %s", err)
	}
	for _, script := range scripts {
		origin := v.relative(script.Command())
		tree.AddLeaf(strings.Split(script.Name(), " "), cmdtree.Leaf{Source: cmdtree.SourceScript, Origin: origin})
		if v.opts.GOOS != "windows" {
			if info, err := os.Stat(script.Command()); err == nil && info.Mode().Perm()&0111 == 0 {
				v.add(SeverityWarning, CheckScripts, origin, "the script is not executable. Run: chmod +x %s", origin)
			}
		}
	}
}

// checkFlags reports flags of declared commands which are c8y global flags, so they never reach the binary
func (v *validator) checkFlags(manifest *extensions.Manifest, commands []extensions.BinaryCommand, exe string) {
	if len(commands) == 0 {
		return
	}
	if !manifest.Binary.CompletionEnabled() {
		v.add(SeverityInfo, CheckFlags, fileAlias, "the flags of the declared commands were not checked, as binary.completion is not '%s'. c8y handles its global flags (e.g. --output, -v), so the binary never receives flags with these names", extensions.CompletionCobra)
		return
	}

	for _, item := range commands {
		path := strings.Join(item.Path(), " ")
		location := v.commandPath(item.Path()...)
		flags, err := v.binaryFlags(exe, item.ExecArgs())
		if err != nil {
			v.add(SeverityWarning, CheckFlags, location, "could not read the flags of the command from the binary. %s", err)
			continue
		}
		// Group the binary flags by the c8y flag they clash with (e.g. -v and --verbose)
		clashes := map[string][]string{}
		names := []string{}
		for _, flag := range flags {
			var global *pflag.Flag
			if strings.HasPrefix(flag, "--") {
				global = v.opts.GlobalFlags.Lookup(flag[2:])
			} else {
				global = v.opts.GlobalFlags.ShorthandLookup(flag[1:])
			}
			if global == nil || globalFlagsHandledByBinary[global.Name] {
				continue
			}
			if _, ok := clashes[global.Name]; !ok {
				names = append(names, global.Name)
			}
			clashes[global.Name] = append(clashes[global.Name], flag)
		}
		for _, name := range names {
			global := v.opts.GlobalFlags.Lookup(name)
			globalName := "--" + global.Name
			if global.Shorthand != "" {
				globalName = "-" + global.Shorthand + "/" + globalName
			}
			v.add(SeverityWarning, CheckFlags, location,
				"flag %s of '%s' clashes with the c8y global flag %s, which c8y handles, so the binary never receives it. Rename the flag, or don't declare the command, so it is run through the mount",
				strings.Join(clashes[name], ", "), path, globalName)
		}
	}
}

// binaryFlags returns the flag names of a binary command, using cobra's __complete command
func (v *validator) binaryFlags(exe string, execArgs []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), v.opts.CompleteTimeout)
	defer cancel()

	args := append([]string{"__complete"}, execArgs...)
	args = append(args, "-")
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Env = append(os.Environ(),
		"C8Y_EXTENSION_NAME="+v.ext.Name(),
		"C8Y_EXTENSION_DIR="+v.ext.Path(),
	)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("timed out after %s", v.opts.CompleteTimeout)
		}
		return nil, err
	}

	flags := []string{}
	seen := map[string]bool{}
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		name, _, _ := strings.Cut(scanner.Text(), "\t")
		name = strings.TrimSpace(name)
		if strings.HasPrefix(name, "-") && len(name) > 1 && !seen[name] {
			seen[name] = true
			flags = append(flags, name)
		}
	}
	return flags, scanner.Err()
}

// CountFindings returns the number of errors and warnings
func CountFindings(findings []Finding) (errorCount int, warningCount int) {
	for _, f := range findings {
		switch f.Severity {
		case SeverityError:
			errorCount++
		case SeverityWarning:
			warningCount++
		}
	}
	return
}
