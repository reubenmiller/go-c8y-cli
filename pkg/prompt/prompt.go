package prompt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cli/safeexec"
	"github.com/kballard/go-shellquote"
	"github.com/manifoldco/promptui"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

var (
	// ErrorPasswordMismatch error when the passwords entered by a user do not match
	ErrorPasswordMismatch = fmt.Errorf("passwords do not match")
	ErrorUserCancelled    = fmt.Errorf("user cancelled input")
)

// ErrNoRetry error where retries will not be attempted as the error is unrecoverable
var ErrNoRetry = errors.New("no retry")

// NewPromptWithPostValidate create a new prompt with a lazy validation function which is only
// run when the user hits enter.
func NewPromptWithPostValidate(p Prompter, validate func(string) error) *PromptWithPostValidate {
	return &PromptWithPostValidate{
		Prompter:     p,
		PostValidate: validate,
		MaxAttempts:  2,
	}
}

// Prompter interface for prompting the user for input
type Prompter interface {
	Prompt(attempt int, maxAttempts int) (string, error)
	Exists() (bool, error)
}

func NewPinEntryPromptWithPostValidate(r Prompter, validate func(string) error) *PromptWithPostValidate {
	return &PromptWithPostValidate{
		Prompter:     r,
		PostValidate: validate,
		MaxAttempts:  2,
	}
}

// CommandLinePrompter prompts the user for input from the command line
type CommandLinePrompter struct {
	prompt *promptui.Prompt
}

func (p *CommandLinePrompter) Prompt(attempt int, maxAttempts int) (string, error) {
	p.prompt.Templates.Prompt = "test me "
	if attempt > 1 {
		p.prompt.Templates.Valid = fmt.Sprintf("{{ \"(Attempt %d of %d)\" | red }} {{ . | bold }}: ", attempt, maxAttempts)
	}
	return p.prompt.Run()
}

func (p *CommandLinePrompter) Exists() (bool, error) {
	return true, nil
}

type ExternalPrompter struct {
	Command string
	Args    []string
}

func NewErrPrompt(err error, message string) *ErrPrompt {
	return &ErrPrompt{
		Inner:   err,
		Message: message,
	}
}

// ErrPrompt wraps an error and provides a custom message.
// It can be used to provide more context to an error without losing the original error.
type ErrPrompt struct {
	Inner   error
	Message string
}

func (e *ErrPrompt) Error() string {
	return e.Message
}

func (e *ErrPrompt) Unwrap() error {
	return e.Inner
}

// Prompt for input by executing an external command and parsing the output
func (p *ExternalPrompter) Prompt(attempt int, maxAttempts int) (string, error) {
	commandPath, err := safeexec.LookPath(p.Command)
	if err != nil {
		return "", NewErrPrompt(ErrNoRetry, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, commandPath, p.Args...)

	var errBuf, outBuf bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// don't passthrough additional values as this could be sensitive information
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}
	if err := cmd.Run(); err != nil {
		return "", NewErrPrompt(ErrNoRetry, fmt.Sprintf("pin entry failed. command=%s, args=%v, error=%s, stderr=%q", commandPath, p.Args, err, bytes.TrimSpace(errBuf.Bytes())))
	}

	return string(bytes.TrimSpace(outBuf.Bytes())), nil
}

func (p *ExternalPrompter) Exists() (bool, error) {
	_, err := safeexec.LookPath(p.Command)
	return err == nil, err
}

type PromptWithPostValidate struct {
	Prompter     Prompter
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
		input, err = p.Prompter.Prompt(p.Attempts, p.MaxAttempts)

		if p.IsUserCancelled(err) {
			err = ErrorUserCancelled
			break
		}
		if errors.Is(err, ErrNoRetry) {
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
	Logger         *logger.Logger
	ShowValueAfter bool
	PinEntry       string
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
func (p *Prompt) EncryptionPassphrase(encryptedData string, key string, initPassphrase string, message string) (string, error) {
	var err error
	secure := encrypt.NewSecureData("{encrypted}")

	validate := func(input string) error {
		if secure.IsEncrypted(encryptedData) != 0 {
			_, err = secure.DecryptString(encryptedData, input)
			return err
		}
		return nil
	}

	var prompter Prompter

	// Check for custom pin entry, but don't fail if it does not exist, just warn the user that it isn't being used for a good reason
	if p.PinEntry != "" {
		pinEntryCommand, err := shellquote.Split(p.PinEntry)
		if err != nil {
			pinEntryCommand = []string{p.PinEntry}
		}

		externalCommandArgs := []string{}
		if len(pinEntryCommand) == 0 {
			externalCommandArgs = append(externalCommandArgs, "get")
		} else {
			externalCommandArgs = append(externalCommandArgs, pinEntryCommand[1:]...)
		}
		externalCommandArgs = append(externalCommandArgs, key)
		pinEntryPrompter := &ExternalPrompter{
			Command: pinEntryCommand[0],
			Args:    externalCommandArgs,
		}
		if _, promptErr := pinEntryPrompter.Exists(); promptErr != nil {
			p.Logger.Warnf("user defined pin entry command (%s) does not exist. The default will be used instead. error=%s", p.PinEntry, promptErr)
		} else {
			p.Logger.Infof("Using external pin entry command. %s %s", pinEntryCommand[0], strings.Join(externalCommandArgs, " "))
			prompter = pinEntryPrompter
		}
	}

	// Fallback to the default cli prompter
	if prompter == nil {
		prompter = &CommandLinePrompter{
			prompt: &promptui.Prompt{
				Stdin:       os.Stdin,
				Stdout:      os.Stderr,
				Default:     "",
				Mask:        ' ',
				HideEntered: true,
				Label:       "Session is encrypted, enter passphrase 🔒 [input is hidden]",
				Templates: &promptui.PromptTemplates{
					Valid: "{{ . | bold }}: ",
				},
			},
		}
	}

	promptWrapper := NewPromptWithPostValidate(prompter, validate)

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
			return fmt.Errorf("password is required")
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
		if strings.TrimSpace(input) == "" {
			return fmt.Errorf("value is required")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Stdin:       os.Stdin,
		Stdout:      os.Stderr,
		Default:     defaultValue,
		HideEntered: false,
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
			return fmt.Errorf("missing TFA code")
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

func (p *Prompt) Input(label string, defaultValue string, required bool, hideEntered bool) (string, error) {
	if label == "" {
		label = "Enter value"
	}
	validate := func(input string) error {
		if required {
			if input == "" {
				return fmt.Errorf("value is required")
			}
		}
		return nil
	}
	prompt := promptui.Prompt{
		Stdin:       os.Stdin,
		Stdout:      os.Stderr,
		Default:     defaultValue,
		HideEntered: hideEntered,
		Label:       label,
		Validate:    validate,
	}
	return prompt.Run()
}
