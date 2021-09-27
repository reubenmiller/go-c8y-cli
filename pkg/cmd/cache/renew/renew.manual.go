package renew

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type CmdRenew struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
}

func NewCmdRenew(f *cmdutil.Factory) *CmdRenew {
	ccmd := &CmdRenew{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "renew",
		Short: "Renew existing cache items",
		Long:  `Renew existing cache items by changing the modified dates of the files to "now" so that the cache has a full time-to-live period`,
		Example: heredoc.Doc(`
			$ c8y cache renew
			Renew all existing cache items. If there is no cache, then nothing will be done
		`),
		RunE: ccmd.RunE,
	}

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdRenew) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	cs := n.factory.IOStreams.ColorScheme()

	cacheDir := cfg.CacheDir()

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		_, bErr := fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Nothing to renew. %s\n", cs.SuccessIconWithColor(cs.Green), cacheDir)
		return bErr
	}

	currentTime := time.Now().Local()
	err = renewCacheFiles(n.factory.IOStreams.ErrOut, cacheDir, currentTime, cfg.DryRun())
	if err != nil {
		return err
	}
	fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Renewed cache %s\n", cs.SuccessIconWithColor(cs.Green), cacheDir)

	return err
}

func renewCacheFiles(w io.Writer, root string, currentTime time.Time, dryRun bool) error {
	err := filepath.WalkDir(root, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(w, "prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() {
			file, err := os.Stat(path)
			if err != nil {
				return nil
			}
			if file.Mode().IsRegular() {
				if dryRun {
					fmt.Fprintf(w, "DRY: Renewing cache file: %s\n", path)
				} else {
					_ = os.Chtimes(path, currentTime, currentTime)
				}
			}
		}

		return nil
	})
	return err
}
