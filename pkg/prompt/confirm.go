package prompt

import (
	"errors"
	"fmt"
	"os"
	"strings"

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

// Confirm prompts for a confirmation from the user
func Confirm(label string, target, defaultValue string, force bool) (ConfirmResult, error) {
	if force {
		return ConfirmYesToAll, nil
	}
	message := fmt.Sprintf("Confirm: ")

	if target != "" {
		message += fmt.Sprintf("%s on %s", label, target)
	} else {
		message += fmt.Sprintf("%s", label)
	}

	prompt := promptui.Prompt{
		Stdin:       os.Stdin,
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

	value, err := prompt.Run()

	// detect control-c
	if err != nil && strings.EqualFold("^C", err.Error()) {
		return ConfirmNoToAll, &ErrAbortAction{
			Message: message,
			Err:     ErrCancelAll,
		}
	}

	// set default
	if value == "" {
		value = defaultValue
	}

	// yes
	if strings.EqualFold(value, "y") {
		return ConfirmYes, nil
	}

	// yes to all
	if strings.EqualFold(value, "a") {
		return ConfirmYesToAll, nil
	}

	// no to all (hidden option)
	if strings.EqualFold(value, "l") {
		return ConfirmNoToAll, &ErrAbortAction{
			Message: message,
			Err:     ErrCancelAll,
		}
	}

	// everything else
	return ConfirmNo, &ErrAbortAction{
		Message: message,
	}
}
