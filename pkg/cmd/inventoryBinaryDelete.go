package cmd

import (
	"context"
	"sync"

	"github.com/spf13/cobra"
)

type deleteBinaryManagedObjectCmd struct {
	*baseCmd
}

func newDeleteBinaryManagedObjectCmd() *deleteBinaryManagedObjectCmd {
	ccmd := &deleteBinaryManagedObjectCmd{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a binary managed object",
		Long:  `Delete a binary managed object`,
		Example: `
			Delete a binary managed object
			c8y inventory binary download --id 12345
		`,
		PreRunE: validateDeleteMode,
		RunE:    ccmd.deleteBinaryManagedObject,
	}

	// Flags
	addIDFlag(cmd)

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *deleteBinaryManagedObjectCmd) deleteBinaryManagedObject(cmd *cobra.Command, args []string) error {

	ids := GetIDs(cmd, args)

	return n.doDeleteBinaryManagedObject(ids)
}

func (n *deleteBinaryManagedObjectCmd) doDeleteBinaryManagedObject(ids []string) error {
	wg := new(sync.WaitGroup)
	wg.Add(len(ids))

	errorsCh := make(chan error, len(ids))

	for i := range ids {
		go func(index int) {
			_, err := client.Inventory.DeleteBinary(
				context.Background(),
				ids[index],
			)

			if err != nil {
				errorsCh <- err
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	close(errorsCh)
	return newErrorSummary("command failed", errorsCh)
}
