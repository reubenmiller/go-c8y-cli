package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ErrClientPending the user has not approved the client yet
var ErrClientPending = errors.New("this MCP client has not been approved yet. Ask the user to approve it in the terminal running \"c8y mcp serve\" (answer y, or a to also allow later connections), then try again")

// ErrClientRejected the user rejected the client
var ErrClientRejected = errors.New("the user rejected this MCP client in the terminal running \"c8y mcp serve\"")

// gate the approval state of a connection
type gate struct {
	done    chan struct{}
	allowed bool
}

// ConnectionGates tracks which MCP sessions the user has approved. Clients can
// connect (and list the tools) before they are approved, so the MCP client never
// times out while the user is asked. Tool calls wait for the approval for a short
// time, and then tell the agent to ask the user to approve the client.
type ConnectionGates struct {
	// MaxWait how long a tool call waits for the approval
	MaxWait time.Duration

	mu    sync.Mutex
	gates map[*mcp.ServerSession]*gate
}

func (g *ConnectionGates) add(ss *mcp.ServerSession) *gate {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.gates == nil {
		g.gates = map[*mcp.ServerSession]*gate{}
	}
	gt := &gate{done: make(chan struct{})}
	g.gates[ss] = gt
	return gt
}

func (g *ConnectionGates) remove(ss *mcp.ServerSession) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.gates, ss)
}

func (g *ConnectionGates) get(ss *mcp.ServerSession) *gate {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.gates[ss]
}

func (gt *gate) resolve(allowed bool) {
	gt.allowed = allowed
	close(gt.done)
}

// Check returns nil if the session has been approved. Otherwise it waits up to
// MaxWait for the user's answer. Unknown sessions are never allowed.
func (g *ConnectionGates) Check(ctx context.Context, ss *mcp.ServerSession) error {
	maxWait := g.MaxWait
	if maxWait <= 0 {
		maxWait = 10 * time.Second
	}
	timer := time.NewTimer(maxWait)
	defer timer.Stop()

	// The session is registered right after it is created, so a request can
	// arrive just before that
	var gt *gate
	for gt == nil {
		if gt = g.get(ss); gt != nil {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return ErrClientPending
		case <-time.After(20 * time.Millisecond):
		}
	}

	select {
	case <-gt.done:
		if gt.allowed {
			return nil
		}
		return ErrClientRejected
	case <-ctx.Done():
		return fmt.Errorf("%w. %w", ErrClientPending, ctx.Err())
	case <-timer.C:
		return ErrClientPending
	}
}
