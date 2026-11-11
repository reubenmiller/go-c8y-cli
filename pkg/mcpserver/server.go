package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

const (
	ToolRead    = "c8y_read"
	ToolWrite   = "c8y_write"
	ToolSession = "c8y_session"

	// PromptSwitchSession prompt to let the user switch the session (stdio mode)
	PromptSwitchSession = "switch_session"
)

// SessionInfo non-secret information about the session used by the server
type SessionInfo struct {
	Host        string `json:"host"`
	Tenant      string `json:"tenant,omitempty"`
	Username    string `json:"username,omitempty"`
	SessionName string `json:"session,omitempty"`
	SessionMode string `json:"sessionMode"`
}

// Server MCP server which exposes go-c8y-cli to an agent
type Server struct {
	Version string

	// Session, Runner the session used by the commands. Use SetSession to change
	// them once the server is running
	Session SessionInfo
	Runner  *Runner

	Policy     Policy
	Classifier *Classifier

	// OnCall optional callback which is called after each tool call, e.g. to log the
	// agent's activity
	OnCall func(tool string, command string, outcome string)

	// Gate optional check which is run before each tool call, e.g. to wait until the
	// user has selected a session
	Gate func(ctx context.Context, ss *mcp.ServerSession) error

	// SwitchSession optional function which lets the user select another session. It is
	// offered as an MCP prompt (in Claude Code a slash command), which only the user can
	// run, not the agent. It returns a message for the conversation, and should not wait
	// for the user's selection: tool calls wait for it instead (see Gate)
	SwitchSession func(ctx context.Context) (string, error)

	// ConfirmDeviceShell optional confirmation by the user of commands which run a shell
	// command on a device, in addition to the normal approval. Unlike the approval, the
	// MCP client can't answer it
	ConfirmDeviceShell func(ctx context.Context, session SessionInfo, command string) error

	mu        sync.RWMutex
	approvals approvalTokens
}

// SetSession changes the session used by the commands, e.g. when the user switches
// the session. Tool calls which have already started keep
// using the previous session, so a command always runs against the tenant it was
// approved for.
func (s *Server) SetSession(info SessionInfo, runner *Runner) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Session = info
	s.Runner = runner
}

// Active returns the session and runner used for new tool calls
func (s *Server) Active() (SessionInfo, *Runner) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Session, s.Runner
}

func (s *Server) logCall(tool, command, outcome string) {
	if s.OnCall != nil {
		s.OnCall(tool, command, outcome)
	}
}

// CommandInput input for the read/write tools
type CommandInput struct {
	Command        string `json:"command" jsonschema:"c8y command to run without a shell, e.g. 'devices list --name \"sensor*\" --select id,name'. The leading 'c8y' is optional. Multiple c8y commands can be joined with ' | ' to pipe the output of one into the next"`
	Stdin          string `json:"stdin,omitempty" jsonschema:"Optional data piped to the first command, e.g. newline separated ids or json lines"`
	TimeoutSeconds int    `json:"timeoutSeconds,omitempty" jsonschema:"Optional timeout in seconds"`
}

func boolPtr(v bool) *bool { return &v }

// Instructions returns the instructions sent to the MCP client
func (s *Server) Instructions() string {
	return heredoc.Docf(`
		Run go-c8y-cli (c8y) commands against a Cumulocity tenant.
		- Use %s for commands that only read data (get, list, find, count, --help, --dry). Use %s for anything that creates, updates or deletes data.
		- Call %s first to see the tenant, the allowed mode and the command policy. The user can switch to another session at any time, so call it again if the tenant matters, e.g. before changing data.
		- Discover commands and flags with "<command> --help" or "<command> --examples" instead of guessing.
		- Keep output small: use --select, --filter, --pageSize and --view off. Output is json by default.
		- Pipe c8y commands with " | " (no shell is used, so jq, grep, &&, ; and redirects are not supported).
		- Preview writes with --dry before running them with %s.
		- To troubleshoot a device, you can run a shell command on it with %s, if the user enabled it (see deviceShell in %s). Both ways need a mode which allows creating data:
		  - via ssh, if Cumulocity remote access is configured for the device: remoteaccess connect ssh --device <id> -- <command>
		    Quote commands which use shell syntax, e.g. -- "sh -c 'ps | grep agent'". Interactive sessions are not supported.
		    Only ssh keys are used (no passwords). Without --user, the user's default ssh user is tried first, then root.
		  - via a c8y_Command operation, if the device supports it: operations create --device <id> --description "<why>" --data 'c8y_Command.text=<command>' | operations wait --status SUCCESSFUL --status FAILED --duration 2m
		    The output is in c8y_Command.result.
	`, ToolRead, ToolWrite, ToolSession, ToolWrite, ToolWrite, ToolSession)
}

