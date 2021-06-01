package prompt

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

// ErrAbortAction indicates that the user did not confirm an action
type ErrAbortAction struct {
	Message string
	Err     error
}

func (e ErrAbortAction) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("User did not confirm action. %s. %s", e.Message, e.Err)
	}
	return fmt.Sprintf("User did not confirm action. %s", e.Message)
}

// ErrCancelAll cancel all operations
var ErrCancelAll = errors.New("cancel all")

// ConfirmResult confirm result
type ConfirmResult int

const (
	// ConfirmYesToAll user has selected yes to current and all future operations
	ConfirmYesToAll ConfirmResult = iota + 1

	// ConfirmYes user has selected yes to current operation
	ConfirmYes

	// ConfirmNo user has selected no to current operation
	ConfirmNo

	// ConfirmNoToAll user has selected no to current and all future operations
	ConfirmNoToAll
)

func (c ConfirmResult) String() string {
	return [...]string{"", "a", "y", "n", "l"}[c]
}

func (c ConfirmResult) FromString(name string) ConfirmResult {
	values := map[string]ConfirmResult{
		"yestoall": ConfirmYesToAll,
		"a":        ConfirmYesToAll,
		"yes":      ConfirmYes,
		"y":        ConfirmYes,
		"notoall":  ConfirmNoToAll,
		"l":        ConfirmNoToAll,
		"no":       ConfirmNo,
		"n":        ConfirmNo,
	}
	if v, ok := values[strings.ToLower(name)]; ok {
		return v
	}
	return c
}

func PromptMultiLine(in *os.File, out *os.File, stderr io.Writer, prefix, message, defaultValue string) (string, error) {
	name := ""
	faint := color.New(color.Faint)
	prompt := &survey.Input{
		Default: defaultValue,
		Message: "Confirm " + prefix + "\n" + message + faint.Sprintf("\n[y] Yes  [a] Yes to All  [n] No  [l] No to All"),
		// Suggest: func(toComplete string) []string {
		// 	return []string{"Yes", "YesToAll", "No", "NoToAll"}
		// },
	}

	err := survey.AskOne(prompt, &name, survey.WithStdio(in, out, stderr))

	return name, err
}

// PromptSingleLine prompt on a single line
func PromptSingleLine(in *os.File, out *os.File, stderr io.Writer, prefix, message, defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Stdin:       in,
		Stdout:      os.Stderr,
		Default:     defaultValue,
		HideEntered: true,
		Label:       message,
		IsConfirm:   true,
		Validate: func(s string) error {
			if s == "" {
				return nil
			}
			s = strings.TrimSpace(strings.ToLower(s))
			if s == "y" || s == "n" || s == "a" || s == "l" {
				return nil
			}
			return fmt.Errorf("Invalid option")
		},
		Templates: &promptui.PromptTemplates{
			Valid:   fmt.Sprintf("%v {{ . }}? {{ \"[y] Yes [a] Yes to All [n] No. Default is '%s'\" | faint }}: ", promptui.IconGood, strings.ToLower(defaultValue)),
			Invalid: fmt.Sprintf("%v {{ . }}? {{ \"[y] Yes [a] Yes to All [n] No. Default is '%s'\" | faint }}: ", promptui.IconBad, strings.ToLower(defaultValue)),
			Confirm: fmt.Sprintf("%v {{ . }}? {{ \"[y] Yes [a] Yes to All [n] No. Default is '%s'\" | faint }}: ", promptui.IconInitial, strings.ToLower(defaultValue)),
		},
	}
	return prompt.Run()
}

// Confirm prompts for a confirmation from the user
func Confirm(prefix, label string, target, defaultValue string, force bool) (ConfirmResult, error) {
	if force {
		return ConfirmYesToAll, nil
	}
	message := ""

	if target != "" {
		message += fmt.Sprintf("%s on %s", label, target)
	} else {
		message += label
	}

	stdIn := os.Stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// stdin is handling Piped input, so we have to prompt on a different input
		if runtime.GOOS == "windows" {
			if file, err := os.Open("CON"); err == nil {
				stdIn = file
			}
		} else {
			tty, err := os.Open("/dev/tty")
			if err == nil {
				stdIn = tty
			}
		}
	}

	value, err := PromptMultiLine(stdIn, os.Stderr, os.Stderr, prefix, message, defaultValue)

	// detect control-c
	if err != nil {
		return ConfirmNoToAll, &ErrAbortAction{
			Message: message,
			Err:     ErrCancelAll,
		}
	}

	confirmValue := ConfirmNo.FromString(defaultValue)
	confirmValue = confirmValue.FromString(value)

	// yes or yes to all
	if confirmValue == ConfirmYes || confirmValue == ConfirmYesToAll {
		return confirmValue, nil
	}

	// no to all (hidden option)
	if confirmValue == ConfirmNoToAll {
		return ConfirmNoToAll, &ErrAbortAction{
			Message: message,
			Err:     ErrCancelAll,
		}
	}

	// everything else
	return confirmValue, &ErrAbortAction{
		Message: message,
	}
}
