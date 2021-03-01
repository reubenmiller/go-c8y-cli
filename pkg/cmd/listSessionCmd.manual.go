package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type listSessionCmd struct {
	*baseCmd

	sessionFilter string
}

func newListSessionCmd() *listSessionCmd {
	ccmd := &listSessionCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a Cumulocity session",
		Long:  `Get a Cumulocity session`,
		Example: `
			Example 1: Show an interactive list of all available sessions

			#> c8y sessions list

			Example 2: Select a session and filter the selection of session by the name "customer"

			#> export C8Y_SESSION=$( c8y sessions list --sessionFilter "customer" )
		`,
		RunE: ccmd.listSession,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.sessionFilter, "sessionFilter", "", "Filter to be applied to the list of sessions even before the values can be selected")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func matchSession(session CumulocitySession, input string) bool {
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

func (n *listSessionCmd) listSession(cmd *cobra.Command, args []string) error {
	config := &CumulocitySessions{}
	config.Sessions = make([]CumulocitySession, 0)

	subDirToSkip := "ignore"

	files := make([]string, 0)

	outputDir := getSessionHomeDir()

	Logger.Infof("using c8y session folder: %s", outputDir)

	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() && info.Name() == subDirToSkip {
			fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}

		// skip settings file
		if strings.HasPrefix(info.Name(), SettingsGlobalName+".") {
			return nil
		}

		Logger.Infof("Walking folder/file: %s", path)
		files = append(files, path)

		if session, err := NewCumulocitySessionFromFile(path); err == nil {
			config.Sessions = append(config.Sessions, *session)
		}
		return nil
	})

	if err != nil {
		cmd.PrintErrf("Failed to walk directory. %s", err)
	}

	// Add index numbers
	for i := range config.Sessions {
		config.Sessions[i].Index = i + 1
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

	filteredSessions := make([]CumulocitySession, 0)
	sessionIndex := 1
	for _, session := range config.Sessions {
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
		Stdout:            &bellSkipper{}, // Workaround to pervent the terminal bell on MacOS
		HideSelected:      true,
		IsVimMode:         false,
		StartInSearchMode: false,
		Label:             "Select a Cumulocity Session",
		Items:             filteredSessions,
		Templates:         templates,
		Size:              10,
		Searcher:          searcher,
	}

	var idx int
	var result string

	if len(filteredSessions) == 1 {
		Logger.Info("Only 1 session found. Selecting it automatically")
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

func (n *listSessionCmd) formatFilename(name string) string {
	if !strings.HasSuffix(name, ".json") {
		name = fmt.Sprintf("%s.json", name)
	}
	return name
}
