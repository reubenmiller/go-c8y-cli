Function New-C8yApiGoLoopTemplate {
    [cmdletbinding()]
    Param()

    $Template = @"
package cmd

import (
    "context"
    "fmt"
    "strings"
    "sync"

    "github.com/reubenmiller/go-c8y/pkg/c8y"
    "github.com/spf13/cobra"
)

type ${Name}Cmd struct {
    *baseCmd
}

func new${NameCamel}Cmd() *${Name}Cmd {
	ccmd := &${Name}Cmd{}

	cmd := &cobra.Command{
		Use:   "$Use",
		Short: "$Description",
		Long:  "$DescriptionLong",
        Example: ``
        $($Examples -join "`n`n")
		``,
		RunE: ccmd.${Name},
	}

    $($CommandArgs.SetFlag -join "`n	")

	ccmd.baseCmd = newBaseCmd(cmd)

	return ccmd
}

func (n *${Name}Cmd) ${Name}(cmd *cobra.Command, args []string) error {

    $($CommandArgs.GetFlag -join "`n	")

	return n.do${NameCamel}($($CommandArgs.FunctionCall -join ", "))
}

func (n *${Name}Cmd) do${NameCamel}($($CommandArgs.FunctionDef -join ", ")) error {
	wg := new(sync.WaitGroup)
	wg.Add(len(ids))

	if len(filter) > 0 {
		// Print csv header
		fmt.Println(strings.Join(filter, ","))
	}

	errorsCh := make(chan error, len(ids))

	for i := range ids {
		go func(index int, filter []string) {
			_, resp, err := client.Inventory.GetManagedObject(
				context.Background(),
				ids[index],
				&c8y.ManagedObjectOptions{
					WithParents:       withParents,
					PaginationOptions: *c8y.NewPaginationOptions(1), // TODO: This should not be required as it is not supported by the api!
				},
			)

			if err != nil {
				errorsCh <- err
			} else {
				if len(filter) > 0 {
					selectedOutput := FilterJSON(*resp.JSON, filter)
					fmt.Println(strings.Join(selectedOutput, ","))
				} else {
					fmt.Println(*resp.JSONData)
				}
			}
			wg.Done()
		}(i, filter)
	}

	wg.Wait()
	close(errorsCh)
	return newErrorSummary("command failed", errorsCh)
}
"@

    $Template
}
