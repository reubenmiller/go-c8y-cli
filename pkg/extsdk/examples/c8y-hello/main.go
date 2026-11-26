// c8y-hello is an example c8y extension written in Go. Build it into the extension folder,
// then install the folder as a local extension:
//
//	go build -o bin/c8y-hello .
//	c8y extensions install .
//	c8y hello services list
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/reubenmiller/go-c8y-cli/pkg/extsdk"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
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
		Example: `  c8y hello services list --status stopped
  # List the stopped services`,
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
			res := client.SendRequest(cmd.Context(), api.RequestOptions{
				Method: http.MethodGet,
				Path:   "/inventory/managedObjects",
				Query: url.Values{
					"query":             {"$filter=has(c8y_IsDevice)"},
					"pageSize":          {"1"},
					"withTotalElements": {"true"},
				},
			})
			if res.Error != nil {
				return extsdk.CheckError(res.Error)
			}
			var page struct {
				Statistics struct {
					TotalElements int `json:"totalElements"`
				} `json:"statistics"`
			}
			if err := json.Unmarshal(res.Body(), &page); err != nil {
				return err
			}
			total := page.Statistics.TotalElements
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

// newSessionCmd is an example session provider. A real provider reads the sessions from a
// password manager or vault, and lets the user pick one
func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "session", Short: "Example session provider"}

	var name string
	selectCmd := &cobra.Command{
		Use:   "select",
		Short: "Select a session (declared as session provider in extension.yaml)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// The secret is declared in extension.yaml, so c8y passes it from the
			// environment or the pin entry. Talk to the user on stderr only
			if os.Getenv("HELLO_VAULT_TOKEN") == "" {
				return fmt.Errorf("HELLO_VAULT_TOKEN is not set")
			}
			fmt.Fprintf(os.Stderr, "Selected session %q from the example vault (active host: %q)\n", name, os.Getenv("C8Y_HOST"))

			authorized := true
			return extsdk.PrintSession(extsdk.Session{
				Host:       "https://" + name + ".example.com",
				Tenant:     "t100",
				Username:   "hello-user",
				Token:      "example-token",
				Authorized: &authorized,
				SessionUri: "hello://" + name,
			})
		},
	}
	selectCmd.Flags().StringVar(&name, "name", "demo", "Name of the session")
	// c8y passes its provider flags (e.g. --clear, -v) to providers, so accept them
	selectCmd.Flags().Bool("clear", false, "Clear the token")
	selectCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	selectCmd.Flags().Bool("debug", false, "Debug output")
	selectCmd.Flags().String("loginType", "", "Login type")
	cmd.AddCommand(selectCmd)
	return cmd
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
	root.AddCommand(newServicesCmd(), newDevicesCmd(), newContextCmd(), newSessionCmd(), newExitCmd())
	extsdk.Execute(root)
}
