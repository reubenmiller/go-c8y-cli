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
			GitHub CLI extensions are repositories that provide additional c8y commands.

			The name of the extension repository must start with "c8y-" and it must contain an
			executable of the same name. All arguments passed to the %[1]sc8y <extname>%[1]s invocation
			will be forwarded to the %[1]sc8y-<extname>%[1]s executable of the extension.

			An extension cannot override any of the core c8y commands.

			See the list of available extensions at <https://github.com/topics/c8y-extension>.
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
				cfg, err := f.Config()

				if err != nil {
					return err
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

					rowText, err := json.Marshal(ext)
					if err != nil {
						return cmderrors.NewUserError("Settings error. ", err)
					}

					f.WriteJSONToConsole(cfg, cmd, "", rowText)
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

					To install an extension in development from the current directory, use "." as the
					value of the repository argument.

					See the list of available extensions at <https://github.com/topics/c8y-extension>.
				`),
				Example: heredoc.Doc(`
					$ c8y extensions install owner/c8y-extension
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
				// Support only one type of extension to start out with
				// options := []string{"Script (Bash, Ruby, Python, etc)", "Go", "Other Precompiled (C++, Rust, etc)"}
				// options := []string{"Script (Bash, Ruby, Python, etc)"}

				// extTmplType, err := prompt.Select("What kind of extension?", options, options[0])

				// TODO: Select type
				// _ = extTmplType
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
				`),
				Args: cobra.MaximumNArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					if cmd.Flags().Changed("precompiled") {
						if flagType != "go" && flagType != "other" {
							return cmderrors.NewUserError("value for --precompiled must be 'go' or 'other'. Got '%s'", flagType)
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
						"- run 'cd %[1]s; c8y extensions install .; c8y %[2]s list' to see your new extension in action",
						fullName, extName)

					cs := io.ColorScheme()
					if tmplType == extensions.GoBinTemplateType {
						goBinChecks = heredoc.Docf(`
						%[1]s Downloaded Go dependencies
						%[1]s Built %[2]s binary
						`, cs.SuccessIcon(), fullName)
						steps = heredoc.Docf(`
						- run 'cd %[1]s; c8y extensions install .; c8y %[2]s' to see your new extension in action
						- use 'go build && c8y %[2]s' to see changes in your code as you develop`, fullName, extName)
					} else if tmplType == extensions.OtherBinTemplateType {
						steps = heredoc.Docf(`
						- run 'cd %[1]s; c8y extensions install .' to install your extension locally
						- fill in script/build.sh with your compilation script for automated builds
						- compile a %[1]s binary locally and run 'c8y %[2]s' to see changes`, fullName, extName)
					}
					link := "https://docs.github.com/github-cli/github-cli/creating-github-cli-extensions"
					out := heredoc.Docf(`
						%[1]s Created directory %[2]s
						%[1]s Initialized git repository
						%[1]s Set up extension scaffolding
						%[6]s
						%[2]s is ready for development!

						%[4]s
						%[5]s
						- commit and push your code to a git repository to share your extension with others.
						  To share on github you can use the 'gh' command: 'gh repo create'

						For more information on writing extensions:
						%[3]s
					`, cs.SuccessIcon(), fullName, link, cs.Bold("Next Steps"), steps, goBinChecks)
					fmt.Fprint(io.Out, out)
					return nil
				},
			}
			// FUTURE: Support other types of extensions in the future
			// cmd.Flags().StringVar(&flagType, "precompiled", "", "Create a precompiled extension. Possible values: go, other")
			completion.WithOptions(
				cmd,
				completion.MarkLocalFlag(),
			)
			return cmd
		}(),
	)

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
		if ext.Name() == commandName {
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
