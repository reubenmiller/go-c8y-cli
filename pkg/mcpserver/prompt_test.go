package mcpserver

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

func TestSwitchSessionPrompt(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("not offered without a session switcher", func(t *testing.T) {
		server := newPolicyTestServer("c8y")
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
		server := newPolicyTestServer("c8y")
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

func promptText(t *testing.T, res *mcp.GetPromptResult) string {
	t.Helper()
	if len(res.Messages) != 1 {
		t.Fatalf("unexpected messages: %+v", res.Messages)
	}
	return res.Messages[0].Content.(*mcp.TextContent).Text
}

func TestModePrompt(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newServer := func(ceiling config.SessionMode) (*Server, *mcp.ClientSession) {
		var logs []string
		s := newModeServer(ceiling, &logs)
		s.EnableModePrompt = true
		s.SetSessionKey("/a", SessionInfo{SessionName: "a", SessionMode: "prod"}, &Runner{Mode: config.SessionModeProduction})
		return s, connectApprovalClient(t, ctx, s, nil)
	}

	t.Run("not offered unless enabled", func(t *testing.T) {
		server := newPolicyTestServer("c8y")
		cs := connectApprovalClient(t, ctx, server, nil)
		if cs.InitializeResult().Capabilities.Prompts != nil {
			list, _ := cs.ListPrompts(ctx, nil)
			if list != nil && len(list.Prompts) != 0 {
				t.Fatalf("unexpected prompts: %+v", list.Prompts)
			}
		}
	})

	t.Run("changes and resets the mode", func(t *testing.T) {
		s, cs := newServer(config.SessionModeDev)
		res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: PromptMode, Arguments: map[string]string{"mode": "dev"}})
		if err != nil {
			t.Fatal(err)
		}
		if text := promptText(t, res); !strings.HasPrefix(text, "Mode is now dev (override; the session's own mode is prod)") {
			t.Errorf("unexpected message: %s", text)
		}
		if _, runner := s.Active(); runner.Mode != config.SessionModeDev {
			t.Errorf("mode not applied: %s", runner.Mode)
		}
		res, err = cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: PromptMode, Arguments: map[string]string{"mode": "reset"}})
		if err != nil {
			t.Fatal(err)
		}
		if text := promptText(t, res); !strings.HasPrefix(text, "Mode is now prod (the session's own mode)") {
			t.Errorf("unexpected message: %s", text)
		}
		if _, runner := s.Active(); runner.Mode != config.SessionModeProduction {
			t.Errorf("mode not reset: %s", runner.Mode)
		}
	})

	t.Run("rejects invalid and above ceiling values", func(t *testing.T) {
		s, cs := newServer(config.SessionModeQual)
		for _, value := range []string{"dev", "bogus", ""} {
			if _, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: PromptMode, Arguments: map[string]string{"mode": value}}); err == nil {
				t.Errorf("expected an error for mode %q", value)
			}
		}
		if _, runner := s.Active(); runner.Mode != config.SessionModeProduction {
			t.Errorf("mode must not change: %s", runner.Mode)
		}
	})

	t.Run("completes the modes under the ceiling", func(t *testing.T) {
		_, cs := newServer(config.SessionModeQual)
		res, err := cs.Complete(ctx, &mcp.CompleteParams{
			Ref:      &mcp.CompleteReference{Type: "ref/prompt", Name: PromptMode},
			Argument: mcp.CompleteParamsArgument{Name: "mode", Value: ""},
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Join(res.Completion.Values, ","); got != "prod,qual,reset" {
			t.Errorf("unexpected completions: %s", got)
		}
		res, err = cs.Complete(ctx, &mcp.CompleteParams{
			Ref:      &mcp.CompleteReference{Type: "ref/prompt", Name: PromptMode},
			Argument: mcp.CompleteParamsArgument{Name: "mode", Value: "q"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Join(res.Completion.Values, ","); got != "qual" {
			t.Errorf("unexpected completions: %s", got)
		}
	})
}

func TestSwitchSessionPromptArguments(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newServer := func(ceiling config.SessionMode) (*Server, *mcp.ClientSession, *[][2]string) {
		var logs []string
		s := newModeServer(ceiling, &logs)
		var calls [][2]string
		s.SwitchSession = func(ctx context.Context) (string, error) { return "picker", nil }
		s.SwitchSessionTo = func(ctx context.Context, session, mode string) (string, error) {
			calls = append(calls, [2]string{session, mode})
			return "switched", nil
		}
		s.SessionNames = func(ctx context.Context) []string { return []string{"alpha", "beta", "Alpine"} }
		return s, connectApprovalClient(t, ctx, s, nil), &calls
	}

	t.Run("without arguments opens the picker", func(t *testing.T) {
		_, cs, calls := newServer(config.SessionModeDev)
		res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: PromptSwitchSession})
		if err != nil {
			t.Fatal(err)
		}
		if promptText(t, res) != "picker" || len(*calls) != 0 {
			t.Errorf("unexpected result %q, calls %v", promptText(t, res), *calls)
		}
	})

	t.Run("with session and mode", func(t *testing.T) {
		_, cs, calls := newServer(config.SessionModeQual)
		res, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: PromptSwitchSession, Arguments: map[string]string{"session": "alpha", "mode": "qual"}})
		if err != nil {
			t.Fatal(err)
		}
		if promptText(t, res) != "switched" || len(*calls) != 1 || (*calls)[0] != [2]string{"alpha", "qual"} {
			t.Errorf("unexpected result %q, calls %v", promptText(t, res), *calls)
		}
	})

	t.Run("rejects a mode above the ceiling before switching", func(t *testing.T) {
		_, cs, calls := newServer(config.SessionModeQual)
		for _, mode := range []string{"dev", "bogus", "reset"} {
			if _, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: PromptSwitchSession, Arguments: map[string]string{"session": "alpha", "mode": mode}}); err == nil {
				t.Errorf("expected an error for mode %q", mode)
			}
		}
		if len(*calls) != 0 {
			t.Errorf("the session must not be switched: %v", *calls)
		}
	})

	t.Run("completes sessions and modes", func(t *testing.T) {
		_, cs, _ := newServer(config.SessionModeQual)
		res, err := cs.Complete(ctx, &mcp.CompleteParams{
			Ref:      &mcp.CompleteReference{Type: "ref/prompt", Name: PromptSwitchSession},
			Argument: mcp.CompleteParamsArgument{Name: "session", Value: "al"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Join(res.Completion.Values, ","); got != "alpha,Alpine" {
			t.Errorf("unexpected session completions: %s", got)
		}
		res, err = cs.Complete(ctx, &mcp.CompleteParams{
			Ref:      &mcp.CompleteReference{Type: "ref/prompt", Name: PromptSwitchSession},
			Argument: mcp.CompleteParamsArgument{Name: "mode", Value: ""},
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Join(res.Completion.Values, ","); got != "prod,qual" {
			t.Errorf("unexpected mode completions: %s", got)
		}
	})
}
