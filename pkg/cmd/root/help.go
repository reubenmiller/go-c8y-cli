package root

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Help levels which can be given to the --help flag, i.e. --help all
const (
	HelpLevelAll        = "all"
	HelpLevelLocal      = "local"
	HelpLevelGlobal     = "global"
	HelpLevelCategories = "category"
)

// annotationHelpCategory is the flag annotation storing the help category of a global flag
const annotationHelpCategory = "c8y.help.category"

// helpCategoryOther is used for global flags without a category
const helpCategoryOther = "other"

// HelpCategory is a group of global flags which can be shown on its own, i.e. --help output
type HelpCategory struct {
	Name        string
	Description string
	Flags       []string
}

// HelpCategories are the categories of the global flags (in display order)
var HelpCategories = []HelpCategory{
	{Name: "session", Description: "Session and authentication", Flags: []string{"session", "sessionUsername", "sessionPassword", "sessionMode"}},
	{Name: "output", Description: "Output format, filtering and selection", Flags: []string{"output", "outputFile", "outputFileRaw", "outputTemplate", "select", "filter", "view", "compact", "raw", "flatten", "noColor"}},
	{Name: "pagination", Description: "Paging through results", Flags: []string{"pageSize", "currentPage", "totalPages", "includeAll", "withTotalPages", "withTotalElements"}},
	{Name: "http", Description: "HTTP requests (headers, proxy, timeouts and retries)", Flags: []string{"header", "customQueryParam", "noAccept", "proxy", "noProxy", "timeout", "retries", "insecure"}},
	{Name: "cache", Description: "Response caching", Flags: []string{"cache", "noCache", "cacheTTL", "cacheBodyPaths"}},
	{Name: "dry", Description: "Dry run (show requests without sending them)", Flags: []string{"dry", "dryFormat"}},
	{Name: "pipeline", Description: "Piped input, workers and batching", Flags: []string{"nullInput", "allowEmptyPipe", "workers", "maxJobs", "delay", "delayBefore", "abortOnErrors", "progress", "noProgress"}},
	{Name: "confirm", Description: "Confirmation prompts", Flags: []string{"force", "confirm", "confirmText"}},
	{Name: "errors", Description: "Error handling and exit codes", Flags: []string{"withError", "silentStatusCodes", "silentExit"}},
	{Name: "logging", Description: "Logging and the activity log", Flags: []string{"verbose", "debug", "noLog", "logMessage"}},
	{Name: "help", Description: "Help and examples", Flags: []string{"help", "examples"}},
}

// helpValue is the value of the --help flag. It acts like a boolean flag (cobra requires
// it to be a bool), but it also accepts an optional help level, i.e. --help=local or --help local
type helpValue struct {
	enabled bool
	level   string
}

func (h *helpValue) String() string { return strconv.FormatBool(h.enabled) }

// Type must be bool, as cobra reads the help flag using GetBool
func (h *helpValue) Type() string { return "bool" }

func (h *helpValue) IsBoolFlag() bool { return true }

func (h *helpValue) Set(s string) error {
	if v, err := strconv.ParseBool(s); err == nil {
		h.enabled = v
		h.level = ""
		return nil
	}
	if !isHelpLevel(s) {
		return fmt.Errorf("invalid help level %q. Valid values: %s", s, strings.Join(HelpLevelNames(), ", "))
	}
	h.enabled = true
	h.level = s
	return nil
}

// HelpLevelNames returns all the valid help levels and categories
func HelpLevelNames() []string {
	names := []string{HelpLevelAll, HelpLevelLocal, HelpLevelGlobal, HelpLevelCategories}
	for _, c := range HelpCategories {
		names = append(names, c.Name)
	}
	return append(names, helpCategoryOther)
}

// isHelpLevel returns true if the value is a valid (comma separated) help level, i.e. local,output
func isHelpLevel(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	valid := HelpLevelNames()
	for _, part := range strings.Split(value, ",") {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "categories" {
			continue
		}
		if !slices.Contains(valid, part) {
			return false
		}
	}
	return true
}

// helpSelection describes which parts of the help should be shown
type helpSelection struct {
	local          bool
	allGlobal      bool
	categories     map[string]bool
	listCategories bool
}

func (s helpSelection) showsAllFlags() bool {
	return s.local && s.allGlobal
}

func (s helpSelection) showsCategory(name string) bool {
	return s.allGlobal || s.categories[name]
}

// parseHelpLevel converts a (comma separated) help level into a help selection.
// An empty or invalid level shows all the flags
func parseHelpLevel(level string) helpSelection {
	sel := helpSelection{categories: map[string]bool{}}
	if !isHelpLevel(level) {
		sel.local = true
		sel.allGlobal = true
		return sel
	}
	for _, part := range strings.Split(level, ",") {
		switch part = strings.ToLower(strings.TrimSpace(part)); part {
		case HelpLevelAll:
			sel.local = true
			sel.allGlobal = true
		case HelpLevelLocal:
			sel.local = true
		case HelpLevelGlobal:
			sel.allGlobal = true
		case HelpLevelCategories, "categories":
			sel.listCategories = true
		default:
			sel.categories[part] = true
		}
	}
	return sel
}

