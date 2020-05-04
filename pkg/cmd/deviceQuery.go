package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

type queryDeviceCmd struct {
	deviceName []string
	deviceType string

	*baseCmd
}

func newQueryDeviceCmd() *queryDeviceCmd {
	ccmd := &queryDeviceCmd{}

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find a collection of devices",
		Long:  `Find a collection of devices`,
		Example: `
			c8y devices find --name myDevice
		`,
		RunE: ccmd.queryDevice,
	}

	// Flags
	cmd.Flags().StringSliceVarP(&ccmd.deviceName, "name", "n", []string{"*"}, "name (accepts wildcards")
	cmd.Flags().StringVar(&ccmd.deviceType, "type", "", "type")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *queryDeviceCmd) queryDevice(cmd *cobra.Command, args []string) error {
	return n.doQueryDevice(n.deviceName, n.deviceType)
}

func (n *queryDeviceCmd) doQueryDevice(deviceName []string, deviceType string) error {
	wg := new(sync.WaitGroup)
	wg.Add(len(deviceName))

	errorsCh := make(chan error, len(deviceName))

	for i := range deviceName {
		go func(index int) {
			query := fmt.Sprintf("has(c8y_IsDevice) and name eq '%s'", deviceName[index])

			if deviceType != "" {
				query = fmt.Sprintf("%s and type eq '%s'", query, deviceType)
			}

			_, resp, err := client.Inventory.GetManagedObjects(
				context.Background(),
				&c8y.ManagedObjectOptions{
					Query: query,
					PaginationOptions: c8y.PaginationOptions{
						PageSize:       globalFlagPageSize,
						WithTotalPages: globalFlagWithTotalPages,
					},
				},
			)
			if err != nil {
				errorsCh <- errors.Wrap(err, "command failed")
			} else {
				fmt.Println(*resp.JSONData)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	close(errorsCh)
	return newErrorSummary("command failed", errorsCh)
}
