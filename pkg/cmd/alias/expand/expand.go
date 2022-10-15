package expand

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/cli/safeexec"
	"github.com/google/shlex"
)

// ExpandAlias processes argv to see if it should be rewritten according to a user's aliases. The
// second return value indicates whether the alias should be executed in a new shell process instead
// of running gh itself.
func ExpandAlias(aliases map[string]string, args []string, findShFunc func() (string, error)) (expanded []string, isShell bool, err error) {
	if len(args) < 2 {
		// the command is lacking a subcommand
		return
	}

	aliasName := args[1]

	// Check if command is being called with completion, if so
	// alias name is offset by 1
	isCompletion := false
	completionCmd := ""
	if len(args) > 2 {
		if strings.HasPrefix(args[1], "__complete") {
			completionCmd = args[1]
			aliasName = args[2]
			isCompletion = true
		}
	}

	expanded = args[1:]

	normalizedAliases := map[string]string{}
	for k, v := range aliases {
		normalizedAliases[strings.ToLower(k)] = v
	}

	expansion, ok := normalizedAliases[strings.ToLower(aliasName)]
	if !ok {
		return
	}
	expandedAliasWithVariables := expansion

	if strings.HasPrefix(expansion, "!") {
		isShell = true
		if findShFunc == nil {
			findShFunc = findSh
		}
		shPath, shErr := findShFunc()
		if shErr != nil {
			err = shErr
			return
		}

		expanded = []string{shPath, "-c", expansion[1:]}

		if len(args[2:]) > 0 {
			expanded = append(expanded, "--")
			expanded = append(expanded, args[2:]...)
		}

		return
	}

	extraArgs := []string{}
	isOption := regexp.MustCompile(`^--?[a-zA-Z][a-zA-Z0-9\-]*(=.*)?$`)
	foundOption := false

	// Positional arguments should be included first (otherwise we have to parse the command using the proper command definition)
	// to find out which are flags expecting options and which are boolean flags.
	for i, a := range args[2:] {
		if !strings.Contains(expansion, "$") {
			extraArgs = append(extraArgs, a)
		} else {
			// Check if user provided a positional argument or not
			// Once an option is found, assume everything afterwards is also options or option values
			if foundOption || isOption.MatchString(a) {
				foundOption = true
				extraArgs = append(extraArgs, a)
			} else {
				expansion = strings.ReplaceAll(expansion, fmt.Sprintf("$%d", i+1), a)
			}
		}
	}
	lingeringRE := regexp.MustCompile(`\$\d`)
	if lingeringRE.MatchString(expansion) {

		missingPositionals := lingeringRE.FindAllString(expansion, -1)
		expectedPositionals := lingeringRE.FindAllString(expandedAliasWithVariables, -1)

		err = fmt.Errorf(
			"not enough arguments for alias, expected %d, got %d\n\n  Expanded alias: %s %s\n\n  Tip: Aliases expect the positional arguments to be given before other flags",
			len(expectedPositionals),
			len(expectedPositionals)-len(missingPositionals),
			expansion,
			strings.Join(extraArgs, " "),
		)
		return
	}

	var newArgs []string
	newArgs, err = shlex.Split(expansion)
	if err != nil {
		return
	}

	expanded = append(newArgs, extraArgs...)

	// Prepend completion command if the user called for completion
	if isCompletion {
		expanded = append([]string{completionCmd}, expanded...)
	}
	return
}

func findSh() (string, error) {
	shPath, err := safeexec.LookPath("sh")
	if err == nil {
		return shPath, nil
	}

	if runtime.GOOS == "windows" {
		winNotFoundErr := errors.New("unable to locate sh to execute the shell alias with. The sh.exe interpreter is typically distributed with Git for Windows.")
		// We can try and find a sh executable in a Git for Windows install
		gitPath, err := safeexec.LookPath("git")
		if err != nil {
			return "", winNotFoundErr
		}

		shPath = filepath.Join(filepath.Dir(gitPath), "..", "bin", "sh.exe")
		_, err = os.Stat(shPath)
		if err != nil {
			return "", winNotFoundErr
		}

		return shPath, nil
	}

	return "", errors.New("unable to locate sh to execute shell alias with")
}