// newHelpFlag adds the --help flag to the root command (as a global flag), so it can accept an optional help level
func newHelpFlag(cmd *cobra.Command) {
	value := &helpValue{}
	flag := cmd.PersistentFlags().VarPF(value, "help", "h", "Show help. Optionally give a level to limit the flags shown: all, local, global, category (list the categories), or one or more categories, i.e. --help local,output")
	flag.NoOptDefVal = "true"

	_ = cmd.RegisterFlagCompletionFunc("help", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		values := []string{
			HelpLevelAll + "\tShow all flags",
			HelpLevelLocal + "\tShow only the flags of the command (no global flags)",
			HelpLevelGlobal + "\tShow only the global flags",
			HelpLevelCategories + "\tList the categories of the global flags",
		}
		for _, c := range HelpCategories {
			values = append(values, c.Name+"\tGlobal flags: "+c.Description)
		}
		return values, cobra.ShellCompDirectiveNoFileComp
	})
}

// annotateHelpCategories stores the help category of each of the global flags
func annotateHelpCategories(flags *pflag.FlagSet) {
	for _, category := range HelpCategories {
		for _, name := range category.Flags {
			if flags.Lookup(name) != nil {
				_ = flags.SetAnnotation(name, annotationHelpCategory, []string{category.Name})
			}
		}
	}
}

func getHelpCategory(flag *pflag.Flag) string {
	if values := flag.Annotations[annotationHelpCategory]; len(values) > 0 {
		return values[0]
	}
	return helpCategoryOther
}

// getHelpLevel returns the help level given by the --help flag, otherwise the default help level
// from the settings (settings.defaults.help or C8Y_SETTINGS_DEFAULTS_HELP)
func getHelpLevel(f *cmdutil.Factory, cmd *cobra.Command) string {
	if flag := cmd.Flags().Lookup("help"); flag != nil {
		if v, ok := flag.Value.(*helpValue); ok && v.level != "" {
			return v.level
		}
	}
	if f != nil {
		if cfg, err := f.Config(); err == nil && cfg != nil {
			return cfg.GetDefaultHelpLevel()
		}
	}
	return ""
}

// commandFlags returns the flags of the command (local) and the global flags (persistent flags of the root command)
func commandFlags(c *cobra.Command) (local *pflag.FlagSet, global *pflag.FlagSet) {
	if !c.HasParent() {
		return c.LocalNonPersistentFlags(), c.PersistentFlags()
	}
	return c.LocalFlags(), c.InheritedFlags()
}

// filterFlags returns the visible flags matching the filter
func filterFlags(flags *pflag.FlagSet, include func(*pflag.Flag) bool) *pflag.FlagSet {
	out := pflag.NewFlagSet("", pflag.ContinueOnError)
	out.SortFlags = flags.SortFlags
	flags.VisitAll(func(flag *pflag.Flag) {
		if !flag.Hidden && include(flag) {
			out.AddFlag(flag)
		}
	})
	return out
}

func hasFlags(flags *pflag.FlagSet) bool {
	found := false
	flags.VisitAll(func(*pflag.Flag) { found = true })
	return found
}

func rpad(s string, padding int) string {
	return fmt.Sprintf("%-*s", padding, s)
}

// newHelpUsageFunc returns a usage function (based on cobra's default) which only shows the flags
// selected by the help level
func newHelpUsageFunc(f *cmdutil.Factory) func(*cobra.Command) error {
	return func(c *cobra.Command) error {
		return writeUsage(c.OutOrStderr(), c, parseHelpLevel(getHelpLevel(f, c)))
	}
}

func writeUsage(w io.Writer, c *cobra.Command, sel helpSelection) error {
	if sel.listCategories && !sel.local && !sel.allGlobal && len(sel.categories) == 0 {
		writeHelpCategories(w, c)
		return nil
	}

	fmt.Fprint(w, "Usage:")
	if c.Runnable() {
		fmt.Fprintf(w, "\n  %s", c.UseLine())
	}
	if c.HasAvailableSubCommands() {
		fmt.Fprintf(w, "\n  %s [command]", c.CommandPath())
	}
	if len(c.Aliases) > 0 {
		fmt.Fprintf(w, "\n\nAliases:\n")
		fmt.Fprintf(w, "  %s", c.NameAndAliases())
	}
	if c.HasExample() {
		fmt.Fprintf(w, "\n\nExamples:\n")
		fmt.Fprintf(w, "%s", c.Example)
	}
	if c.HasAvailableSubCommands() {
		writeSubCommands(w, c)
	}

	local, global := commandFlags(c)
	if sel.local {
		if flags := filterFlags(local, func(*pflag.Flag) bool { return true }); hasFlags(flags) {
			fmt.Fprintf(w, "\n\nFlags:\n")
			fmt.Fprint(w, strings.TrimRight(flags.FlagUsages(), " \t\n"))
		}
	}

	hiddenGlobal := writeGlobalFlags(w, global, sel)

	if sel.listCategories {
		fmt.Fprint(w, "\n")
		writeHelpCategories(w, c)
	}

	if c.HasHelpSubCommands() {
		fmt.Fprintf(w, "\n\nAdditional help topics:")
		for _, subcmd := range c.Commands() {
			if subcmd.IsAdditionalHelpTopicCommand() {
				fmt.Fprintf(w, "\n  %s %s", rpad(subcmd.CommandPath(), subcmd.CommandPathPadding()), subcmd.Short)
			}
		}
	}
	if c.HasAvailableSubCommands() {
		fmt.Fprintf(w, "\n\nUse \"%s [command] --help\" for more information about a command.", c.CommandPath())
	}
	if hiddenGlobal || !sel.local {
		fmt.Fprintf(w, "\n\nUse \"%s --help all\" to show all flags, or \"%s --help category\" to list the flag categories.", c.CommandPath(), c.CommandPath())
	}
	fmt.Fprintln(w)
	return nil
}

