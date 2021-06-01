package get

import (
	"fmt"
	"path"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/spf13/cobra"
)

type CmdClone struct {
	*subcommand.SubCommand

	name     string
	fileType string
	modeType string

	factory *cmdutil.Factory
}

func NewCmdCloneSession(f *cmdutil.Factory) *CmdClone {
	ccmd := &CmdClone{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone session",
		Long: heredoc.Doc(`
			Clone/copy currently active session file. If no session file is activated, then an error will be returned.
		`),
		Example: heredoc.Doc(`
			$ c8y sessions clone --newName "myNewSession"
			Clone the existing activated
			
			$ c8y sessions clone --newName "customer-prod" --type "prod"
			Clone the existing session and rename it to "another" and also change the mode to production to disable all dangerous commands

			$ c8y sessions clone --newName "dev-otheruser" --fileType yaml
			Clone the existing session and rename it to "dev-otheruser" and change the session file format to yaml.
		`),
		RunE: ccmd.RunE,
	}

	cmd.Flags().StringVar(&ccmd.name, "newName", "", "Name of the new session file which will be created (required)")
	cmd.Flags().StringVar(&ccmd.fileType, "fileType", "json", "Session file type to save as. i.e. json, yaml, toml etc.")
	cmd.Flags().StringVar(&ccmd.modeType, "type", "", "Session type of the cloned session, i.e. dev, qual, prod")

	completion.WithOptions(cmd,
		completion.WithLazyRequired("newName"),
		completion.WithLazyRequired("fileType"),
		completion.WithValidateSet(
			"fileType",
			config.ConfigExtensions...,
		),
		completion.WithLazyRequired("type"),
		completion.WithValidateSet(
			"type",
			"prod\tProduction mode (read only)",
			"qual\tQA mode (delete disabled)",
			"dev\tDevelopment mode (no restrictions)",
		),
	)
	_ = cmd.MarkFlagRequired("newName")
	cmd.SilenceUsage = true

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdClone) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	if n.name == "" {
		return cmderrors.NewUserError("Missing required parameter. newName")
	}

	if n.modeType != "" {
		if err := config.SetMode(cfg.Persistent, n.modeType); err != nil {
			return cmderrors.NewUserError(err)
		}
	}

	dstPath := path.Join(cfg.GetSessionHomeDir(), n.name)

	if !strings.HasSuffix(dstPath, "."+n.fileType) {
		dstPath += "." + n.fileType
	}

	if err := cfg.Persistent.WriteConfigAs(dstPath); err != nil {
		return err
	}

	IO := n.factory.IOStreams
	if IO.IsStdoutTTY() {
		cs := IO.ColorScheme()
		fmt.Fprintf(IO.ErrOut, "%s Cloned session file to %s\n", cs.SuccessIconWithColor(cs.Green), dstPath)
	} else {
		fmt.Fprintln(IO.Out, dstPath)
	}

	return nil
}
