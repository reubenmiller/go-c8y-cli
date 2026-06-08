package decrypttext

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/stream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
	"github.com/spf13/cobra"
)

type CmdDecryptText struct {
	passphrase string

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdDecryptText(f *cmdutil.Factory) *CmdDecryptText {
	ccmd := &CmdDecryptText{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "decryptText",
		Short: "Decrypt text",
		Long:  `Decrypt text based on the same encryption used to store sensitive data a cumulocity session`,
		Example: `
Example 1:
c8y sessions decryptText --text "{encrypted}asdfasdfasdfasdfasdf"

Encrypt the text "Hello World". You will be prompted for the passphrase to encrypt the data.

Example 2:
c8y sessions encryptText --text "Hello World" --passphrase "so4methIng-7hat-Matters"

Encrypt the text "Hello World", the text will be encrypted using the given passphrase (without being prompted)
		`,
		RunE: ccmd.RunE,
	}

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	cmd.Flags().String("text", "", "Encrypted text. (required) (accepts pipeline)")
	cmd.Flags().StringVar(&ccmd.passphrase, "passphrase", "", "Passphrase to use for encrypting the text. Read from env C8Y_PASSPHRASE or prompted if missing")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdDecryptText) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}

	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}
	_ = log

	var iter iterator.Iterator
	_, input, err := flags.WithPipelineIterator(&flags.PipelineOptions{
		Name:     "text",
		Disabled: inputIterators.PipeOptions.Disabled,
		Formatter: func(b []byte) []byte {
			return bytes.TrimPrefix(bytes.TrimSpace(b), cfg.SecureData.Prefix)
		},
		InputFilter: func(b []byte) bool {
			return true
		},
		Required: true,
		Mode:     stream.ModeText,
	})(cmd, inputIterators)

	if err != nil {
		return &flags.ParameterError{
			Name: "text",
			Err:  fmt.Errorf("missing required parameter or pipeline input. %w", flags.ErrParameterMissing),
		}
	}

	switch v := input.(type) {
	case iterator.Iterator:
		iter = v
	default:
		// use a single input iterator
		iter = iterator.NewRepeatIterator("", 1)
	}

	return n.factory.RunWithGenericWorkers(cmd, inputIterators, iter, func(j worker.Job) (any, error) {
		if v, ok := j.Value.([]byte); ok {

			if n.passphrase == "" {
				inputPassphrase, err := cfg.PromptPassphrase()
				if err != nil {
					return nil, err
				}
				n.passphrase = string(inputPassphrase)
			}

			data, err := cfg.SecureData.DecryptString(string(v), n.passphrase)

			// Don't treat the text (most likely being unencrypted as an error)
			if errors.Is(err, encrypt.ErrNotEncrypted) {
				err = nil
			}

			if err != nil {
				return "", fmt.Errorf("failed to decrypt text. %w", err)
			}

			err = n.factory.WriteOutputWithoutPropertyGuess([]byte(data), cmdutil.OutputContext{})
			return "", err
		}
		return "", nil
	}, nil)
}
