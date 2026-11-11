package mcpserver

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSwitchSessionPrompt(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("not offered without a session switcher", func(t *testing.T) {
		server, _ := newApprovalTestServer(t)
		cs := connectApprovalClient(t, ctx, server, nil)
		if cs.InitializeResult().Capabilities.Prompts != nil {
			res, err := cs.ListPrompts(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Prompts) != 0 {
				t.Fatalf("expected no prompts, got %d", len(res.Prompts))
			}
		}
	})

	t.Run("returns the switcher's message", func(t *testing.T) {
		server, _ := newApprovalTestServer(t)
		calls := 0
		server.SwitchSession = func(ctx context.Context) (string, error) {
			calls++
			return "switching", nil
		}
		cs := connectApprovalClient(t, ctx, server, nil)
		list, err := cs.ListPrompts(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(list.Prompts) != 1 || list.Prompts[0].Name != PromptSwitchSession {
			t.Fatalf("unexpected prompts: %+v", list.Prompts)
		}
		res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: PromptSwitchSession})
		if err != nil {
			t.Fatal(err)
		}
		if calls != 1 {
			t.Fatalf("switcher called %d times", calls)
		}
		if len(res.Messages) != 1 || res.Messages[0].Role != "user" || res.Messages[0].Content.(*mcp.TextContent).Text != "switching" {
			t.Fatalf("unexpected messages: %+v", res.Messages)
		}
	})
}
