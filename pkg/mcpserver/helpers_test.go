package mcpserver

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

// newPolicyTestServer returns a server in qualified mode which runs commands
// with the given executable (which does not need to exist if no command is run).
// It has no unix-only code so tests using it also run on Windows.
func newPolicyTestServer(executable string) *Server {
	policy := DefaultPolicy()
	policy.Mode = config.SessionModeQual
	policy.Approval = ApprovalElicit
	return &Server{
		Version:    "test",
		Session:    SessionInfo{Host: "https://example.com", Tenant: "t123"},
		Policy:     policy,
		Classifier: &Classifier{Root: newTestRoot(), Policy: policy},
		Runner:     &Runner{Executable: executable, Mode: config.SessionModeQual, BaseEnv: []string{"PATH=/usr/bin:/bin"}},
	}
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
