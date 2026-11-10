//go:build linux || darwin

package mcpserver

import (
	"context"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServeDoesNotBlockConnectionsWhileWaitingForApproval(t *testing.T) {
	for _, allow := range []bool{true, false} {
		name := "approved"
		if !allow {
			name = "rejected"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			socketPath := filepath.Join(shortTempDir(t), "s.sock")
			listener, err := Listen(socketPath)
			if err != nil {
				t.Fatal(err)
			}

			gates := &ConnectionGates{MaxWait: 300 * time.Millisecond}
			server := newTestServer()
			server.Gate = gates.Check
			answer := make(chan bool)
			serveCtx, stopServe := context.WithCancel(ctx)
			served := make(chan error, 1)
			go func() {
				served <- Serve(serveCtx, server.New(), listener, ServeOptions{
					Gates: gates,
					Authorize: func(ctx context.Context, conn net.Conn) bool {
						select {
						case v := <-answer:
							return v
						case <-ctx.Done():
							return false
						}
					},
				})
			}()

			conn, err := net.Dial("unix", socketPath)
			if err != nil {
				t.Fatal(err)
			}
			client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
			// the handshake completes while the user hasn't answered yet
			start := time.Now()
			cs, err := client.Connect(ctx, &mcp.IOTransport{Reader: conn, Writer: conn}, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer cs.Close()
			if time.Since(start) > 2*time.Second {
				t.Errorf("connecting should not wait for the approval")
			}
			if tools, err := cs.ListTools(ctx, nil); err != nil || len(tools.Tools) != 3 {
				t.Errorf("tools should be listed before the approval. err=%v", err)
			}

			call := func() (*mcp.CallToolResult, error) {
				return cs.CallTool(ctx, &mcp.CallToolParams{Name: ToolSession})
			}
			res, err := call()
			if err != nil {
				t.Fatal(err)
			}
			if !res.IsError || !strings.Contains(resultText(res), "not been approved yet") || strings.Contains(resultText(res), "t123") {
				t.Errorf("tool call should wait for the approval: %s", resultText(res))
			}

			answer <- allow
			res, err = call()
			if allow {
				if err != nil || res.IsError || !strings.Contains(resultText(res), "t123") {
					t.Errorf("tool call should work after the approval. err=%v, result=%s", err, resultText(res))
				}
			} else if err == nil && !res.IsError {
				t.Errorf("tool call should fail after the client was rejected: %s", resultText(res))
			}

			stopServe()
			if err := <-served; err != nil {
				t.Errorf("unexpected serve error: %s", err)
			}
		})
	}
}

func TestConnectionGatesUnknownSession(t *testing.T) {
	gates := &ConnectionGates{MaxWait: 100 * time.Millisecond}
	if err := gates.Check(context.Background(), nil); err != ErrClientPending {
		t.Errorf("unknown sessions should never be allowed, got %v", err)
	}
}
