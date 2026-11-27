//go:build !windows

package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

// newEscalationTestServer returns a server in a session with the given mode, whose fake c8y
// records the mode and arguments of every command which isn't a dry run. The dry run shows
// the HTTP method of the command (DELETE for delete commands, POST otherwise)
func newEscalationTestServer(t *testing.T, ceiling, sessionMode config.SessionMode, approval Approval) (*Server, string) {
	t.Helper()
	dir := t.TempDir()
	marker := filepath.Join(dir, "ran")
	script := filepath.Join(dir, "c8y")
	content := "#!/bin/sh\n" +
		"case \"$*\" in *--dry*)\n" +
		"  case \"$*\" in *delete*) echo '### DELETE /inventory/managedObjects/1' >&2;; *) echo '### POST /inventory/managedObjects' >&2;; esac\n" +
		"  exit 0;; esac\n" +
		"echo \"$C8Y_MODE $@\" >> " + marker + "\necho ran\n"
	if err := os.WriteFile(script, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}
	policy := DefaultPolicy()
	policy.Mode = ceiling
	policy.Approval = approval
	info := SessionInfo{Host: "https://example.com", Tenant: "t123", SessionName: "test.json", SessionMode: sessionMode.String()}
	server := &Server{
		Version:    "test",
		Policy:     policy,
		Classifier: &Classifier{Root: newTestRoot(), Policy: policy},
	}
	server.SetSession(info, &Runner{Executable: script, Mode: sessionMode, BaseEnv: []string{"PATH=/usr/bin:/bin"}})
	return server, marker
}

// answerClient answers every approval request with the given action and scope, and records the messages
func answerClient(t *testing.T, ctx context.Context, server *Server, action, scope string) (*mcp.ClientSession, *[]string) {
	t.Helper()
	var mu sync.Mutex
	messages := &[]string{}
	cs := connectApprovalClient(t, ctx, server, &mcp.ClientOptions{
		ElicitationHandler: func(ctx context.Context, req *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
			mu.Lock()
			*messages = append(*messages, req.Params.Message)
			mu.Unlock()
			result := &mcp.ElicitResult{Action: action}
			if scope != "" {
				result.Content = map[string]any{escalationScopeID: scope}
			}
			return result, nil
		},
	})
	return cs, messages
}

func callEscalated(t *testing.T, ctx context.Context, cs *mcp.ClientSession, command string) *mcp.CallToolResult {
	t.Helper()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: ToolWrite, Arguments: map[string]any{"command": command}})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func readMarker(marker string) string {
	content, _ := os.ReadFile(marker)
	return string(content)
}

func TestRequiredModeFromPreview(t *testing.T) {
	c := newTestClassifier()
	plan := func(command string) *Plan {
		t.Helper()
		p, err := c.Plan(command)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	cases := []struct {
		command string
		preview string
		want    config.SessionMode
	}{
		{"devices list", "", config.SessionModeProduction},
		{"devices create --name a", "", config.SessionModeQual},
		{"devices delete --id 1", "", config.SessionModeDev},
		{"devices create --name a", "### POST /inventory/managedObjects", config.SessionModeQual},
		{"devices create --name a", "### PUT /inventory/managedObjects/1", config.SessionModeQual},
		{"devices create --name a", "### DELETE /inventory/managedObjects/1", config.SessionModeDev},
		{"devices delete --id 1", "### POST /inventory/managedObjects", config.SessionModeQual},
	}
	for _, tc := range cases {
		if got := RequiredModeFromPreview(plan(tc.command), tc.preview); got != tc.want {
			t.Errorf("%q with preview %q: got %s, want %s", tc.command, tc.preview, got, tc.want)
		}
	}
}

func TestApprovalKeyIncludesRequiredMode(t *testing.T) {
	plan, err := newTestClassifier().Plan("devices create --name a")
	if err != nil {
		t.Fatal(err)
	}
	session := SessionInfo{Host: "https://example.com", Tenant: "t123"}
	in := CommandInput{Command: "devices create --name a"}
	qual := approvalKey(session, plan, in, config.SessionModeQual, config.SessionModeQual)
	dev := approvalKey(session, plan, in, config.SessionModeDev, config.SessionModeQual)
	if qual == dev {
		t.Fatalf("the key should depend on the required mode")
	}
	if prod := approvalKey(session, plan, in, config.SessionModeQual, config.SessionModeProduction); prod == qual {
		t.Fatalf("the key should depend on the mode of the session when asked")
	}
	tokens := &approvalTokens{}
	state := tokens.issue(qual)
	if err := tokens.use(state, dev); err == nil {
		t.Errorf("an approval for qual must not be usable for dev")
	}
	if err := tokens.use(state, qual); err != nil {
		t.Errorf("the approval should be usable for the mode it was given for: %s", err)
	}
}

func TestEscalationAllowOnce(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	server, marker := newEscalationTestServer(t, config.SessionModeDev, config.SessionModeProduction, ApprovalElicit)
	cs, messages := answerClient(t, ctx, server, "accept", escalationOnce)

	res := callEscalated(t, ctx, cs, "devices create --name a")
	if res.IsError || !strings.Contains(resultText(res), "ran") {
		t.Fatalf("command should run: %s", resultText(res))
	}
	if len(*messages) != 1 || !strings.Contains((*messages)[0], "This needs qual; the session is in prod") {
		t.Errorf("expected the escalation in the approval request, got %q", *messages)
	}
	if got := readMarker(marker); !strings.HasPrefix(got, "qual ") {
		t.Errorf("command should run in qual mode, got %q", got)
	}
	if mode := server.EffectiveMode("test.json", "prod"); mode != config.SessionModeProduction {
		t.Errorf("the session mode should not change, got %s", mode)
	}
	if _, ok := server.ModeOverride("test.json"); ok {
		t.Errorf("allow once should not set an override")
	}
	if _, runner := server.Active(); runner.Mode != config.SessionModeProduction {
		t.Errorf("the runner of the session should keep its mode, got %s", runner.Mode)
	}

	// it only applies to one command: the next one asks again
	callEscalated(t, ctx, cs, "devices create --name b")
	if len(*messages) != 2 {
		t.Errorf("the next command should need a new approval, got %d requests", len(*messages))
	}
}

func TestEscalationAllowForSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	server, marker := newEscalationTestServer(t, config.SessionModeDev, config.SessionModeQual, ApprovalElicit)
	cs, messages := answerClient(t, ctx, server, "accept", escalationSession)

	res := callEscalated(t, ctx, cs, "devices delete --id 1")
	if res.IsError {
		t.Fatalf("command should run: %s", resultText(res))
	}
	if len(*messages) != 1 || !strings.Contains((*messages)[0], "This needs dev; the session is in qual") {
		t.Errorf("expected the escalation in the approval request, got %q", *messages)
	}
	if got := readMarker(marker); !strings.HasPrefix(got, "dev ") {
		t.Errorf("command should run in dev mode, got %q", got)
	}
	if mode, ok := server.ModeOverride("test.json"); !ok || mode != config.SessionModeDev {
		t.Errorf("the override should be dev, got %s (set: %v)", mode, ok)
	}
	if _, runner := server.Active(); runner.Mode != config.SessionModeDev {
		t.Errorf("the runner of the session should be in dev, got %s", runner.Mode)
	}

	// the session is now in dev: no more questions about the mode
	callEscalated(t, ctx, cs, "devices delete --id 2")
	if len(*messages) != 2 || strings.Contains((*messages)[1], "This needs") {
		t.Errorf("no escalation should be needed any more, got %q", *messages)
	}
}

