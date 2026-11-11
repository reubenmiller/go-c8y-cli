package mcpserver

import (
	"os"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/spf13/cobra"
)

func newTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "c8y"}
	root.PersistentFlags().String("session", "", "")
	root.PersistentFlags().BoolP("force", "f", false, "")
	root.PersistentFlags().Bool("dry", false, "")
	root.PersistentFlags().StringArray("select", nil, "")

	noop := func(cmd *cobra.Command, args []string) error { return nil }
	group := func(name string, children ...string) *cobra.Command {
		cmd := &cobra.Command{Use: name}
		for _, child := range children {
			cmd.AddCommand(&cobra.Command{Use: child, RunE: noop})
		}
		root.AddCommand(cmd)
		return cmd
	}
	devices := group("devices", "list", "get", "create", "delete", "update", "setRequiredAvailability")
	devices.AddCommand(&cobra.Command{Use: "listen", RunE: noop})
	group("inventory", "get", "find", "count")
	group("sessions", "list", "get")
	group("microservices", "list", "getBootstrapUser")
	group("operations", "wait", "list", "create")
	group("template", "execute")

	remoteAccess := &cobra.Command{Use: "remoteaccess"}
	connect := &cobra.Command{Use: "connect"}
	for _, name := range []string{"ssh", "run"} {
		sub := &cobra.Command{Use: name, RunE: noop}
		sub.Flags().StringSlice("device", nil, "")
		sub.Flags().String("user", "", "")
		sub.Flags().String("listen", "127.0.0.1:0", "")
		sub.Flags().StringP("port-forward", "L", "", "")
		connect.AddCommand(sub)
	}
	remoteAccess.AddCommand(connect)
	remoteAccess.AddCommand(&cobra.Command{Use: "server", RunE: noop})
	root.AddCommand(remoteAccess)

	api := &cobra.Command{Use: "api", RunE: noop}
	api.Flags().String("method", "GET", "")
	api.Flags().String("url", "", "")
	api.Flags().String("data", "", "")
	root.AddCommand(api)
	return root
}

func newTestClassifier(mutate ...func(p *Policy)) *Classifier {
	policy := DefaultPolicy()
	for _, m := range mutate {
		m(&policy)
	}
	return &Classifier{Root: newTestRoot(), Policy: policy}
}

func TestPlanReadWriteClassification(t *testing.T) {
	cases := []struct {
		command string
		read    bool
	}{
		{"devices list", true},
		{"c8y devices list --select id,name", true},
		{"devices get --id 1", true},
		{"inventory find --query 'name eq x*'", true},
		{"inventory count", true},
		{"operations wait --id 1", true},
		{"template execute --template '{}'", true},
		{"devices create --name test", false},
		{"devices delete --id 1", false},
		{"devices setRequiredAvailability --id 1", false},
		{"devices listen", false},
		{"devices delete --id 1 --dry", true},
		{"devices delete --id 1 --dry=false", false},
		{"devices delete --help", true},
		{"api /inventory/managedObjects", true},
		{"api GET /inventory/managedObjects", true},
		{"api --url /inventory/managedObjects", true},
		{"api --method POST --url /inventory/managedObjects", false},
		{"api --method=DELETE /inventory/managedObjects/1", false},
		{"api DELETE /inventory/managedObjects/1", false},
		{"api POST", false},
	}

	c := newTestClassifier()
	for _, tc := range cases {
		t.Run(tc.command, func(t *testing.T) {
			plan, err := c.Plan(tc.command)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if plan.ReadOnly() != tc.read {
				t.Errorf("expected read=%v, got %v", tc.read, plan.ReadOnly())
			}
		})
	}
}

func TestPlanRejections(t *testing.T) {
	cases := []struct {
		command string
		errText string
	}{
		{"sessions list", "denied"},
		{"microservices getBootstrapUser --id 1", "denied"},
		{"devices list --session other", "--session"},
		{"devices list --session=other", "--session"},
		{"devices delete --id 1 --force", "--force"},
		{"devices delete --id 1 -f", "-f"},
		{"devices list --sessionMode dev", "--sessionMode"},
		{"devices list && devices delete --id 1", "shell operator"},
		{"devices list > out.json", "shell operator"},
		{"devices list | jq .id", "unknown c8y command"},
		{"foo bar", "unknown c8y command"},
		{"--select id", "must start with a c8y command"},
		{"api --url https://example.com/x", "absolute urls"},
		{"api GET https://example.com/x", "absolute urls"},
		{"devices list | ", "empty command"},
		{"devices list -- --force", "terminator"},
	}

	c := newTestClassifier()
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
}

func TestRelativeTimeValuesAreNotTreatedAsFlags(t *testing.T) {
	c := newTestClassifier()
	if _, err := c.Plan("devices list --dateFrom -1d --dateTo -10min"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestPipeline(t *testing.T) {
	c := newTestClassifier()
	plan, err := c.Plan("c8y devices list --select id | c8y devices delete")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(plan.Segments))
	}
	if !plan.Segments[0].Read || plan.Segments[1].Read {
		t.Errorf("unexpected classification: %+v", plan.Segments)
	}
	if plan.ReadOnly() {
		t.Errorf("pipeline with a delete should not be read-only")
	}
	if got := plan.String(); got != "c8y devices list --select id | c8y devices delete" {
		t.Errorf("unexpected string: %s", got)
	}
}

