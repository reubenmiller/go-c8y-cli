package list

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/manifoldco/promptui"
	"github.com/reubenmiller/go-c8y-cli/pkg/c8ysession"
	createCmd "github.com/reubenmiller/go-c8y-cli/pkg/cmd/sessions/create"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/utilities/bellskipper"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type CmdList struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
	Config  func() (*config.Config, error)
	Client  func() (*c8y.Client, error)

	sessionFilter string
}

func NewCmdList(f *cmdutil.Factory) *CmdList {
	ccmd := &CmdList{
		factory: f,
		Config:  f.Config,
		Client:  f.Client,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a Cumulocity session",
		Long:  `Get a Cumulocity session`,
		Example: heredoc.Doc(`
			Example 1: Show an interactive list of all available sessions

			#> c8y sessions list

			Example 2: Select a session and filter the selection of session by the name "customer"

			#> export C8Y_SESSION=$( c8y sessions list --sessionFilter "customer" )
		`),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.sessionFilter, "sessionFilter", "", "Filter to be applied to the list of sessions even before the values can be selected")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func matchSession(session c8ysession.CumulocitySession, input string) bool {
	// strip url scheme
	uri := strings.ReplaceAll(session.Host, "https://", "")
	uri = strings.ReplaceAll(uri, "http://", "")

	name := strings.ToLower(fmt.Sprintf("#%02d %s %s %s %s",
		session.Index,
		filepath.Base(session.Path),
		uri,
		session.Tenant,
		session.Username,
	))
	input = strings.ToLower(input)

	searchTerms := strings.Split(input, " ")

	match := true
	for _, term := range searchTerms {
		if !strings.Contains(name, term) {
			match = false
		}
	}
	return match || input == ""
}

func (n *CmdList) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	sessions := &c8ysession.CumulocitySessions{}
	sessions.Sessions = make([]c8ysession.CumulocitySession, 0)

	subDirToSkip := "ignore"

	files := make([]string, 0)

	outputDir := cfg.GetSessionHomeDir()

	log.Infof("using c8y session folder: %s", outputDir)

	err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() && info.Name() == subDirToSkip {
			fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}

		// skip settings file
		if strings.HasPrefix(info.Name(), config.SettingsGlobalName+".") {
			return nil
		}

		log.Infof("Walking folder/file: %s", path)
		files = append(files, path)

		if session, err := createCmd.NewCumulocitySessionFromFile(path, log, cfg); err == nil {
			sessions.Sessions = append(sessions.Sessions, *session)
		}
		return nil
	})

	if err != nil {
		cmd.PrintErrf("Failed to walk directory. %s", err)
	}

	// Add index numbers
	for i := range sessions.Sessions {
		sessions.Sessions[i].Index = i + 1
	}

	// template.Fun
	funcMap := promptui.FuncMap

	funcMap["hide"] = func(v interface{}) string {
		if !strings.EqualFold(os.Getenv(c8y.EnvVarLoggerHideSensitive), "true") {
			return fmt.Sprintf("%v", v)
		}
		return "*****"
	}

	funcMap["hideUser"] = func(v interface{}) string {
		msg := fmt.Sprintf("%v", v)
		if !strings.EqualFold(os.Getenv(c8y.EnvVarLoggerHideSensitive), "true") {
			return msg
		}
		if os.Getenv("USERNAME") != "" {
			msg = strings.ReplaceAll(msg, os.Getenv("USERNAME"), "******")
		}
		return msg
	}

	templates := &promptui.SelectTemplates{
		// Label:    "{{ .Host }}?",
		Active:   `-> {{ printf "#%02d: %-25s" .Index .Name | cyan }} {{ .Host | hide | magenta }} {{ printf "(%s/" .Tenant | hide | red }}{{ printf "%s)" .Username | hide | red }}`,
		Inactive: `   {{ printf "#%02d: %-25s" .Index .Name | cyan }} {{ .Host | hide | magenta }} {{ printf "(%s/" .Tenant | hide | red }}{{ printf "%s)" .Username | hide | red }}`,
		Selected: "{{ .Path | hideUser }}",
		FuncMap:  funcMap,
		Details: `
--------- Details ----------
{{ "File:" | faint }}	{{ .Path | hideUser }}
{{ "Host:" | faint }}	{{ .Host | hide }}
{{ "Tenant:" | faint }}	{{ .Tenant | hide }}
{{ "Username:" | faint }}	{{ .Username | hide }}
`,
	}
	templates.Help = fmt.Sprintf(`{{ "Use the arrow keys to navigate (some terminals require you to hold shift as well):" | faint }} {{ .NextKey | faint }} ` +
		`{{ .PrevKey | faint }} {{ .PageDownKey | faint }} {{ .PageUpKey | faint }} ` +
		`{{ if .Search }} {{ "and" | faint }} {{ .SearchKey | faint }} {{ "toggles search" | faint }}{{ end }}`)

	filteredSessions := make([]c8ysession.CumulocitySession, 0)
	sessionIndex := 1
	for _, session := range sessions.Sessions {
		if matchSession(session, n.sessionFilter) {
			session.Index = sessionIndex
			filteredSessions = append(filteredSessions, session)
			sessionIndex++
		}
	}

	searcher := func(input string, index int) bool {
		session := filteredSessions[index]
		return matchSession(session, input)
	}

	prompt := promptui.Select{
		Stdout:            &bellskipper.BellSkipper{}, // Workaround to pervent the terminal bell on MacOS
		HideSelected:      true,
		IsVimMode:         false,
		StartInSearchMode: false,
		Label:             "Select a Cumulocity Session",
		Items:             filteredSessions,
		Templates:         templates,
		Size:              10,
		Searcher:          searcher,
	}

	// Customize select keys
	prompt.Keys = &promptui.SelectKeys{
		Prev:     promptui.Key{Code: promptui.KeyPrev, Display: promptui.KeyPrevDisplay + ""},
		Next:     promptui.Key{Code: promptui.KeyNext, Display: promptui.KeyNextDisplay + ""},
		PageUp:   promptui.Key{Code: promptui.KeyBackward, Display: promptui.KeyBackwardDisplay + ""},
		PageDown: promptui.Key{Code: promptui.KeyForward, Display: promptui.KeyForwardDisplay + ""},
		Search:   promptui.Key{Code: '/', Display: "/"},
	}

	var idx int
	var result string

	if len(filteredSessions) == 1 {
		log.Info("Only 1 session found. Selecting it automatically")
		idx = 0
		result = filteredSessions[0].Path
		err = nil
	} else {
		idx, result, err = prompt.Run()
	}

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}

	// check if the user cancelled the select (i.e. ctrl+c)
	if result == "" {
		// required inorder to flush the screen buffer
		fmt.Println("")
		return nil
	}

	if idx >= 0 && idx < len(filteredSessions) {
		fmt.Printf("%s", filteredSessions[idx].Path)
	} else {
		fmt.Println("")
	}

	return nil
}
