package encrypttext

import (
	"fmt"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/worker"
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

	inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	if err != nil {
		return err
	}

	var iter iterator.Iterator
	_, input, err := flags.WithPipelineIterator(&flags.PipelineOptions{
		Name:     "text",
		Disabled: inputIterators.PipeOptions.Disabled,
		Required: true,
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
			encryptedText := v
			if cfg.SecureData.IsEncrypted(string(encryptedText)) != 1 {
				data, err := cfg.SecureData.EncryptString(string(encryptedText), n.passphrase)
				if err != nil {
					return nil, err
				}
				encryptedText = []byte(data)
			} else {
				log.Info("Text is already encrypted")
			}

			err = n.factory.WriteOutputWithoutPropertyGuess(encryptedText, cmdutil.OutputContext{})
		}
		return "", nil
	}, nil)
}
