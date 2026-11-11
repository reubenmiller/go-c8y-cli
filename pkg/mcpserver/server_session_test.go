package mcpserver

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

func TestSetSessionChangesActiveSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server := newTestServer()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ss, err := server.New().Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	call := func() string {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: ToolSession})
		if err != nil {
			t.Fatal(err)
		}
		return res.Content[0].(*mcp.TextContent).Text
	}

	if text := call(); !strings.Contains(text, `"tenant": "t123"`) || !strings.Contains(text, `"mode": "prod"`) {
		t.Fatalf("unexpected session: %s", text)
	}

	// switching the session applies to the existing connection
	server.SetSession(
		SessionInfo{Host: "https://other.example.com", Tenant: "t999", SessionMode: "qual"},
		&Runner{Executable: "/bin/false", Mode: config.SessionModeQual},
	)
	text := call()
	if !strings.Contains(text, `"tenant": "t999"`) || !strings.Contains(text, `"host": "https://other.example.com"`) || !strings.Contains(text, `"mode": "qual"`) {
		t.Errorf("session was not switched: %s", text)
	}

	// the agent is not told a fixed tenant, as it can change
	if strings.Contains(server.Instructions(), "example.com") {
		t.Errorf("instructions should not include the tenant: %s", server.Instructions())
	}
}