// New creates the MCP server
func (s *Server) New() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "go-c8y-cli",
		Title:   "Cumulocity (go-c8y-cli)",
		Version: s.Version,
	}, &mcp.ServerOptions{
		Instructions: s.Instructions(),
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:  ToolRead,
		Title: "Run a read-only c8y command",
		Description: heredoc.Doc(`
			Run a go-c8y-cli command which only reads data from Cumulocity, e.g. "devices list", "alarms list --device 1234",
			"inventory get --id 1234", "api /tenant/currentTenant", or any command with --help, --examples or --dry.
			The command runs in read-only mode, so write requests are rejected.
		`),
		Annotations: &mcp.ToolAnnotations{
			Title:           "Read Cumulocity data",
			ReadOnlyHint:    true,
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(true),
		},
	}, s.handleRead)

	mcp.AddTool(server, &mcp.Tool{
		Name:  ToolWrite,
		Title: "Run a c8y command which changes data",
		Description: heredoc.Docf(`
			Run a go-c8y-cli command which creates, updates or deletes data in Cumulocity, e.g. "devices create --name test01",
			or runs a shell command on a device, e.g. "remoteaccess connect ssh --device 1234 -- systemctl status".
			The user may be asked to approve the command. The allowed actions depend on the mode set by the user (see %s);
			prod: read only, qual: create and update, dev: create, update and delete.
			Preview the requests first with --dry using %s.
		`, ToolSession, ToolRead),
		Annotations: &mcp.ToolAnnotations{
			Title:           "Change Cumulocity data",
			ReadOnlyHint:    false,
			DestructiveHint: boolPtr(true),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(true),
		},
	}, s.handleWrite)

	mcp.AddTool(server, &mcp.Tool{
		Name:        ToolSession,
		Title:       "Show the Cumulocity session and policy",
		Description: "Show the Cumulocity tenant, user, allowed mode and command policy used by this MCP server. The user can switch the session at any time. No secrets are returned.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Show session",
			ReadOnlyHint:    true,
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
		},
	}, s.handleSession)

	if s.SwitchSession != nil {
		server.AddPrompt(&mcp.Prompt{
			Name:        PromptSwitchSession,
			Title:       "Switch the Cumulocity session",
			Description: "Open a session picker in the browser to change the Cumulocity session used by the c8y tools",
		}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			message, err := s.SwitchSession(ctx)
			if err != nil {
				return nil, err
			}
			return &mcp.GetPromptResult{
				Description: "Switch the Cumulocity session",
				Messages: []*mcp.PromptMessage{
					{Role: "user", Content: &mcp.TextContent{Text: message}},
				},
			}, nil
		})
	}

	return server
}

func errorResult(format string, a ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, a...)}},
	}
}

func (s *Server) timeout(seconds int) time.Duration {
	timeout := s.Policy.Timeout
	if seconds > 0 {
		timeout = time.Duration(seconds) * time.Second
	}
	if s.Policy.MaxTimeout > 0 && timeout > s.Policy.MaxTimeout {
		timeout = s.Policy.MaxTimeout
	}
	return timeout
}

// checkGate returns an error result if the client is not allowed to call tools (yet)
func (s *Server) checkGate(ctx context.Context, req *mcp.CallToolRequest, tool string, command string) *mcp.CallToolResult {
	if s.Gate == nil {
		return nil
	}
	var ss *mcp.ServerSession
	if req != nil {
		ss = req.Session
	}
	if err := s.Gate(ctx, ss); err != nil {
		s.logCall(tool, command, "client not approved")
		return errorResult("%s", err)
	}
	return nil
}

func (s *Server) handleSession(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	if res := s.checkGate(ctx, req, ToolSession, ""); res != nil {
		return res, nil, nil
	}
	session, runner := s.Active()
	info := map[string]any{
		"session":  session,
		"mode":     runner.Mode.String(),
		"modeInfo": runner.Mode.Description(),
		"approval": string(s.Policy.Approval),
		"deviceShell": map[string]any{
			"enabled":           s.Policy.AllowDeviceShell,
			"confirmedByServer": s.ConfirmDeviceShell != nil,
		},
		"policy": map[string]any{
			"allow":        s.Policy.Allow,
			"deny":         s.Policy.Deny,
			"readCommands": s.Policy.ReadCommands,
			"deniedFlags":  DeniedFlags,
		},
	}
	b, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}

