package extension

import (
	"encoding/json"
	"errors"
	"fmt"
	_io "io"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/internal/ghrepo"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/git"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/prompt"
	"github.com/spf13/cobra"
)

var ExtPrefix = "c8y-"

func NewCmdExtension(f *cmdutil.Factory) *cobra.Command {
	m := f.ExtensionManager
	io := f.IOStreams
	prompter := prompt.NewPrompt(nil)

	extCmd := cobra.Command{
		Use:   "extensions",
		Short: "Manage c8y extensions",
		Long: heredoc.Docf(`
			go-c8y-cli extensions are repositories that provide additional c8y commands, aliases,
			views and templates.

			The name of the extension repository must start with "c8y-". Its commands are available
			under %[1]sc8y [extname]%[1]s, and can be provided by API specifications (api/), scripts
			(commands/), or a compiled binary declared in extension.yaml.

			Extensions are installed from git repositories, GitHub releases (binary extensions), or
			local folders. Extensions in system-wide folders (installed by package managers) are
			also loaded, see the extensions.systemDirs setting.

			An extension cannot override any of the core c8y commands.

			See the list of available extensions at https://github.com/topics/c8y-extension
		`, "`"),
		Aliases: []string{"extension"},
	}

	extCmd.AddCommand(
		&cobra.Command{
			Use:     "list",
			Short:   "List installed extension commands",
			Aliases: []string{"ls"},
			Args:    cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmds := m().List()
				if len(cmds) == 0 {
					return cmderrors.NewSystemError("no installed extensions found")
				}

				for _, c := range cmds {
					ext := map[string]interface{}{}
					var repo string
					if u, err := git.ParseURL(c.URL()); err == nil {
						repo = u.String()
						if r, err := ghrepo.FromURL(u); err == nil {
							repo = ghrepo.FullName(r)
						}
					}

					version := displayExtensionVersion(c, c.CurrentVersion())

					ext["name"] = c.Name()
					ext["repo"] = repo
					ext["version"] = version
					ext["pinned"] = c.IsPinned()
					ext["path"] = c.Path()
					ext["isLocal"] = c.IsLocal()
					ext["isBinary"] = c.IsBinary()
					ext["kind"] = string(c.Kind())
					ext["executable"] = c.Executable()

					providerNames := []string{}
					if manifest, err := c.Manifest(); err == nil {
						providers, _ := manifest.ValidSessionProviders()
						for _, provider := range providers {
							providerNames = append(providerNames, provider.Name)
						}
					}
					ext["sessionProviders"] = providerNames

					rowText, err := json.Marshal(ext)
					if err != nil {
						return cmderrors.NewUserError("Settings error. ", err)
					}

					if err := f.WriteOutputWithoutPropertyGuess(rowText, cmdutil.OutputContext{}); err != nil {
						return cmderrors.NewSystemError("Failed to write to console. ", err)
					}
				}
				return nil
			},
		},
		func() *cobra.Command {
			var pinFlag string
			var nameFlag string
			cmd := &cobra.Command{
				Use:   "install <repository>",
				Short: "Install a c8y extensions from a repository",
				Long: heredoc.Doc(`
					Install a GitHub repository locally as a Cumulocity CLI extension.

					The repository argument can be specified in "owner/repo" format as well as a full URL.
					The URL format is useful when the repository is not hosted on github.com.

					If the latest GitHub release (or the release given by --pin) has an asset for the current
					platform, then the binary extension is downloaded from the release. The download is verified
					using the release's checksums.txt file if available. Otherwise the repository is installed
					using git. Set GH_TOKEN or GITHUB_TOKEN to avoid GitHub API rate limits.

					To install an extension in development from the current directory, use "." as the
					value of the repository argument.

					See the list of available extensions at https://github.com/topics/c8y-extension
				`),
				Example: heredoc.Doc(`
					$ c8y extensions install owner/c8y-extension
					$ c8y extensions install owner/c8y-extension --pin v1.2.0
					$ c8y extensions install https://git.example.com/owner/c8y-extension
					$ c8y extensions install .
				`),
				Args: func(cmd *cobra.Command, args []string) error {
					err := cobra.MinimumNArgs(1)(cmd, args)
					if err != nil {
						return fmt.Errorf("%w. must specify a repository to install from", err)
					}
					return err
				},
				RunE: func(cmd *cobra.Command, args []string) error {
					cfg, err := f.Config()
					if err != nil {
						return err
					}
					if fileInfo, err := os.Stat(args[0]); err == nil && fileInfo.IsDir() {
						if pinFlag != "" {
							return fmt.Errorf("local extensions cannot be pinned")
						}

						sourcePath, err := filepath.Abs(args[0])
						if err != nil {
							return err
						}

						if err := m().InstallLocal(sourcePath, nameFlag); err != nil {
							return err
						}
						if io.IsStdoutTTY() {
							successStr := "Installed"
							if cfg.DryRun() {
								successStr = "Would have installed"
							}
							cs := io.ColorScheme()
							fmt.Fprintf(io.Out, "%s %s extension %s\n", cs.SuccessIcon(), successStr, filepath.Base(sourcePath))
						}
						if !cfg.DryRun() {
							localName := nameFlag
							if localName == "" {
								localName = filepath.Base(sourcePath)
							}
							printValidationSummary(m(), io, localName)
						}
						return nil
					}

					repo, err := ghrepo.FromFullName(args[0])
					if err != nil {
						return err
					}

					extName := nameFlag
					skipNameCheck := true
					if extName == "" {
						extName = repo.RepoName()
						skipNameCheck = false
					} else {
						if !strings.HasPrefix(extName, ExtPrefix) {
							extName = ExtPrefix + extName
						}
					}
					if err := checkValidExtension(cmd.Root(), m(), extName, skipNameCheck); err != nil {
						return err
					}

					cs := io.ColorScheme()
					if err := m().Install(repo, extName, pinFlag); err != nil {
						if errors.Is(err, ErrReleaseNotFound) {
							return fmt.Errorf("%s Could not find a release of %s for %s",
								cs.FailureIcon(), args[0], cs.Cyan(pinFlag))
						} else if errors.Is(err, ErrCommitNotFound) {
							return fmt.Errorf("%s %s does not exist in %s",
								cs.FailureIcon(), cs.Cyan(pinFlag), args[0])
						} else if errors.Is(err, ErrRepositoryNotFound) {
							return fmt.Errorf("%s Could not find extension '%s' on host %s",
								cs.FailureIcon(), args[0], repo.RepoHost())
						}
						return err
					}

					if io.IsStdoutTTY() {
						fmt.Fprintf(io.Out, "%s Installed extension %s\n", cs.SuccessIcon(), extName)
						if pinFlag != "" {
							fmt.Fprintf(io.Out, "%s Pinned extension at %s\n", cs.SuccessIcon(), cs.Cyan(pinFlag))
						}
					}
					if !cfg.DryRun() {
						printValidationSummary(m(), io, extName)
					}
					return nil
				},
			}
			cmd.Flags().StringVar(&pinFlag, "pin", "", "pin extension to a release tag or commit ref")
			cmd.Flags().StringVar(&nameFlag, "name", "", "use custom name for the extension")
			completion.WithOptions(
				cmd,
				completion.MarkLocalFlag(),
			)
			return cmd
		}(),
		func() *cobra.Command {
			var flagAll bool
			cmd := &cobra.Command{
				Use:   "update {<name> | --all}",
				Short: "Update installed extensions",
				ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
					cmds := m().List()
					names := []string{}
					for _, c := range cmds {
						names = append(names, c.Name())
					}
					return names, cobra.ShellCompDirectiveNoFileComp
				},
				Args: func(cmd *cobra.Command, args []string) error {
					if len(args) == 0 && !flagAll {
						return cmderrors.NewUserError("specify an extension to update or `--all`")
					}
					if len(args) > 0 && flagAll {
						return cmderrors.NewUserError("cannot use `--all` with extension name")
					}
					if len(args) > 1 {
						return cmderrors.NewUserError("too many arguments")
					}
					return nil
				},
				RunE: func(cmd *cobra.Command, args []string) error {
					var name string
					if len(args) > 0 {
						name = normalizeExtensionSelector(args[0])
					}
					cfg, cfgErr := f.Config()
					if cfgErr != nil {
						return cfgErr
					}
					if cfg.DryRun() {
						m().EnableDryRunMode()
					}
					cs := io.ColorScheme()
					err := m().Upgrade(name, cfg.Force())
					if err != nil && !errors.Is(err, ErrUpToDate) {
						if name != "" {
							fmt.Fprintf(io.ErrOut, "%s Failed updating extension %s: %s\n", cs.FailureIcon(), name, err)
						} else if errors.Is(err, ErrNoExtensionsInstalled) {
							return cmderrors.NewSystemError("no installed extensions found")
						} else {
							fmt.Fprintf(io.ErrOut, "%s Failed updating extensions\n", cs.FailureIcon())
						}
						return cmderrors.NewSilentError()
					}
					if io.IsStdoutTTY() {
						successStr := "Successfully"

						if cfg.DryRun() {
							successStr = "Would have"
						}
						if errors.Is(err, ErrUpToDate) {
							fmt.Fprintf(io.Out, "%s Extension already up to date\n", cs.SuccessIcon())
						} else if name != "" {
							fmt.Fprintf(io.Out, "%s %s updated extension %s\n", cs.SuccessIcon(), successStr, name)
						} else {
							fmt.Fprintf(io.Out, "%s %s updated extensions\n", cs.SuccessIcon(), successStr)
						}
					}
					return nil
				},
			}
			cmd.Flags().BoolVar(&flagAll, "all", false, "Update all extensions")
			completion.WithOptions(
				cmd,
				completion.MarkLocalFlag(),
			)
			return cmd
		}(),
		&cobra.Command{
			Use:   "delete <name>",
			Short: "Remove an installed extension",
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				cmds := m().List()
				names := []string{}
				for _, c := range cmds {
					names = append(names, c.Name())
				}
				return names, cobra.ShellCompDirectiveNoFileComp
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, cfgErr := f.Config()
				if cfgErr != nil {
					return cfgErr
				}
				if cfg.DryRun() {
					m().EnableDryRunMode()
				}

				inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
				if err != nil {
					return err
				}
				var iter iterator.Iterator
				_, input, err := flags.WithPipelineIterator(&flags.PipelineOptions{
					Name:     "name",
					Aliases:  []string{"name"},
					Disabled: inputIterators.PipeOptions.Disabled,
					Values:   args, // manual values
					Required: false,
				})(cmd, inputIterators)

				if err != nil {
					return &flags.ParameterError{
						Name: "Name",
						Err:  fmt.Errorf("missing required argument or pipeline input. %w", flags.ErrParameterMissing),
					}
				}

				extName := ""
				switch v := input.(type) {
				case iterator.Iterator:
					iter = v
				default:
					return fmt.Errorf("unknown iterator type")
				}

				bounded := iter.IsBound()
				errors := []error{}
				for {
					responseText, _, err := iter.GetNext()
					if err != nil {
						if err == _io.EOF {
							break
						}
						return err
					}

					extName = normalizeExtensionSelector(string(responseText))

					if err := m().Remove(extName); err != nil {
						errors = append(errors, err)
						continue
					}

					if io.IsStdoutTTY() {
						successStr := "Removed"
						if cfg.DryRun() {
							successStr = "Would have removed"
						}
						cs := io.ColorScheme()
						fmt.Fprintf(io.Out, "%s %s extension %s\n", cs.SuccessIconWithColor(cs.Red), successStr, extName)
					}

					if !bounded {
						break
					}
				}

				switch count := len(errors); count {
				case 0:
					return nil
				case 1:
					return errors[0]
				default:
					errorMessages := []string{}
					for _, m := range errors {
						errorMessages = append(errorMessages, m.Error())
					}
					return fmt.Errorf("failed to remove %d extension/s. errors=%v", count, strings.Join(errorMessages, ", "))
				}
			},
		},
		func() *cobra.Command {
			promptCreate := func() (string, extensions.ExtTemplateType, error) {
				extName, err := prompter.Input("Extension name", "", true, false)
				if err != nil {
					return extName, -1, err
				}
				return extName, extensions.ExtTemplateType(extensions.GitTemplateType), err
			}
			var flagType string
			cmd := &cobra.Command{
				Use:   "create [<name>]",
				Short: "Create a new extension",
				Example: heredoc.Doc(`
					# Use interactively
					c8y extensions create

					# Create a script-based extension
					c8y extensions create foobar

					# Create a binary extension written in Go
					c8y extensions create foobar --precompiled go

					# Create a binary extension written in another language
					c8y extensions create foobar --precompiled other
				`),
				Args: cobra.MaximumNArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					if cmd.Flags().Changed("precompiled") {
						if flagType != "go" && flagType != "other" {
							return cmderrors.NewUserError(fmt.Sprintf("value for --precompiled must be 'go' or 'other'. Got '%s'", flagType))
						}
					}
					var extName string
					var err error
					tmplType := extensions.GitTemplateType
					if len(args) == 0 {
						if io.IsStdoutTTY() {
							extName, tmplType, err = promptCreate()
							if err != nil {
								return fmt.Errorf("could not prompt: %w", err)
							}
						}
					} else {
						extName = args[0]
						if flagType == "go" {
							tmplType = extensions.GoBinTemplateType
						} else if flagType == "other" {
							tmplType = extensions.OtherBinTemplateType
						}
					}

					var fullName string

					if strings.HasPrefix(extName, ExtPrefix) {
						fullName = extName
						extName = extName[len(ExtPrefix):]
					} else {
						fullName = ExtPrefix + extName
					}
					if err := m().Create(fullName, tmplType); err != nil {
						return err
					}
					if !io.IsStdoutTTY() {
						return nil
					}

					var goBinChecks string

					steps := fmt.Sprintf(
						"- run 'cd %[1]s; c8y extensions install .; c8y %[2]s --help' to see your new extension in action",
						fullName, extName)

					cs := io.ColorScheme()
					if tmplType == extensions.GoBinTemplateType {
						if _, err := os.Stat(filepath.Join(fullName, "bin", fullName)); err == nil {
							goBinChecks = heredoc.Docf(`
							%[1]s Downloaded Go dependencies
							%[1]s Built %[2]s binary
							`, cs.SuccessIcon(), fullName)
						} else {
							goBinChecks = heredoc.Docf(`
							%[1]s The binary was not built (see the messages above). Run 'go mod tidy' and 'go build -o bin/%[2]s .'
							`, cs.WarningIcon(), fullName)
						}
						steps = heredoc.Docf(`
						- run 'cd %[1]s; c8y extensions install .; c8y %[2]s services list' to see your new extension in action
						- use 'go build -o bin/%[1]s . && c8y %[2]s services list' to see changes in your code as you develop
						- push a tag starting with "v" to publish release archives with GoReleaser`, fullName, extName)
					} else if tmplType == extensions.OtherBinTemplateType {
						steps = heredoc.Docf(`
						- compile a binary to bin/%[1]s, then run 'cd %[1]s; c8y extensions install .; c8y %[2]s --help'
						- fill in script/build.sh to build the release archives
						- push a tag starting with "v" to publish a release`, fullName, extName)
					}
					link := "https://c8y.app/docs/tutorials/extensions/"
					out := heredoc.Docf(`
						%[1]s Created directory %[2]s
						%[1]s Initialized git repository
						%[1]s Set up extension scaffolding
						%[6]s
						%[2]s is ready for development!

						%[4]s
						%[5]s
						- commit and push your code to a git repository to share your extension with others.

						For more information on writing extensions:
						  %[3]s
					`, cs.SuccessIcon(), fullName, link, cs.Bold("Next Steps"), steps, goBinChecks)
					fmt.Fprint(io.Out, out)
					return nil
				},
			}
			cmd.Flags().StringVar(&flagType, "precompiled", "", "Create a precompiled (binary) extension. Possible values: go, other")
			completion.WithOptions(
				cmd,
				completion.MarkLocalFlag(),
			)
			return cmd
		}(),
	)

	extCmd.AddCommand(newCmdValidate(f))

	for _, cmd := range extCmd.Commands() {
		cmdutil.DisableAuthCheck(cmd)
	}
	return &extCmd
}

