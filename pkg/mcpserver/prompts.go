package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

// Prompt arguments
const (
	argSession = "session"
	argMode    = "mode"
)

// promptResult returns a message for the conversation, which tells the agent what happened
func promptResult(description string, message string) *mcp.GetPromptResult {
	return &mcp.GetPromptResult{
		Description: description,
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: message}},
		},
	}
}

// addPrompts registers the prompts. Prompts are started by the user (in Claude Code a
// slash command), never by the agent, so they can change the session and the mode
func (s *Server) addPrompts(server *mcp.Server) {
	if s.SwitchSession != nil || s.SwitchSessionTo != nil {
		server.AddPrompt(&mcp.Prompt{
			Name:        PromptSwitchSession,
			Title:       "Switch the Cumulocity session",
			Description: "Switch the Cumulocity session used by the c8y tools. Without a session, a session picker opens in the browser",
			Arguments: []*mcp.PromptArgument{
				{Name: argSession, Title: "Session", Description: "Name of the session to switch to. Opens the picker if the session needs a passphrase or code"},
				{Name: argMode, Title: "Mode", Description: "Mode the agent can use in the session: prod, qual or dev (limited by the maximum mode of the server)"},
			},
		}, s.handleSwitchPrompt)
	}
	if s.EnableModePrompt {
		server.AddPrompt(&mcp.Prompt{
			Name:        PromptMode,
			Title:       "Change the mode of the Cumulocity session",
			Description: "Change what the agent can do in the current session: prod (read only), qual (create and update), dev (also delete) or reset (the session's own mode)",
			Arguments: []*mcp.PromptArgument{
				{Name: argMode, Title: "Mode", Description: "prod, qual, dev or reset", Required: true},
			},
		}, s.handleModePrompt)
	}
}

func (s *Server) handleSwitchPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	var session, modeName string
	if req != nil && req.Params != nil {
		session = strings.TrimSpace(req.Params.Arguments[argSession])
		modeName = strings.TrimSpace(req.Params.Arguments[argMode])
	}
	var message string
	var err error
	switch {
	case session == "" && modeName == "":
		if s.SwitchSession == nil {
			return nil, fmt.Errorf("opening the session picker is not supported")
		}
		message, err = s.SwitchSession(ctx)
	case s.SwitchSessionTo == nil:
		return nil, fmt.Errorf("switching to a named session is not supported")
	default:
		if modeName != "" {
			// reject an invalid mode before the session is switched
			mode, perr := ParseModeOrReset(modeName)
			if perr != nil {
				return nil, perr
			}
			if mode == config.SessionModeUnset {
				// reset has no meaning for a session which was not used before
				return nil, fmt.Errorf("invalid mode %q. Use prod, qual or dev", modeName)
			}
			if ceiling := s.Policy.Ceiling(); mode > ceiling {
				return nil, fmt.Errorf("mode %s is above the maximum mode of the MCP server (%s)", mode, ceiling)
			}
		}
		message, err = s.SwitchSessionTo(ctx, session, modeName)
	}
	if err != nil {
		return nil, err
	}
	return promptResult("Switch the Cumulocity session", message), nil
}

func (s *Server) handleModePrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	var name string
	if req != nil && req.Params != nil {
		name = req.Params.Arguments[argMode]
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("the mode is required: prod, qual, dev or reset")
	}
	mode, err := ParseModeOrReset(name)
	if err != nil {
		return nil, err
	}
	change, err := s.SetMode(mode, "prompt")
	if err != nil {
		return nil, fmt.Errorf("%w. Select a session first if none is selected, and use a mode up to %s", err, s.Policy.Ceiling())
	}
	return promptResult("Change the mode of the Cumulocity session", change.Message()), nil
}

// Message describes the change for the conversation, e.g. "Mode is now dev (override; the
// session's own mode is prod)"
func (c ModeChange) Message() string {
	var detail string
	if c.Source == ModeSourceOverride {
		if c.To == c.SessionMode {
			detail = "override, same as the session's own mode"
		} else {
			detail = fmt.Sprintf("override; the session's own mode is %s", c.SessionMode)
		}
	} else {
		detail = "the session's own mode"
	}
	msg := fmt.Sprintf("Mode is now %s (%s)", c.To, detail)
	if c.Session != "" {
		msg += fmt.Sprintf(" for session %s", c.Session)
	}
	return msg + ". Call c8y_session to see the allowed mode. Don't change any data as part of this."
}

// allowedModes returns the modes up to the ceiling
func (s *Server) allowedModes() []string {
	ceiling := s.Policy.Ceiling()
	var modes []string
	for _, mode := range []config.SessionMode{config.SessionModeProduction, config.SessionModeQual, config.SessionModeDev} {
		if mode <= ceiling {
			modes = append(modes, mode.String())
		}
	}
	return modes
}

// complete completes the arguments of the prompts: the session names, and the modes
// allowed under the ceiling
func (s *Server) complete(ctx context.Context, req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	values := []string{}
	if req != nil && req.Params != nil && req.Params.Ref != nil && req.Params.Ref.Type == "ref/prompt" {
		prefix := strings.ToLower(req.Params.Argument.Value)
		var candidates []string
		switch {
		case req.Params.Argument.Name == argMode && req.Params.Ref.Name == PromptMode:
			candidates = append(s.allowedModes(), "reset")
		case req.Params.Argument.Name == argMode && req.Params.Ref.Name == PromptSwitchSession:
			candidates = s.allowedModes()
		case req.Params.Argument.Name == argSession && req.Params.Ref.Name == PromptSwitchSession && s.SessionNames != nil:
			candidates = s.SessionNames(ctx)
		}
		for _, c := range candidates {
			if strings.HasPrefix(strings.ToLower(c), prefix) {
				values = append(values, c)
			}
		}
	}
	if len(values) > 100 {
		values = values[:100]
	}
	return &mcp.CompleteResult{Completion: mcp.CompletionResultDetails{Values: values, Total: len(values)}}, nil
}
