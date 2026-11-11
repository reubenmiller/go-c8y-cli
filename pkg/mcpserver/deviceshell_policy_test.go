//go:build !windows

package mcpserver

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMarkDeviceCommands(t *testing.T) {
	c := newTestClassifier()
	cases := []struct {
		command string
		stdin   string
		want    bool
	}{
		{`operations create --device 1 --data "c8y_Command.text=df -h"`, "", true},
		{`operations create --device 1 --template "{c8y_Command:{text:'uptime'}}"`, "", true},
		{`api POST /devicecontrol/operations --data "deviceId=1,c8y_Command.text=ls"`, "", true},
		{`operations create --device 1 --template input.value`, `{"c8y_Command":{"text":"ls"}}`, true},
		{`operations create --device 1 --data "c8y_Restart={}"`, "", false},
		{`operations list --fragmentType c8y_Command`, "", false},
	}
	for _, tc := range cases {
		plan, err := c.Plan(tc.command)
		if err != nil {
			t.Fatalf("%s: %s", tc.command, err)
		}
		plan.MarkDeviceCommands(tc.stdin)
		if plan.HasDeviceShell() != tc.want {
			t.Errorf("%s: got device shell=%v, want %v", tc.command, plan.HasDeviceShell(), tc.want)
		}
	}
}

func callWrite(t *testing.T, ctx context.Context, cs *mcp.ClientSession, command string) *mcp.CallToolResult {
	t.Helper()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: ToolWrite, Arguments: map[string]any{"command": command}})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestDeviceShellPolicy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ssh := "remoteaccess connect ssh --device 1 -- uptime"
	operation := `operations create --device 1 --data "c8y_Command.text=uptime"`

	t.Run("disabled by default", func(t *testing.T) {
		server, marker := newApprovalTestServer(t)
		server.Policy.Approval = ApprovalNone
		cs := connectApprovalClient(t, ctx, server, nil)
		for _, command := range []string{ssh, operation} {
			res := callWrite(t, ctx, cs, command)
			if !res.IsError || !strings.Contains(resultText(res), "--allowDeviceShell") {
				t.Errorf("%s: expected device commands to be disabled, got %s", command, resultText(res))
			}
		}
		if _, err := os.Stat(marker); err == nil {
			t.Errorf("no command should have run")
		}
		// other operations are not affected
		if res := callWrite(t, ctx, cs, `operations create --device 1 --data "c8y_Restart={}"`); res.IsError {
			t.Errorf("other operations should be allowed: %s", resultText(res))
		}
	})

	t.Run("allowed", func(t *testing.T) {
		server, marker := newApprovalTestServer(t)
		server.Policy.Approval = ApprovalNone
		server.Policy.AllowDeviceShell = true
		cs := connectApprovalClient(t, ctx, server, nil)
		for _, command := range []string{ssh, operation} {
			if res := callWrite(t, ctx, cs, command); res.IsError {
				t.Errorf("%s: expected the command to run, got %s", command, resultText(res))
			}
		}
		content, _ := os.ReadFile(marker)
		if !strings.Contains(string(content), "--force --batch -- uptime") {
			t.Errorf("injected flags should be before --: %s", content)
		}
	})

	for _, confirm := range []bool{true, false} {
		name := "confirmed by the user"
		if !confirm {
			name = "not confirmed by the user"
		}
		t.Run(name, func(t *testing.T) {
			server, marker := newApprovalTestServer(t)
			server.Policy.Approval = ApprovalNone
			server.Policy.AllowDeviceShell = true
			asked := []string{}
			server.ConfirmDeviceShell = func(ctx context.Context, session SessionInfo, command string) error {
				asked = append(asked, session.Tenant+": "+command)
				if confirm {
					return nil
				}
				return errors.New("the user did not allow running the command on the device")
			}
			cs := connectApprovalClient(t, ctx, server, nil)
			res := callWrite(t, ctx, cs, ssh)
			_, statErr := os.Stat(marker)
			ran := statErr == nil
			if len(asked) != 1 || !strings.Contains(asked[0], "t123: c8y remoteaccess connect ssh --device 1 -- uptime") {
				t.Errorf("expected one confirmation for the command, got %q", asked)
			}
			if confirm != ran || confirm == res.IsError {
				t.Errorf("confirm=%v: ran=%v isError=%v %s", confirm, ran, res.IsError, resultText(res))
			}
			// commands which don't run on the device are not confirmed
			callWrite(t, ctx, cs, `devices create --name x`)
			if len(asked) != 1 {
				t.Errorf("only device commands should be confirmed")
			}
		})
	}
}

func TestDeviceShellConfirmer(t *testing.T) {
	session := SessionInfo{Host: "https://example.com", Tenant: "t123"}
	ctx := context.Background()

	t.Run("dialog", func(t *testing.T) {
		for _, allow := range []bool{true, false} {
			var message string
			c := &DeviceShellConfirmer{dialog: func(ctx context.Context, title, msg string, timeout time.Duration) (bool, error) {
				message = msg
				return allow, nil
			}}
			err := c.Confirm(ctx, session, "c8y remoteaccess connect ssh --device 1 -- reboot")
			if (err == nil) != allow {
				t.Errorf("allow=%v: got %v", allow, err)
			}
			if !strings.Contains(message, "t123") || !strings.Contains(message, "-- reboot") {
				t.Errorf("dialog should show the tenant and the command: %s", message)
			}
		}
	})

	noDialog := func(ctx context.Context, title, msg string, timeout time.Duration) (bool, error) {
		return false, errNoDialog
	}
	t.Run("no dialog", func(t *testing.T) {
		c := &DeviceShellConfirmer{dialog: noDialog}
		if err := c.Confirm(ctx, session, "c8y remoteaccess connect ssh --device 1 -- uptime"); err == nil {
			t.Errorf("command should be rejected when it can't be confirmed")
		}
	})
}
