package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type downloadBinaryManagedObjectCmd struct {
	*baseCmd
}

func newDownloadBinaryManagedObjectCmd() *downloadBinaryManagedObjectCmd {
	ccmd := &downloadBinaryManagedObjectCmd{}

	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download a managed object",
		Long:  `Download a managed object`,
		Example: `
			Download a managed object
			c8y inventory download --id 12345
		`,
		RunE: ccmd.downloadBinaryManagedObject,
	}

	// Flags
	addIDFlag(cmd)
	cmd.Flags().StringP("outputDir", "o", "", "Output directory")
	cmd.Flags().StringP("outputFile", "f", "", "Output file")
	addInventoryOptions(cmd)

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *downloadBinaryManagedObjectCmd) downloadBinaryManagedObject(cmd *cobra.Command, args []string) error {

	ids := GetIDs(cmd, args)
	outputDir, _ := cmd.Flags().GetString("outputDir")
	outputFile, _ := cmd.Flags().GetString("outputFile")

	return n.doDownloadBinaryManagedObject(ids, outputDir, outputFile)
}

func (n *downloadBinaryManagedObjectCmd) doDownloadBinaryManagedObject(ids []string, outputDir string, outputFile string) error {
	wg := new(sync.WaitGroup)
	wg.Add(len(ids))

	errorsCh := make(chan error, len(ids))

	if outputDir == "" {
		if wd, err := os.Getwd(); err == nil {
			outputDir = wd
		} else {
			errorsCh <- newSystemError(err)
		}
	}

	for i := range ids {
		go func(index int) {
			tempfile, err := client.Inventory.DownloadBinary(
				context.Background(),
				ids[index],
			)

			if err != nil {
				errorsCh <- err
			} else {

				outputFile := path.Join(outputDir, filepath.Base(tempfile))
				if err := os.Rename(tempfile, outputFile); err != nil {
					errorsCh <- errors.Wrap(err, "failed to rename file")
				}
				if v, err := filepath.Abs(outputFile); err != nil {
					errorsCh <- errors.Wrap(err, "failed to resolve path")
				} else {
					fmt.Println(v)
				}
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	close(errorsCh)
	return newErrorSummary("command failed", errorsCh)
}
