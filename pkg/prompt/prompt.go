package prompt

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/reubenmiller/go-c8y-cli/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
)

var (
	// ErrorPasswordMismatch error when the passwords entered by a user do not match
	ErrorPasswordMismatch = fmt.Errorf("passwords do not match")
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

// Run prompts the user for input
func (p *PromptWithPostValidate) Run() (string, error) {
	var err error
	var input string
	p.Attempts = 1
	for {
		if p.Attempts > 1 {
			p.prompt.Templates.Valid = fmt.Sprintf("{{ \"(Attempt %d of %d)\" | red }} {{ . | bold }}: ", p.Attempts, p.MaxAttempts)
		}
		input, err = p.prompt.Run()
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
func (p *Prompt) EncryptionPassphrase(encryptedData string) (string, error) {
	var err error
	secure := encrypt.NewSecureData("{encrypted}")

	validate := func(input string) error {
		_, err = secure.DecryptString(encryptedData, input)
		return err
	}
	prompt := promptui.Prompt{
		Stdin:   os.Stdin,
		Stdout:  os.Stderr,
		Default: "",
		Mask:    '*',
		Label:   "Enter passphrase 🔒",
		Templates: &promptui.PromptTemplates{
			Valid: "{{ . | bold }}: ",
		},
	}
	promptWrapper := NewPromptWithPostValidate(&prompt, validate)
	return promptWrapper.Run()
}

// Password prompts the user for a password without confirmation
func (p *Prompt) Password(passwordType string) (string, error) {
	if passwordType == "" {
		passwordType = "c8y"
	}
	validate := func(input string) error {
		if input == "" {
			return fmt.Errorf("Empty password")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Stdin:    os.Stdin,
		Stdout:   os.Stderr,
		Default:  "",
		Mask:     '*',
		Label:    fmt.Sprintf("Enter %s password 🔒", passwordType),
		Validate: validate,
	}
	return prompt.Run()
}

// PasswordWithConfirm prompts the user for a password and confirms it by getting
// the user to type it in again
func (p *Prompt) PasswordWithConfirm(passwordType string) (string, error) {

	pass, err := p.Password(passwordType)

	if err != nil {
		return "", err
	}

	passConfirm, err := p.Password(passwordType)

	if err != nil {
		return "", err
	}

	if pass != passConfirm {
		return "", ErrorPasswordMismatch
	}
	return pass, nil
}