func TestAllowDenyLists(t *testing.T) {
	c := newTestClassifier(func(p *Policy) {
		p.Allow = []string{"devices", "inventory get"}
		p.Deny = append(p.Deny, "* delete")
	})
	allowed := []string{"devices list", "devices create --name x", "inventory get --id 1"}
	for _, cmd := range allowed {
		if _, err := c.Plan(cmd); err != nil {
			t.Errorf("%q should be allowed: %s", cmd, err)
		}
	}
	denied := []string{"devices delete --id 1", "inventory find", "api /inventory"}
	for _, cmd := range denied {
		if _, err := c.Plan(cmd); err == nil {
			t.Errorf("%q should be denied", cmd)
		}
	}
}

func TestMatchCommand(t *testing.T) {
	cases := []struct {
		pattern string
		path    string
		match   bool
	}{
		{"devices", "devices", true},
		{"devices", "devices list", true},
		{"devices", "devicegroups list", false},
		{"devices *", "devices list", true},
		{"devices *", "devices children list", true},
		{"* delete", "devices delete", true},
		{"* delete", "devices list", false},
		{"* list", "devices children list", false},
		{"*", "devices list", true},
		{"Devices LIST", "devices list", true},
	}
	for _, tc := range cases {
		if got := MatchCommand(tc.pattern, tc.path); got != tc.match {
			t.Errorf("MatchCommand(%q, %q) = %v, expected %v", tc.pattern, tc.path, got, tc.match)
		}
	}
}

func TestCapMode(t *testing.T) {
	cases := []struct {
		policy, session, expected config.SessionMode
	}{
		{config.SessionModeDev, config.SessionModeProduction, config.SessionModeProduction},
		{config.SessionModeQual, config.SessionModeDev, config.SessionModeQual},
		{config.SessionModeDev, config.SessionModeCI, config.SessionModeDev},
		{config.SessionModeDev, config.SessionModeUnset, config.SessionModeProduction},
		{config.SessionModeProduction, config.SessionModeDev, config.SessionModeProduction},
		// unset policy mode: the session's mode is used
		{config.SessionModeUnset, config.SessionModeDev, config.SessionModeDev},
		{config.SessionModeUnset, config.SessionModeQual, config.SessionModeQual},
		{config.SessionModeUnset, config.SessionModeUnset, config.SessionModeProduction},
		{config.SessionModeUnset, config.SessionModeCI, config.SessionModeDev},
	}
	for _, tc := range cases {
		if got := CapMode(tc.policy, tc.session); got != tc.expected {
			t.Errorf("CapMode(%s, %s) = %s, expected %s", tc.policy, tc.session, got, tc.expected)
		}
	}
}

func TestRunnerEnv(t *testing.T) {
	r := &Runner{
		SessionFile: "/tmp/session.json",
		BaseEnv: []string{
			"PATH=/usr/bin",
			"CI=true",
			"BUILD_ID=1",
			"C8Y_MODE=dev",
			"C8Y_SETTINGS_CI=true",
			"C8Y_SETTINGS_MODE_ENABLEDELETE=true",
			"C8Y_SETTINGS_DEFAULTS_FORCE=true",
			"C8Y_SETTINGS_DEFAULTS_DRY=true",
			"C8Y_PASSPHRASE=secret",
		},
	}
	env := strings.Join(r.Env(config.SessionModeProduction, nil), "\n")
	for _, unwanted := range []string{"CI=true", "BUILD_ID=", "C8Y_MODE=dev", "C8Y_SETTINGS_CI=true", "ENABLEDELETE", "DEFAULTS_FORCE", "DEFAULTS_DRY"} {
		if strings.Contains(env, unwanted) {
			t.Errorf("env should not contain %q", unwanted)
		}
	}
	for _, wanted := range []string{"PATH=/usr/bin", "C8Y_PASSPHRASE=secret", "C8Y_MODE=prod", "C8Y_SETTINGS_CI=false", "C8Y_SESSION=/tmp/session.json", "C8Y_SETTINGS_LOGGER_HIDESENSITIVE=true"} {
		if !strings.Contains(env, wanted) {
			t.Errorf("env should contain %q", wanted)
		}
	}
}

func TestRunnerEnvDeniedPaths(t *testing.T) {
	r := &Runner{
		SessionFile: "/tmp/session.json",
		DeniedPaths: []string{"/home/user/.cumulocity", "/home/user/.ssh"},
		BaseEnv:     []string{"PATH=/usr/bin", EnvDeniedPaths + "=/nothing"},
	}
	env := r.Env(config.SessionModeProduction, nil)
	want := EnvDeniedPaths + "=/home/user/.cumulocity" + string(os.PathListSeparator) + "/home/user/.ssh"
	found := 0
	for _, item := range env {
		if strings.HasPrefix(item, EnvDeniedPaths+"=") {
			found++
			if item != want {
				t.Errorf("got %q, want %q", item, want)
			}
		}
	}
	if found != 1 {
		t.Errorf("expected exactly one %s entry (the inherited one is replaced), got %d", EnvDeniedPaths, found)
	}
}

func TestLimitedWriter(t *testing.T) {
	buf := &strings.Builder{}
	w := &limitedWriter{w: buf, n: 5}
	n, err := w.Write([]byte("hello world"))
	if err != nil || n != 11 {
		t.Fatalf("unexpected write result: %d, %v", n, err)
	}
	if buf.String() != "hello" || !w.truncated {
		t.Errorf("unexpected output %q, truncated=%v", buf.String(), w.truncated)
	}
}
