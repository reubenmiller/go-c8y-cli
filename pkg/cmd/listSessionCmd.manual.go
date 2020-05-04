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
}

func newListSessionCmd() *listSessionCmd {
	ccmd := &listSessionCmd{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get a Cumulocity session",
		Long:  `Get a Cumulocity session`,
		Example: `

		`,
		RunE: ccmd.listSession,
	}

	cmd.SilenceUsage = true

	cmd.Flags().String("host", "", "Host. .e.g. test.cumulocity.com. (required)")
	cmd.Flags().String("tenant", "", "Tenant. (required)")
	cmd.Flags().String("username", "", "Username (without tenant). (required)")
	cmd.Flags().String("password", "", "Password. (required)")
	cmd.Flags().String("description", "", "Description about the session")
	cmd.Flags().String("name", "", "Name of the session")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
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
		// cmd.Printf("visited file or dir: %q\n", path)
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
		if strings.ToLower(os.Getenv(c8y.EnvVarLoggerHideSensitive)) != "true" {
			return fmt.Sprintf("%v", v)
		}
		return "*****"
	}

	funcMap["hideUser"] = func(v interface{}) string {
		msg := fmt.Sprintf("%v", v)
		if strings.ToLower(os.Getenv(c8y.EnvVarLoggerHideSensitive)) != "true" {
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

	searcher := func(input string, index int) bool {
		session := config.Sessions[index]

		name := strings.ToLower(fmt.Sprintf("#%02d %s %s %s %s",
			session.Index,
			filepath.Base(session.Path),
			session.Host,
			session.Tenant,
			session.Username,
		))
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Stdout:            cmd.OutOrStderr(),
		HideSelected:      false,
		IsVimMode:         false,
		StartInSearchMode: false,
		Label:             "Select a Cumulocity Session",
		Items:             config.Sessions,
		Templates:         templates,
		Size:              10,
		Searcher:          searcher,
	}

	idx, result, err := prompt.Run()

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

	if idx >= 0 && idx < len(config.Sessions) {
		fmt.Printf("%s", config.Sessions[idx].Path)
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