// writeGlobalFlags writes the global flags of the categories selected by the help level.
// It returns true if some of the global flags were not shown
func writeGlobalFlags(w io.Writer, global *pflag.FlagSet, sel helpSelection) (hiddenGlobal bool) {
	categories := append([]HelpCategory{}, HelpCategories...)
	categories = append(categories, HelpCategory{Name: helpCategoryOther, Description: "Other"})
	for _, category := range categories {
		flags := filterFlags(global, func(flag *pflag.Flag) bool { return getHelpCategory(flag) == category.Name })
		if !hasFlags(flags) {
			continue
		}
		if !sel.showsCategory(category.Name) {
			hiddenGlobal = true
			continue
		}
		fmt.Fprintf(w, "\n\nGlobal Flags (%s):\n", category.Name)
		fmt.Fprint(w, strings.TrimRight(flags.FlagUsages(), " \t\n"))
	}
	return hiddenGlobal
}

func writeSubCommands(w io.Writer, c *cobra.Command) {
	cmds := c.Commands()
	isVisible := func(subcmd *cobra.Command) bool {
		return subcmd.IsAvailableCommand() || subcmd.Name() == "help"
	}
	if len(c.Groups()) == 0 {
		fmt.Fprintf(w, "\n\nAvailable Commands:")
		for _, subcmd := range cmds {
			if isVisible(subcmd) {
				fmt.Fprintf(w, "\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
			}
		}
		return
	}
	for _, group := range c.Groups() {
		fmt.Fprintf(w, "\n\n%s", group.Title)
		for _, subcmd := range cmds {
			if subcmd.GroupID == group.ID && isVisible(subcmd) {
				fmt.Fprintf(w, "\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
			}
		}
	}
	if !c.AllChildCommandsHaveGroup() {
		fmt.Fprintf(w, "\n\nAdditional Commands:")
		for _, subcmd := range cmds {
			if subcmd.GroupID == "" && isVisible(subcmd) {
				fmt.Fprintf(w, "\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
			}
		}
	}
}

func writeHelpCategories(w io.Writer, c *cobra.Command) {
	fmt.Fprintf(w, "Help Levels:\n")
	levels := [][2]string{
		{HelpLevelAll, "Show all flags"},
		{HelpLevelLocal, "Show only the flags of the command"},
		{HelpLevelGlobal, "Show only the global flags"},
		{HelpLevelCategories, "List the categories of the global flags"},
	}
	for _, level := range levels {
		fmt.Fprintf(w, "  %s %s\n", rpad(level[0], 12), level[1])
	}
	fmt.Fprintf(w, "\nGlobal Flag Categories:\n")
	for _, category := range HelpCategories {
		fmt.Fprintf(w, "  %s %s\n", rpad(category.Name, 12), category.Description)
	}
	fmt.Fprintf(w, "\nLevels can be combined, i.e. \"%s --help local,output\".\n", c.CommandPath())
	fmt.Fprintf(w, "The default level can be set using the settings.defaults.help setting (or C8Y_SETTINGS_DEFAULTS_HELP env variable).\n")
}

// isHelpArg returns true if the argument is the help flag, i.e. -h, --help, --help=local
func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || strings.HasPrefix(arg, "--help=") || strings.HasPrefix(arg, "-h=")
}

// NormalizeHelpArgs supports giving the help level as a separate argument, i.e. "--help local"
// is converted to "--help=local". The arguments are only changed if the value is a valid help level,
// and the command is not an extension command (which handles its own help)
func NormalizeHelpArgs(rootCmd *cobra.Command, args []string) []string {
	if target, _, err := rootCmd.Find(args); err == nil && target != nil && (target.DisableFlagParsing || IsExtensionCommand(target)) {
		return args
	}
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			out = append(out, args[i:]...)
			break
		}
		if (arg == "--help" || arg == "-h") && i+1 < len(args) && isHelpLevel(args[i+1]) {
			out = append(out, "--help="+args[i+1])
			i++
			continue
		}
		out = append(out, arg)
	}
	return out
}
