package prompt

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/reubenmiller/go-c8y-cli/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

var (
	// ErrorPasswordMismatch error when the passwords entered by a user do not match
	ErrorPasswordMismatch = fmt.Errorf("passwords do not match")
	ErrorUserCancelled    = fmt.Errorf("user cancelled input")
)

// NewPromptWithPostValidate create a new prompt with a lazy validation function which is only
// run when the user hits enter.
func NewPromptWithPostValidate(p *promptui.Prompt, validate func(string) error) *PromptWithPostValidate {
	return &PromptWithPostValidate{
		prompt:       p,
		PostValidate: validate,
		MaxAttempts:  2,
	}
}

type PromptWithPostValidate struct {
	prompt       *promptui.Prompt
	Attempts     int
	MaxAttempts  int
	PostValidate func(string) error
}

func (p *PromptWithPostValidate) IsUserCancelled(err error) bool {
	return err != nil && strings.HasSuffix(err.Error(), "^C")
}

// Run prompts the user for input
func (p *PromptWithPostValidate) Run() (string, error) {
	var err error
	var input string
	p.Attempts = 1
	for {
		p.prompt.Templates.Prompt = "test me "
		if p.Attempts > 1 {
			p.prompt.Templates.Valid = fmt.Sprintf("{{ \"(Attempt %d of %d)\" | red }} {{ . | bold }}: ", p.Attempts, p.MaxAttempts)
		}
		input, err = p.prompt.Run()

		if p.IsUserCancelled(err) {
			err = ErrorUserCancelled
			break
		}
		err = p.PostValidate(input)
		if err == nil || p.Attempts >= p.MaxAttempts {
			break
		}
		p.Attempts++
	}
	return input, err
}

type Validator interface {
	Validate(string) error
}

// Validate function used to validate the user's input in a prompt ui
type Validate func(string) error

// Prompt used to provide various interactive prompts which can be used
// within the cli
type Prompt struct {
	Logger *logger.Logger
}

// NewPrompt returns a new Prompt which can be used to prompt the user for
// different information
func NewPrompt(l *logger.Logger) *Prompt {
	return &Prompt{
		Logger: l,
	}
}

// EncryptionPassphrase prompt for the encryption passphrase, and test the
// passphrase against the encrypted content to see if it is valid
func (p *Prompt) EncryptionPassphrase(encryptedData string, initPassphrase string, message string) (string, error) {
	var err error
	secure := encrypt.NewSecureData("{encrypted}")

	validate := func(input string) error {
		if secure.IsEncrypted(encryptedData) != 0 {
			_, err = secure.DecryptString(encryptedData, input)
			return err
		}
		return nil
	}
	prompt := promptui.Prompt{
		Stdin:       os.Stdin,
		Stdout:      os.Stderr,
		Default:     "",
		Mask:        ' ',
		HideEntered: true,
		Label:       "Session is encrypted, enter passphrase 🔒 [input is hidden]",
		Templates: &promptui.PromptTemplates{
			Valid: "{{ . | bold }}: ",
		},
	}
	promptWrapper := NewPromptWithPostValidate(&prompt, validate)

	// check if init passphrase is ok without prompting the user
	if err := validate(initPassphrase); err == nil {
		return initPassphrase, nil
	}
	if message != "" {
		p.ShowMessage(message)
	}
	return promptWrapper.Run()
}

func (p *Prompt) ShowMessage(m string) {
	faint := promptui.Styler(promptui.FGFaint)
	os.Stderr.WriteString(faint(m + "\n"))
}

// Password prompts the user for a password without confirmation
func (p *Prompt) Password(label string, message string) (string, error) {
	if label == "" {
		label = "Enter password"
	}
	validate := func(input string) error {
		if input == "" {
			return fmt.Errorf("Empty password")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Stdin:       os.Stdin,
		Stdout:      os.Stderr,
		Default:     "",
		Mask:        ' ',
		HideEntered: true,
		Label:       fmt.Sprintf("%s %s", label, "🔒"),
		Validate:    validate,
	}
	if message != "" {
		p.ShowMessage(message)
	}
	return prompt.Run()
}

// PasswordWithConfirm prompts the user for a password and confirms it by getting
// the user to type it in again
func (p *Prompt) PasswordWithConfirm(label string, message string) (string, error) {
	pass, err := p.Password(fmt.Sprintf("Enter %s", label), message)

	if err != nil {
		return "", err
	}

	passConfirm, err := p.Password(fmt.Sprintf("Confirm %s", label), "")

	if err != nil {
		return "", err
	}

	if pass != passConfirm {
		return "", ErrorPasswordMismatch
	}
	return pass, nil
}

// Username prompts for a username on the console
func (p *Prompt) Username(label string, defaultValue string) (string, error) {
	if label == "" {
		label = "Enter username/email"
	}
	validate := func(input string) error {
		if input == "" {
			return fmt.Errorf("Empty username")
		}
		return nil
	}
	pointer := promptui.DefaultCursor
	if defaultValue != "" {
		// use piped cursor so it does not block the first character of the default text
		pointer = promptui.PipeCursor
	}
	prompt := promptui.Prompt{
		Stdin:       os.Stdin,
		Stdout:      os.Stderr,
		Default:     defaultValue,
		Pointer:     pointer,
		HideEntered: true,
		Label:       label,
		Validate:    validate,
	}
	return prompt.Run()
}

// TOTPCode prompts for a TOTP code and validates using the given Cumulocity client
func (p *Prompt) TOTPCode(host, username string, code string, client *c8y.Client, initRequest string) (string, error) {
	os.Stderr.WriteString(fmt.Sprintf("Session details:\nHost=%s, username=%s\n", host, username))

	validateTOTP := func(input string) error {
		if len(strings.ReplaceAll(input, " ", "")) < 6 {
			return fmt.Errorf("Missing TFA code")
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(15000)*time.Millisecond)
		defer cancel()

		code = input

		if err := client.LoginUsingOAuth2(ctx, initRequest); err != nil {
			p.Logger.Errorf("OAuth2 failed. %s", err)
			return err
		}
		return nil
	}

	if err := validateTOTP(code); err == nil {
		return code, nil
	}

	prompt := promptui.Prompt{
		Stdin:       os.Stdin,
		Stdout:      os.Stderr,
		Default:     code,
		HideEntered: true,
		Label:       "Enter Two-Factor code",
		Validate:    validateTOTP,
	}

	return prompt.Run()
}
