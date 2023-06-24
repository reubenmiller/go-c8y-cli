package delete

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

type CmdDelete struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
	age     string
	expired bool
}

func NewCmdDelete(f *cmdutil.Factory) *CmdDelete {
	ccmd := &CmdDelete{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete the cache",
		Long:  `Delete existing cache items either partially or cached items older than a time period`,
		Example: heredoc.Doc(`
			$ c8y cache delete
			Remove all of the existing cache. If there is no cache, then nothing will be done

			$ c8y cache delete --age -1h
			Remove cache items which are older than 1 hour
		`),
		RunE: ccmd.RunE,
	}

	cmd.Flags().StringVar(&ccmd.age, "age", "", "Age duration, i.e. 30s, 2m, 1h, 1d")
	cmd.Flags().BoolVar(&ccmd.expired, "expired", false, "Only removed expired cached items")

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdDelete) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	cs := n.factory.IOStreams.ColorScheme()

	cacheDir := cfg.CacheDir()

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		_, bErr := fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Nothing to delete. %s\n", cs.SuccessIconWithColor(cs.Green), cacheDir)
		return bErr
	}

	if n.age == "" && !n.expired {
		if err := os.RemoveAll(cacheDir); err != nil {
			return err
		}
		fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Deleted cache. %s\n", cs.SuccessIconWithColor(cs.Red), cacheDir)
		return nil
	}

	dryRun := cfg.DryRun()
	var ageDuration time.Duration
	if n.expired {
		ageDuration = cfg.CacheTTL()
	} else {
		ageDuration, err = flags.GetDuration(n.age, true, time.Hour)
		if err != nil {
			return err
		}
	}

	err = removeOldFiles(n.factory.IOStreams.ErrOut, cacheDir, ageDuration, dryRun)
	if err != nil {
		return err
	}
	if err := removeEmptyDirectories(n.factory.IOStreams.ErrOut, cacheDir, dryRun); err != nil {
		return err
	}
	fmt.Fprintf(n.factory.IOStreams.ErrOut, "%s Deleted cache %s\n", cs.SuccessIconWithColor(cs.Red), cacheDir)

	return err
}

func removeOldFiles(w io.Writer, root string, age time.Duration, dryRun bool) error {
	err := filepath.WalkDir(root, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(w, "prevent panic by handling failure accessing a path %q: %v\n", path, err)
			// return nil
			return err
		}
		if !info.IsDir() {
			file, err := os.Stat(path)
			if err != nil {
				return nil
			}
			if file.Mode().IsRegular() {
				// older than
				if time.Since(file.ModTime()) > age {
					if dryRun {
						fmt.Fprintf(w, "DRY: Deleting cache file: %s\n", path)
					} else {
						os.Remove(path)
					}
				}
			}
		}

		return nil
	})
	return err
}

func removeEmptyDirectories(w io.Writer, root string, dryRun bool) error {
	err := filepath.WalkDir(root, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			// ignore as we may have deleted the dir ourselves
			// fmt.Fprintf(w, "prevent dir panic by handling failure accessing a path %q: %v\n", path, err)
			return filepath.SkipDir
		}
		if info.IsDir() {
			files, err := os.ReadDir(path)

			if err != nil {
				return err
			}

			if len(files) == 0 {
				if dryRun {
					fmt.Fprintf(w, "DRY: Deleting empty directory: %s\n", path)
				} else {
					return os.Remove(path)
				}
			}
		}

		return nil
	})
	return err
}
