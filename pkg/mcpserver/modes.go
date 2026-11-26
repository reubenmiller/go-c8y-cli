package mcpserver

import (
	"fmt"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
)

// Mode sources, used in the log and in the mode change result
const (
	ModeSourceSession  = "session"
	ModeSourceOverride = "override"
)

// normalizeMode maps the ci mode to dev, so the agent never gets more than dev
func normalizeMode(mode config.SessionMode) config.SessionMode {
	if mode != config.SessionModeUnset && mode > config.SessionModeDev {
		return config.SessionModeDev
	}
	return mode
}

// Ceiling returns the maximum mode the agent can use: the policy mode, or dev if unset
func (p Policy) Ceiling() config.SessionMode {
	if p.Mode == config.SessionModeUnset {
		return config.SessionModeDev
	}
	return normalizeMode(p.Mode)
}

// SessionDefaultMode returns the mode of the session itself, prod (read only) if it has none
func SessionDefaultMode(sessionMode config.SessionMode) config.SessionMode {
	if sessionMode == config.SessionModeUnset {
		return config.SessionModeProduction
	}
	return normalizeMode(sessionMode)
}

// ResolveMode returns the mode the agent can use:
// min(ceiling, override if set, else the session mode). An unset ceiling is dev, an unset
// session mode is prod, and the ci mode counts as dev.
func ResolveMode(ceiling, sessionMode, override config.SessionMode) config.SessionMode {
	if ceiling == config.SessionModeUnset {
		ceiling = config.SessionModeDev
	}
	ceiling = normalizeMode(ceiling)
	mode := SessionDefaultMode(sessionMode)
	if override != config.SessionModeUnset {
		mode = normalizeMode(override)
	}
	if ceiling < mode {
		return ceiling
	}
	return mode
}

// RequiredMode returns the lowest mode which allows the plan. Plans which only read
// data don't need a mode, and any other plan needs to be able to create data
func RequiredMode(plan *Plan) config.SessionMode {
	if plan.ReadOnly() {
		return config.SessionModeProduction
	}
	return config.SessionModeQual
}

// CheckMode checks that the effective mode allows a plan which needs the required mode.
// It is shared by every entry point which runs commands.
func CheckMode(effective, required config.SessionMode) error {
	if effective >= required {
		return nil
	}
	if effective == config.SessionModeProduction {
		return fmt.Errorf("the MCP server is running in read-only mode (prod) so write commands are disabled. Ask the user to change the mode if this is needed")
	}
	return fmt.Errorf("the command needs mode %s but the MCP server is running in mode %s. Ask the user to change the mode if this is needed", required, effective)
}

// WithMode returns a copy of the runner which uses another mode. The credentials are
// shared, so no new login is needed
func (r *Runner) WithMode(mode config.SessionMode) *Runner {
	c := *r
	c.Mode = mode
	return &c
}

// ModeChange describes the result of a mode change
type ModeChange struct {
	Session string
	From    config.SessionMode
	To      config.SessionMode
	// Source is ModeSourceOverride, or ModeSourceSession after a reset
	Source string
	// SessionMode the session's own mode
	SessionMode config.SessionMode
}

// ModeController changes the mode of the active session. Only the user may call it, via
// a transport's picker or prompts, never via a tool of the agent
type ModeController interface {
	SetMode(mode config.SessionMode, via string) (ModeChange, error)
}

// ParseModeOrReset parses a mode name, or "reset" (returned as unset)
func ParseModeOrReset(name string) (config.SessionMode, error) {
	if strings.EqualFold(strings.TrimSpace(name), "reset") {
		return config.SessionModeUnset, nil
	}
	return ParseMode(name)
}

// ModeInfo describes what the agent can do in a mode
func ModeInfo(mode config.SessionMode) string {
	switch mode {
	case config.SessionModeQual:
		return "read, create and update"
	case config.SessionModeDev:
		return "read, create, update and delete"
	default:
		return "read only"
	}
}