func (s *Server) handleRead(ctx context.Context, req *mcp.CallToolRequest, in CommandInput) (*mcp.CallToolResult, any, error) {
	if res := s.checkGate(ctx, req, ToolRead, in.Command); res != nil {
		return res, nil, nil
	}
	plan, err := s.Classifier.Plan(in.Command)
	if err != nil {
		s.logCall(ToolRead, in.Command, "rejected: "+err.Error())
		return errorResult("%s", err), nil, nil
	}
	if !plan.ReadOnly() {
		writes := []string{}
		for _, seg := range plan.Segments {
			if !seg.Read {
				writes = append(writes, seg.CommandPath)
			}
		}
		s.logCall(ToolRead, in.Command, "rejected: command may change data")
		if plan.HasDeviceShell() {
			return errorResult("%q runs a command on the device, which may change it, so it can't be run with %s. Use %s instead", strings.Join(writes, ", "), ToolRead, ToolWrite), nil, nil
		}
		return errorResult("%q may change data so it can't be run with %s. Use %s instead (or add --dry to preview the requests)", strings.Join(writes, ", "), ToolRead, ToolWrite), nil, nil
	}
	_, runner := s.Active()
	return s.run(ctx, runner, plan, in, ToolRead)
}

func (s *Server) handleWrite(ctx context.Context, req *mcp.CallToolRequest, in CommandInput) (*mcp.CallToolResult, any, error) {
	if res := s.checkGate(ctx, req, ToolWrite, in.Command); res != nil {
		return res, nil, nil
	}
	plan, err := s.Classifier.Plan(in.Command)
	if err != nil {
		s.logCall(ToolWrite, in.Command, "rejected: "+err.Error())
		return errorResult("%s", err), nil, nil
	}
	// Use the same session for the approval and the command, even if the user switches the session meanwhile
	session, runner := s.Active()
	plan.MarkDeviceCommands(in.Stdin)
	if plan.HasDeviceShell() && !s.Policy.AllowDeviceShell {
		s.logCall(ToolWrite, in.Command, "rejected: device commands are disabled")
		return errorResult("running commands on devices is disabled. The user can enable it by starting the MCP server with --allowDeviceShell"), nil, nil
	}
	if !plan.ReadOnly() {
		if runner.Mode == config.SessionModeProduction {
			s.logCall(ToolWrite, in.Command, "rejected: read-only mode")
			return errorResult("the MCP server is running in read-only mode (prod) so write commands are disabled. Ask the user to change the mode if this is needed"), nil, nil
		}
		inputRequest, err := s.approve(ctx, req, session, runner, plan, in)
		if err != nil {
			s.logCall(ToolWrite, in.Command, "not approved")
			return errorResult("%s", err), nil, nil
		}
		if inputRequest != nil {
			// the client asks the user and calls the tool again with the answer
			s.logCall(ToolWrite, in.Command, "waiting for approval")
			return inputRequest, nil, nil
		}
		if plan.HasDeviceShell() && s.ConfirmDeviceShell != nil {
			s.logCall(ToolWrite, in.Command, "waiting for confirmation in a dialog")
			if err := s.ConfirmDeviceShell(ctx, session, plan.String()); err != nil {
				s.logCall(ToolWrite, in.Command, "not confirmed")
				return errorResult("%s", err), nil, nil
			}
		}
	}
	return s.run(ctx, runner, plan, in, ToolWrite)
}

func (s *Server) run(ctx context.Context, runner *Runner, plan *Plan, in CommandInput, tool string) (*mcp.CallToolResult, any, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout(in.TimeoutSeconds))
	defer cancel()

	result, err := runner.Run(ctx, plan, in.Stdin, "mcp:"+tool, s.Policy.MaxOutputBytes)
	if result == nil {
		s.logCall(tool, in.Command, "failed: "+err.Error())
		return errorResult("%s", err), nil, nil
	}

	text := &strings.Builder{}
	stdout := strings.TrimSpace(result.Stdout)
	if stdout != "" {
		text.WriteString(stdout)
		text.WriteString("\n")
	}
	if result.Truncated {
		fmt.Fprintf(text, "\n[output truncated after %d bytes. Use --select, --filter or --pageSize to reduce the output]\n", s.Policy.MaxOutputBytes)
	}
	failed := err != nil || result.ExitCode != 0
	if failed {
		fmt.Fprintf(text, "\n[exit code %d]\n", result.ExitCode)
		if err != nil {
			fmt.Fprintf(text, "%s\n", err)
		}
	}
	if result.Stderr != "" {
		fmt.Fprintf(text, "\n[stderr]\n%s\n", result.Stderr)
	}
	if text.Len() == 0 {
		text.WriteString("(no output)")
	}
	if failed {
		s.logCall(tool, in.Command, fmt.Sprintf("failed (exit code %d)", result.ExitCode))
	} else {
		s.logCall(tool, in.Command, "ok")
	}
	return &mcp.CallToolResult{
		IsError: failed,
		Content: []mcp.Content{&mcp.TextContent{Text: text.String()}},
	}, nil, nil
}

