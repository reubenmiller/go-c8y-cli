package selectsession

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ysession"
	createCmd "github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/sessions/create"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/utilities/bellskipper"
)

const (
	esc        = "\033["
	noLineWrap = esc + "\x1b[?7l"
	doLineWrap = esc + "\x1b[?7h"
)

func customStyle(attr ...color.Attribute) func(interface{}) string {
	return func(v interface{}) string {
		return color.New(attr...).Sprintf("%v", v)
	}
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

// SelectSession select a Cumulocity session interactively
func SelectSession(io *iostreams.IOStreams, cfg *config.Config, log *logger.Logger, filter string) (sessionFile string, err error) {
	sessions := &c8ysession.CumulocitySessions{}
	sessions.Sessions = make([]c8ysession.CumulocitySession, 0)

	subDirToSkip := strings.ToLower(":ignore:" + config.ActivityLogDirName + ":")

	files := make([]string, 0)

	srcdir := cfg.GetSessionHomeDir()
	log.Infof("using c8y session folder: %s", srcdir)

	err = filepath.Walk(srcdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Prevent panic by handling failure accessing a path %q: %v", path, err)
			return err
		}
		if info.IsDir() && strings.Contains(subDirToSkip, ":"+strings.ToLower(info.Name())+":") {
			log.Printf("Ignoring dir: %+v", info.Name())
			return filepath.SkipDir
		}
		if info.IsDir() && info.Name() == ".git" {
			log.Printf("Ignoring dir: %+v", info.Name())
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		// skip settings file
		if strings.HasPrefix(info.Name(), config.SettingsGlobalName+".") ||
			strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		log.Infof("Walking folder/file: %s", path)
		files = append(files, path)

		if session, err := createCmd.NewCumulocitySessionFromFile(path, log, cfg); err == nil {
			sessions.Sessions = append(sessions.Sessions, *session)
		} else {
			log.Infof("Failed to read file: file=%s, err=%s", path, err)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	// Add index numbers
	for i := range sessions.Sessions {
		sessions.Sessions[i].Index = i + 1
	}

	funcMap := promptui.FuncMap
	funcMap["highlight"] = customStyle()

	funcMap["hide"] = func(v interface{}) string {
		if cfg.HideSensitive() {
			return "*****"
		}
		return fmt.Sprintf("%v", v)
	}

	funcMap["hideUser"] = func(v interface{}) string {
		msg := fmt.Sprintf("%v", v)

		if !cfg.HideSensitive() {
			return msg
		}
		if os.Getenv("USERNAME") != "" {
			msg = strings.ReplaceAll(msg, os.Getenv("USERNAME"), "******")
		}
		return msg
	}

	templates := &promptui.SelectTemplates{
		Active:   `▶ {{ printf "#%02d %4.4s" .Index .Extension | highlight }} {{ printf "%-30.29s" .Name | highlight }} {{ .Host | hide | highlight }} {{ printf "(%s/" .Tenant | hide | highlight }}{{ printf "%s)" .Username | hide | highlight }}`,
		Inactive: `  {{ printf "#%02d %4.4s" .Index .Extension | faint }} {{ printf "%-30.29s" .Name | cyan }} {{ .Host | hide | magenta }} {{ printf "(%s/" .Tenant | hide | red }}{{ printf "%s)" .Username | hide | red }}`,
		Selected: "{{ .Path | hideUser }}",
		FuncMap:  funcMap,
		Details: `
--------- Details ----------
{{ printf "%10s" "File:" | faint }}  {{ .Path | hideUser }}
{{ printf "%10s" "Host:" | faint }}  {{ .Host | hide }}
{{ printf "%10s" "Tenant:" | faint }}  {{ .Tenant | hide }}
{{ printf "%10s" "Username:" | faint }}  {{ .Username | hide }}
`,
	}
	templates.Help = `{{ "Use arrow keys (holding shift) to navigate" | faint }} {{ .NextKey | faint }} ` +
		`{{ .PrevKey | faint }} {{ .PageDownKey | faint }} {{ .PageUpKey | faint }} ` +
		`{{ if .Search }} {{ "and" | faint }} {{ .SearchKey | faint }} {{ "toggles search" | faint }}{{ end }}`

	filteredSessions := make([]c8ysession.CumulocitySession, 0)
	sessionIndex := 1
	for _, session := range sessions.Sessions {
		if session.Host == "" && session.Username == "" && session.Tenant == "" {
			continue
		}
		if matchSession(session, filter) {
			session.Index = sessionIndex
			filteredSessions = append(filteredSessions, session)
			sessionIndex++
		}
	}

	searcher := func(input string, index int) bool {
		session := filteredSessions[index]
		return matchSession(session, input)
	}

	// always enable color
	color.NoColor = false
	prompt := promptui.Select{
		Stdout:            bellskipper.NewBellSkipper(os.Stderr), // Workaround to pervent the terminal bell on MacOS
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
		// Prevent wrapping errors as promptui expects all templates to only take
		// up one line. When multiline templates are supported by promptui, then this can be removed
		// https://github.com/manifoldco/promptui/issues/92
		fmt.Fprint(os.Stderr, noLineWrap)
		idx, result, err = prompt.Run()
		fmt.Fprint(os.Stderr, doLineWrap)
	}

	if err != nil {
		log.Warnf("Prompt failed %v\n", err)
		return "", err
	}

	if result != "" && idx >= 0 && idx < len(filteredSessions) {
		sessionFile = filteredSessions[idx].Path
	}

	return
}
