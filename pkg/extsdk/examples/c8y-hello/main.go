// c8y-hello is an example c8y extension written in Go. Build it into the extension folder,
// then install the folder as a local extension:
//
//	go build -o bin/c8y-hello .
//	c8y extensions install .
//	c8y hello services list
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/reubenmiller/go-c8y-cli/pkg/extsdk"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

// Service is an example item. Its type matches the view in views/services.json
type Service struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

var services = []Service{
	{ID: "1", Name: "ingest", Type: "c8yHello_service", Status: "running"},
	{ID: "2", Name: "reports", Type: "c8yHello_service", Status: "stopped"},
	{ID: "3", Name: "alerts", Type: "c8yHello_service", Status: "running"},
}

func newServicesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "services", Short: "Manage services"}

	var status string
	list := &cobra.Command{
		Use:   "list",
		Short: "List services",
		RunE: func(cmd *cobra.Command, args []string) error {
			items := []Service{}
			for _, s := range services {
				if status == "" || s.Status == status {
					items = append(items, s)
				}
			}
			// Declared with "output: json" in extension.yaml, so c8y formats the output
			return extsdk.Print(items)
		},
	}
	list.Flags().StringVar(&status, "status", "", "Only include services with the given status")
	_ = list.RegisterFlagCompletionFunc("status", cobra.FixedCompletions([]string{"running", "stopped"}, cobra.ShellCompDirectiveNoFileComp))
	cmd.AddCommand(list)
	return cmd
}

func newDevicesCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "devices", Short: "Device commands"}
	cmd.AddCommand(&cobra.Command{
		Use:   "count",
		Short: "Count the devices in the tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := extsdk.FromEnv()
			if err != nil {
				return err
			}
			if ctx.DryRun() {
				// Only write json to stdout, as c8y processes it
				fmt.Fprintf(cmd.ErrOrStderr(), "dry run: would count the devices in %s\n", ctx.Host)
				return nil
			}
			client, err := ctx.Client()
			if err != nil {
				return err
			}
			col, _, err := client.Inventory.GetDevices(cmd.Context(), &c8y.PaginationOptions{
				PageSize:       1,
				WithTotalPages: true,
			})
			if err != nil {
				return extsdk.CheckError(err)
			}
			total := 0
			if col.Statistics != nil && col.Statistics.TotalPages != nil {
				total = *col.Statistics.TotalPages
			}
			return extsdk.Print(map[string]any{"count": total})
		},
	})
	return cmd
}

func newContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "context",
		Short: "Show the runtime context passed by c8y",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := extsdk.FromEnv()
			if err != nil {
				return err
			}
			return extsdk.Print(map[string]any{
				"name":    ctx.Name,
				"dir":     ctx.Dir,
				"bin":     ctx.Bin,
				"host":    ctx.Host,
				"tenant":  ctx.Tenant,
				"mode":    ctx.Mode,
				"dry":     ctx.DryRun(),
				"version": ctx.Version,
			})
		},
	}
}

func newExitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "exit <code>",
		Short: "Exit with the given exit code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "exiting with code %d\n", code)
			os.Exit(code)
			return nil
		},
	}
}

func main() {
	root := extsdk.NewRootCommand("c8y-hello")
	root.Short = "Example c8y extension written in Go"
	root.AddCommand(newServicesCmd(), newDevicesCmd(), newContextCmd(), newExitCmd())
	extsdk.Execute(root)
}