func TestEscalationDeclineRejects(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	server, marker := newEscalationTestServer(t, config.SessionModeDev, config.SessionModeProduction, ApprovalElicit)
	cs, _ := answerClient(t, ctx, server, "decline", "")
	res := callEscalated(t, ctx, cs, "devices create --name a")
	if !res.IsError || !strings.Contains(resultText(res), "did not approve") || readMarker(marker) != "" {
		t.Errorf("declined command should not run: %s", resultText(res))
	}
	if _, ok := server.ModeOverride("test.json"); ok {
		t.Errorf("declining should not set an override")
	}
}

func TestEscalationNeedsElicitingApproval(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	for _, approval := range []Approval{ApprovalNone, ApprovalClient} {
		t.Run(string(approval), func(t *testing.T) {
			server, marker := newEscalationTestServer(t, config.SessionModeDev, config.SessionModeProduction, approval)
			cs, messages := answerClient(t, ctx, server, "accept", escalationSession)
			res := callEscalated(t, ctx, cs, "devices create --name a")
			if !res.IsError || !strings.Contains(resultText(res), PromptMode) || readMarker(marker) != "" || len(*messages) != 0 {
				t.Errorf("command should be rejected with a hint to the mode prompt: %s", resultText(res))
			}
			if _, ok := server.ModeOverride("test.json"); ok {
				t.Errorf("no override should be set")
			}
		})
	}

	t.Run("no elicitation", func(t *testing.T) {
		server, marker := newEscalationTestServer(t, config.SessionModeDev, config.SessionModeProduction, ApprovalAuto)
		cs := connectApprovalClient(t, ctx, server, nil)
		res := callEscalated(t, ctx, cs, "devices create --name a")
		if !res.IsError || !strings.Contains(resultText(res), PromptMode) || readMarker(marker) != "" {
			t.Errorf("command should be rejected with a hint to the mode prompt: %s", resultText(res))
		}
	})
}

func TestEscalationRespectsCeiling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	server, marker := newEscalationTestServer(t, config.SessionModeQual, config.SessionModeQual, ApprovalElicit)
	cs, messages := answerClient(t, ctx, server, "accept", escalationSession)
	res := callEscalated(t, ctx, cs, "devices delete --id 1")
	if !res.IsError || !strings.Contains(resultText(res), "maximum mode") || readMarker(marker) != "" {
		t.Errorf("command above the ceiling should be rejected: %s", resultText(res))
	}
	if len(*messages) != 0 {
		t.Errorf("the user should not be asked for something the ceiling forbids, got %q", *messages)
	}
	if _, ok := server.ModeOverride("test.json"); ok {
		t.Errorf("no override should be set")
	}
}

func TestNoEscalationKeepsApproval(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	server, marker := newEscalationTestServer(t, config.SessionModeDev, config.SessionModeQual, ApprovalElicit)
	cs, messages := answerClient(t, ctx, server, "accept", "")
	res := callEscalated(t, ctx, cs, "devices create --name a")
	if res.IsError || len(*messages) != 1 || strings.Contains((*messages)[0], "This needs") {
		t.Fatalf("expected a plain approval: %s %q", resultText(res), *messages)
	}
	if got := readMarker(marker); !strings.HasPrefix(got, "qual ") {
		t.Errorf("command should run in the mode of the session, got %q", got)
	}
}
