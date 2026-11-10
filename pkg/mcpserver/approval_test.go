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

func TestApprovalTokens(t *testing.T) {
	tokens := &approvalTokens{}
	state := tokens.issue("devices create --name a")

	if err := tokens.use(state, "devices create --name b"); err == nil {
		t.Errorf("approval for another command should be rejected")
	}
	if err := tokens.use(state+"0", "devices create --name a"); err == nil {
		t.Errorf("tampered approval should be rejected")
	}
	// a rejected attempt with the right signature consumes nothing, so the right command still works
	if err := tokens.use(state, "devices create --name a"); err != nil {
		t.Errorf("approval should be accepted: %s", err)
	}
	if err := tokens.use(state, "devices create --name a"); err == nil {
		t.Errorf("approval should only be usable once")
	}
	if err := (&approvalTokens{}).use(state, "devices create --name a"); err == nil {
		t.Errorf("approval from another server should be rejected")
	}
}

// newApprovalTestServer returns a server whose commands are run by a fake c8y,
// which records that it was called
func newApprovalTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	dir := t.TempDir()
	marker := filepath.Join(dir, "ran")
	script := filepath.Join(dir, "c8y")
	// dry runs (used for the preview in the approval request) are not recorded
	content := "#!/bin/sh\ncase \"$*\" in *--dry*) echo dry-run >&2; exit 0;; esac\necho \"$@\" >> " + marker + "\necho ran\n"
	if err := os.WriteFile(script, []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}
	policy := DefaultPolicy()
	policy.Mode = config.SessionModeQual
	policy.Approval = ApprovalElicit
	return &Server{
		Version:    "test",
		Session:    SessionInfo{Host: "https://example.com", Tenant: "t123"},
		Policy:     policy,
		Classifier: &Classifier{Root: newTestRoot(), Policy: policy},
		Runner:     &Runner{Executable: script, Mode: config.SessionModeQual, BaseEnv: []string{"PATH=/usr/bin:/bin"}},
	}, marker
}

func connectApprovalClient(t *testing.T, ctx context.Context, server *Server, opts *mcp.ClientOptions) *mcp.ClientSession {
	t.Helper()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ss, err := server.New().Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, opts).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

func resultText(res *mcp.CallToolResult) string {
	if res == nil || len(res.Content) == 0 {
		return ""
	}
	return res.Content[0].(*mcp.TextContent).Text
}

func TestWriteApprovalMultiRoundTrip(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	for _, action := range []string{"accept", "decline", "cancel"} {
		t.Run(action, func(t *testing.T) {
			server, marker := newApprovalTestServer(t)
			var mu sync.Mutex
			messages := []string{}
			cs := connectApprovalClient(t, ctx, server, &mcp.ClientOptions{
				ElicitationHandler: func(ctx context.Context, req *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
					mu.Lock()
					messages = append(messages, req.Params.Message)
					mu.Unlock()
					return &mcp.ElicitResult{Action: action}, nil
				},
			})

			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: ToolWrite, Arguments: map[string]any{"command": "devices create --name test01"}})
			if err != nil {
				t.Fatal(err)
			}
			if len(messages) != 1 || !strings.Contains(messages[0], "c8y devices create --name test01") || !strings.Contains(messages[0], "t123") {
				t.Errorf("expected one approval request for the command, got %q", messages)
			}
			_, statErr := os.Stat(marker)
			ran := statErr == nil
			if action == "accept" {
				if res.IsError || !ran || !strings.Contains(resultText(res), "ran") {
					t.Errorf("approved command should run. isError=%v ran=%v text=%s", res.IsError, ran, resultText(res))
				}
			} else {
				if !res.IsError || ran || !strings.Contains(resultText(res), "did not approve") {
					t.Errorf("command should not run. isError=%v ran=%v text=%s", res.IsError, ran, resultText(res))
				}
			}
		})
	}
}

func TestWriteApprovalCanNotBeForged(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	server, marker := newApprovalTestServer(t)
	// Handle the round trips manually, like a client which answers on its own
	cs := connectApprovalClient(t, ctx, server, &mcp.ClientOptions{
		ElicitationHandler: func(ctx context.Context, req *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
			return &mcp.ElicitResult{Action: "accept"}, nil
		},
		MultiRoundTrip: &mcp.MultiRoundTripOptions{Disabled: true},
	})
	accept := mcp.InputResponseMap{approvalInputID: &mcp.ElicitResult{Action: "accept"}}
	call := func(command string, state string, responses mcp.InputResponseMap) *mcp.CallToolResult {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{
			Name:           ToolWrite,
			Arguments:      map[string]any{"command": command},
			InputResponses: responses,
			RequestState:   state,
		})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	first := call("devices create --name approved", "", nil)
	if !first.NeedsInput() || first.RequestState == "" {
		t.Fatalf("expected an approval request, got %+v", first)
	}

	cases := []struct {
		name    string
		command string
		state   string
	}{
		{"forged state", "devices create --name approved", "0123.abcd"},
		{"no state", "devices create --name approved", ""},
		{"other command", "devices delete --id 1", first.RequestState},
	}
	for _, tc := range cases {
		res := call(tc.command, tc.state, accept)
		if !res.IsError || !strings.Contains(resultText(res), "does not match") {
			t.Errorf("%s: expected the approval to be rejected, got %s", tc.name, resultText(res))
		}
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatalf("no command should have run")
	}

	// the real answer still works once, as failed attempts don't use it up...
	if res := call("devices create --name approved", first.RequestState, accept); res.IsError {
		t.Fatalf("approved command should run: %s", resultText(res))
	}
	// ...but it can't be replayed
	if res := call("devices create --name approved", first.RequestState, accept); !res.IsError {
		t.Errorf("approval should only be usable once")
	}
	content, _ := os.ReadFile(marker)
	if n := strings.Count(string(content), "devices create"); n != 1 {
		t.Errorf("command should have run once, ran %d times", n)
	}

	// switching the session invalidates approvals for the previous session
	second := call("devices create --name other", "", nil)
	server.SetSession(SessionInfo{Host: "https://other.example.com", Tenant: "t999"}, server.Runner)
	if res := call("devices create --name other", second.RequestState, accept); !res.IsError || !strings.Contains(resultText(res), "does not match") {
		t.Errorf("approval for the previous session should be rejected, got %s", resultText(res))
	}
}
