package sessions

import (
	"context"
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/spf13/cobra"
)

type CmdEncryptText struct {
	passphrase string

	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdEncryptText(f *cmdutil.Factory) *CmdEncryptText {
	ccmd := &CmdEncryptText{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "encryptText",
		Short: "Encrypt text",
		Long:  `Encrypt text using the same encryption used for securely storing sensitive Cumulocity session information`,
		Example: `
Example 1: Encrypt the text "Hello World". You will be prompted for the passphrase to encrypt the data.

> c8y sessions encryptText --text "Hello World"
Enter passphrase 🔒: [input is hidden] 
{encrypted}ec5b837a03408ffb731307584eac40ac047989a002951e4b7139fa60189e504b6840bc027cece28b3f36717839d96af1c5dba8c850b9a9079846066ee1596cc8d26f4138f76ce3

Example 2: Encrypt the text "Hello World", the text will be encrypted using the given passphrase (without being prompted)

> c8y sessions encryptText --text "Hello World" --passphrase "so4methIng-7hat-Matters"
{encrypted}ec5b837a03408ffb731307584eac40ac047989a002951e4b7139fa60189e504b6840bc027cece28b3f36717839d96af1c5dba8c850b9a9079846066ee1596cc8d26f4138f76ce3
		`,
		RunE: ccmd.RunE,
	}

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	cmd.Flags().String("text", "", "Text to be encrypted. (required) (accepts pipeline)")
	cmd.Flags().StringVar(&ccmd.passphrase, "passphrase", "", "Passphrase to use for encrypting the text. Read from env C8Y_PASSPHRASE or prompted if missing")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdEncryptText) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	log, err := n.factory.Logger()
	if err != nil {
		return err
	}
	if n.passphrase == "" {
		inputPassphrase, err := cfg.PromptPassphrase()
		if err != nil {
			return err
		}
		n.passphrase = string(inputPassphrase)
	}

	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	if err := r.InputPipeline(flags.PipelineOptions{Name: "text", Required: true}); err != nil {
		return &flags.ParameterError{
			Name: "text",
			Err:  fmt.Errorf("missing required parameter or pipeline input. %w", flags.ErrParameterMissing),
		}
	}

	// No client: each item is encrypted locally.
	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		text := string(in.Input())
		return func(context.Context) any {
			if cfg.SecureData.IsEncrypted(text) == 1 {
				log.Info("Text is already encrypted")
				return c8ystream.FromDocs(jsondoc.New([]byte(text)))
			}
			data, err := cfg.SecureData.EncryptString(text, n.passphrase)
			if err != nil {
				return err
			}
			return c8ystream.FromDocs(jsondoc.New([]byte(data)))
		}, nil
	})
}