func checkValidExtension(rootCmd *cobra.Command, m extensions.ExtensionManager, extName string, skipNameCheck bool) error {
	// Allow prefix anyway in the extension name
	if !skipNameCheck {
		if !strings.HasPrefix(extName, ExtPrefix) && !strings.Contains(extName, ExtPrefix) {
			return fmt.Errorf("extension repository name must start with `%s`", ExtPrefix)
		}
	}

	commandName := strings.TrimPrefix(extName, ExtPrefix)
	if c, _, err := rootCmd.Traverse([]string{commandName}); err != nil {
		return err
	} else if c != rootCmd {
		return fmt.Errorf("%q matches the name of a built-in command", commandName)
	}

	for _, ext := range m.List() {
		// Extensions in the user's extension folder take precedence over system extensions
		if ext.Name() == commandName && ext.Kind() != extensions.KindSystem {
			return fmt.Errorf("there is already an installed extension that provides the %q command", commandName)
		}
	}

	return nil
}

func normalizeExtensionSelector(n string) string {
	if idx := strings.IndexRune(n, '/'); idx >= 0 {
		n = n[idx+1:]
	}
	return strings.TrimPrefix(n, ExtPrefix)
}

func displayExtensionVersion(ext extensions.Extension, version string) string {
	if !ext.IsBinary() && len(version) > 8 {
		return version[:8]
	}
	return version
}

