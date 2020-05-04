package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type newBinaryManagedObjectCmd struct {
	files []string
	*baseCmd
}

func newNewBinaryManagedObjectCmd() *newBinaryManagedObjectCmd {
	ccmd := &newBinaryManagedObjectCmd{}

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a binary managed object",
		Long:  `Upload a binary managed object`,
		Example: `
		Upload a binary managed object
		c8y inventory binary upload --file ./mybinary.zip

		c8y inventory binary upload --file ./test.zip --data "name=test,type=application/json"

		`,
		RunE: ccmd.newBinaryManagedObject,
	}

	cmd.Flags().StringArrayVarP(&ccmd.files, inventoryFlagFile, "f", []string{}, "input file to upload as a binary managed object")
	cmd.MarkFlagRequired(inventoryFlagFile)
	addDataFlag(cmd)

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

// newSite creates a new Hugo site and initializes a structured Hugo directory.
func (n *newBinaryManagedObjectCmd) newBinaryManagedObject(cmd *cobra.Command, args []string) error {
	if len(n.files) < 1 {
		return newUserError("at least one file needs to be provided")
	}

	data := getDataFlag(cmd)

	return n.doNewBinaryManagedObject(n.files, data)
}

func (n *newBinaryManagedObjectCmd) doNewBinaryManagedObject(files []string, data map[string]interface{}) error {
	wg := new(sync.WaitGroup)
	wg.Add(len(files))

	failures := make([]error, len(files))

	for i := range files {
		go func(index int) {

			// Set type if not already set
			if _, exists := data["type"]; !exists {
				// Guess the file type by reading the first 512 bytes
				// https://golangcode.com/get-the-content-type-of-file/
				f, err := os.Open(files[index])
				if err != nil {
					failures[index] = err
					wg.Done()
					return
				}
				defer f.Close()

				// Get the content
				contentType, err := GetFileContentType(f)
				if err != nil {
					failures[index] = err
					wg.Done()
					return
				}
				data["type"] = contentType
			}

			// Set name if blank
			if _, exists := data["name"]; !exists {
				data["name"] = filepath.Base(files[index])
			}

			_, resp, err := client.Inventory.CreateBinary(
				context.Background(),
				files[index],
				data,
			)

			if err != nil {
				failures[index] = err
				Logger.Debugf("file=%s, error`=%s", files[index], err)
			} else {
				fmt.Println(*resp.JSONData)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()

	var errorSummary error

	for _, err := range failures {
		if err != nil {
			if errorSummary == nil {
				errorSummary = errors.New("could not upload binary/s")
			}
			errorSummary = errors.WithStack(err)
		}
	}

	return errorSummary
}
