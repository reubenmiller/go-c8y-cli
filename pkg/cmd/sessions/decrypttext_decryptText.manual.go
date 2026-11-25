package sessions

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/encrypt"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/stream"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
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

	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	err = r.InputPipeline(flags.PipelineOptions{
		Name: "text",
		Formatter: func(b []byte) []byte {
			return bytes.TrimPrefix(bytes.TrimSpace(b), cfg.SecureData.Prefix)
		},
		InputFilter: func(b []byte) bool { return true },
		Required:    true,
		Mode:        stream.ModeText,
	})
	if err != nil {
		return &flags.ParameterError{
			Name: "text",
			Err:  fmt.Errorf("missing required parameter or pipeline input. %w", flags.ErrParameterMissing),
		}
	}

	// No client: each item is decrypted locally. The passphrase is only asked
	// for once there is something to decrypt.
	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		if n.passphrase == "" {
			inputPassphrase, err := cfg.PromptPassphrase()
			if err != nil {
				return nil, err
			}
			n.passphrase = string(inputPassphrase)
		}
		text := string(in.Input())
		return func(context.Context) any {
			data, err := cfg.SecureData.DecryptString(text, n.passphrase)
			// Text that is not encrypted is passed on as it is.
			if err != nil && !errors.Is(err, encrypt.ErrNotEncrypted) {
				return fmt.Errorf("failed to decrypt text. %w", err)
			}
			return c8ystream.FromDocs(jsondoc.New([]byte(data)))
		}, nil
	})
}