func clientSupportsElicitation(req *mcp.CallToolRequest) bool {
	if req == nil || req.Session == nil {
		return false
	}
	params := req.Session.InitializeParams()
	return params != nil && params.Capabilities != nil && params.Capabilities.Elicitation != nil
}

// approve asks the user to approve a write command. The question is returned as an input
// request in the tool result, and the client calls the tool again with the user's answer
// (multi round-trip requests, SEP-2322). For clients using an older protocol version, the
// MCP SDK asks via an elicitation request instead and calls the tool again itself.
//
// It returns the input request to send to the client, nil if the command is approved, or
// an error if it isn't.
func (s *Server) approve(ctx context.Context, req *mcp.CallToolRequest, session SessionInfo, runner *Runner, plan *Plan, in CommandInput) (*mcp.CallToolResult, error) {
	switch s.Policy.Approval {
	case ApprovalNone, ApprovalClient:
		return nil, nil
	}

	if !clientSupportsElicitation(req) {
		if s.Policy.Approval == ApprovalAuto {
			// rely on the client's tool permission prompt
			return nil, nil
		}
		return nil, fmt.Errorf("write commands need to be approved by the user, but the MCP client does not support elicitation. Ask the user to run the command themselves, or change the approval setting of the MCP server")
	}

	key := approvalKey(session, plan, in)

	// The user's answer, when the client calls the tool again
	if req.Params != nil && req.Params.InputResponses != nil {
		response, ok := req.Params.InputResponses[approvalInputID]
		if !ok {
			return nil, fmt.Errorf("the approval of the user is missing")
		}
		if err := s.approvals.use(req.Params.RequestState, key); err != nil {
			return nil, fmt.Errorf("%s. Run the command again to ask the user for approval", err)
		}
		result, ok := response.(*mcp.ElicitResult)
		if !ok || result.Action != "accept" {
			action := "cancel"
			if ok && result.Action != "" {
				action = result.Action
			}
			return nil, fmt.Errorf("the user did not approve the command (%s). Don't retry it unless the user asks you to", action)
		}
		return nil, nil
	}

	message := &strings.Builder{}
	fmt.Fprintf(message, "Allow the agent to run this command against %s", session.Host)
	if session.Tenant != "" {
		fmt.Fprintf(message, " (tenant %s", session.Tenant)
		if session.Username != "" {
			fmt.Fprintf(message, ", user %s", session.Username)
		}
		message.WriteString(")")
	}
	fmt.Fprintf(message, "?\n\n%s\n", plan.String())
	if in.Stdin != "" {
		stdin := in.Stdin
		if len(stdin) > 500 {
			stdin = stdin[:500] + "..."
		}
		fmt.Fprintf(message, "\nInput:\n%s\n", stdin)
	}
	if plan.HasDeviceShell() {
		message.WriteString("\nThe command after \"--\" runs on the device via ssh (Cumulocity remote access).\n")
	} else if preview := s.preview(ctx, runner, plan, in); preview != "" {
		fmt.Fprintf(message, "\nRequests (dry run):\n%s\n", preview)
	}

	return &mcp.CallToolResult{
		InputRequests: mcp.InputRequestMap{
			approvalInputID: &mcp.ElicitParams{
				Message: message.String(),
				RequestedSchema: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
		RequestState: s.approvals.issue(key),
	}, nil
}

// preview uses the c8y dry run mode to show which requests would be sent
func (s *Server) preview(ctx context.Context, runner *Runner, plan *Plan, in CommandInput) string {
	// Commands without a dry run mode would be executed by the preview
	if len(plan.Segments) != 1 || plan.HasDeviceShell() {
		return ""
	}
	seg := plan.Segments[0]
	dryPlan := &Plan{Segments: []Segment{{
		Args:        append(append([]string{}, seg.Args...), "--dry", "--dryFormat", "markdown"),
		CommandPath: seg.CommandPath,
		Read:        true,
	}}}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	result, err := runner.Run(ctx, dryPlan, in.Stdin, "mcp:preview", 4000)
	if err != nil || result == nil || result.ExitCode != 0 {
		return ""
	}
	// dry run output is written to stderr
	out := strings.TrimSpace(result.Stderr)
	if out == "" {
		out = strings.TrimSpace(result.Stdout)
	}
	if len(out) > 3000 {
		out = out[:3000] + "\n..."
	}
	return out
}
