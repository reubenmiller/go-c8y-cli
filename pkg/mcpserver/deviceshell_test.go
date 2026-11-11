package mcpserver

import (
	"reflect"
	"strings"
	"testing"
)

func TestDeviceShellClassification(t *testing.T) {
	c := newTestClassifier()
	cases := []struct {
		command string
		read    bool
		shell   bool
	}{
		{`remoteaccess connect ssh --device 1234 -- systemctl status`, false, true},
		{`remoteaccess connect ssh --device 1234 --user root -- "sh -c 'cat /etc/os-release | grep ID'"`, false, true},
		// --dry doesn't stop the command, so it must stay a write
		{`remoteaccess connect ssh --device 1234 --dry -- uptime`, false, true},
		// arguments after -- are passed to the device, not interpreted as c8y flags
		{`remoteaccess connect ssh --device 1234 -- rm -f /tmp/x --force`, false, true},
		{`remoteaccess connect ssh --device 1234 -- echo --dry --help`, false, true},
		{`remoteaccess connect ssh --help`, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.command, func(t *testing.T) {
			plan, err := c.Plan(tc.command)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if plan.ReadOnly() != tc.read || plan.HasDeviceShell() != tc.shell {
				t.Errorf("got read=%v shell=%v, want read=%v shell=%v", plan.ReadOnly(), plan.HasDeviceShell(), tc.read, tc.shell)
			}
		})
	}
}

func TestDeviceShellRejections(t *testing.T) {
	c := newTestClassifier()
	cases := []struct {
		command string
		errText string
	}{
		{`remoteaccess connect ssh --device 1234`, "Interactive sessions are not supported"},
		{`remoteaccess connect ssh --device 1234 --`, "Interactive sessions are not supported"},
		{`remoteaccess connect ssh --device 1234 --listen 0.0.0.0:2222 -- uptime`, "--listen"},
		{`remoteaccess connect ssh --device 1234 --listen=unix:///tmp/x.sock -- uptime`, "--listen"},
		{`remoteaccess connect ssh --device 1234 -L 1883 -- uptime`, "-L"},
		{`remoteaccess connect ssh --device 1234 -L1883:127.0.0.1:1883 -- uptime`, "-L"},
		{`remoteaccess connect ssh --device 1234 --port-forward=1883 -- uptime`, "--port-forward"},
		{`remoteaccess connect ssh --device 1234 --session other -- uptime`, "--session"},
		{`remoteaccess connect run --device 1234 -- sh -c id`, "denied"},
		{`remoteaccess server --device 1234`, "denied"},
		{`devices list -- --force`, "argument terminator"},
		{`remoteaccess connect ssh --device 1234 -- uptime && reboot`, "shell operator"},
	}
	for _, tc := range cases {
		t.Run(tc.command, func(t *testing.T) {
			_, err := c.Plan(tc.command)
			if err == nil {
				t.Fatalf("expected an error")
			}
			if !strings.Contains(err.Error(), tc.errText) {
				t.Errorf("expected error to contain %q, got %q", tc.errText, err)
			}
		})
	}

	// the device shell can be denied like any other command
	denied := newTestClassifier(func(p *Policy) { p.Deny = append(p.Deny, "remoteaccess connect ssh") })
	if _, err := denied.Plan(`remoteaccess connect ssh --device 1234 -- uptime`); err == nil || !strings.Contains(err.Error(), "denied") {
		t.Errorf("expected the command to be denied, got %v", err)
	}
}

func TestInjectFlags(t *testing.T) {
	cases := []struct {
		args []string
		want []string
	}{
		{[]string{"devices", "list"}, []string{"devices", "list", "--logMessage", "mcp", "--force"}},
		{
			[]string{"remoteaccess", "connect", "ssh", "--device", "1", "--", "uptime", "-p"},
			[]string{"remoteaccess", "connect", "ssh", "--device", "1", "--logMessage", "mcp", "--force", "--", "uptime", "-p"},
		},
	}
	for _, tc := range cases {
		if got := injectFlags(tc.args, "--logMessage", "mcp", "--force"); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("injectFlags(%v): got=%v, want=%v", tc.args, got, tc.want)
		}
	}
}