// findInstalledExtension returns the installed extension with the given name (with or without the c8y- prefix)
func findInstalledExtension(m extensions.ExtensionManager, name string) extensions.Extension {
	name = normalizeExtensionSelector(name)
	for _, ext := range m.List() {
		if ext.Name() == name {
			return ext
		}
	}
	return nil
}

// printValidationSummary validates an installed extension (without running its binary), and
// prints the number of problems to stderr
func printValidationSummary(m extensions.ExtensionManager, io *iostreams.IOStreams, name string) {
	ext := findInstalledExtension(m, name)
	if ext == nil {
		return
	}
	errorCount, warningCount := CountFindings(ValidateExtension(ext, ValidateOptions{SkipFlagCheck: true}))
	if errorCount+warningCount == 0 {
		return
	}
	cs := io.ColorScheme()
	fmt.Fprintf(io.ErrOut, "%s Extension %s has %d error(s) and %d warning(s). Run 'c8y extensions validate %s' for details\n", cs.WarningIcon(), ext.Name(), errorCount, warningCount, ext.Name())
}

func newCmdValidate(f *cmdutil.Factory) *cobra.Command {
	var strict bool
	cmd := &cobra.Command{
		Use:   "validate [<name> | <path>]",
		Short: "Validate an extension",
		Long: heredoc.Doc(`
			Check an extension for problems, for example typos in extension.yaml, a missing binary,
			commands which are ignored because another command has the same name, or flags of declared
			binary commands which are c8y global flags (so the binary never receives them).

			The argument is the folder of an extension, or the name of an installed extension. The
			current folder is used if no argument is given.

			The command exits with a non-zero exit code if there are errors, or warnings when using --strict,
			so it can be used in CI.
		`),
		Example: heredoc.Doc(`
			$ c8y extensions validate
			Validate the extension in the current folder

			$ c8y extensions validate myext
			Validate the installed extension myext

			$ c8y extensions validate . --strict
			Fail on warnings as well, e.g. in a release workflow
		`),
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			names := []string{}
			for _, ext := range f.ExtensionManager().List() {
				names = append(names, ext.Name())
			}
			return names, cobra.ShellCompDirectiveDefault
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "."
			if len(args) > 0 {
				target = args[0]
			}

			var ext extensions.Extension
			if info, err := os.Stat(target); err == nil && info.IsDir() {
				dir, err := filepath.Abs(target)
				if err != nil {
					return err
				}
				ext = &Extension{path: dir, isLocal: true, manifest: &manifestCache{}}
			} else if installed := findInstalledExtension(f.ExtensionManager(), target); installed != nil {
				ext = installed
			} else {
				return cmderrors.NewUserError(fmt.Sprintf("'%s' is neither a folder nor an installed extension", target))
			}

			// Show the columns in a fixed order (by default the table sorts them, so the severity
			// column would be cut off by long messages)
			if flag := cmd.Flags().Lookup("select"); flag != nil && !flag.Changed {
				if err := flag.Value.Set("severity,check,location,message"); err == nil {
					flag.Changed = true
				}
			}

			findings := ValidateExtension(ext, ValidateOptions{GlobalFlags: cmd.Root().PersistentFlags()})
			for _, finding := range findings {
				row, err := json.Marshal(finding)
				if err != nil {
					return err
				}
				if err := f.WriteOutputWithoutPropertyGuess(row, cmdutil.OutputContext{}); err != nil {
					return err
				}
			}

			io := f.IOStreams
			cs := io.ColorScheme()
			errorCount, warningCount := CountFindings(findings)
			if len(findings) == 0 {
				fmt.Fprintf(io.ErrOut, "%s No problems found in extension %s\n", cs.SuccessIcon(), ext.Name())
				return nil
			}
			fmt.Fprintf(io.ErrOut, "Extension %s: %d error(s), %d warning(s), %d info\n", ext.Name(), errorCount, warningCount, len(findings)-errorCount-warningCount)
			if errorCount > 0 || (strict && warningCount > 0) {
				return cmderrors.NewSilentErrorWithExitCode(cmderrors.ExitUserError, fmt.Errorf("extension validation failed"))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&strict, "strict", false, "Fail on warnings as well")
	completion.WithOptions(cmd, completion.MarkLocalFlag())
	return cmd
}
