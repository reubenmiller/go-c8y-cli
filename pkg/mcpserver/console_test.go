//go:build linux || darwin

package mcpserver

import (
	"context"
	"os"
	"strings"
	"sync"
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

func TestRunConsoleSharesTerminalWithPrompts(t *testing.T) {
	requireClientKey(t)
	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stdinReader.Close()
	defer stdinWriter.Close()

	ctx, cancel := context.WithCancel(context.Background())
	terminal := &sync.Mutex{}
	commands := make(chan string, 10)
	consoleDone := make(chan error, 1)
	go func() {
		consoleDone <- RunConsole(ctx, stdinReader, terminal, func(command string) { commands <- command })
	}()

	expectCommand := func(want string) {
		t.Helper()
		select {
		case got := <-commands:
			if got != want {
				t.Errorf("unexpected command. got=%q, want=%q", got, want)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("command %q was not received", want)
		}
	}

	if _, err := stdinWriter.Write([]byte("s\n")); err != nil {
		t.Fatal(err)
	}
	expectCommand("s")

	// While a connection prompt is shown, the answer goes to the prompt and not the console
	stderr := &syncBuffer{}
	approver := &ConnectionApprover{Stdin: stdinReader, Stderr: stderr, Terminal: terminal, RelayExecutable: testExecutable(t)}
	server, _ := newConnPair(t)
	result := approveAsync(approver, server)
	deadline := time.Now().Add(5 * time.Second)
	for !strings.Contains(stderr.String(), "[y/N") {
		if time.Now().After(deadline) {
			t.Fatalf("prompt was not shown")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := stdinWriter.Write([]byte("y\n")); err != nil {
		t.Fatal(err)
	}
	if !waitResult(t, "approve", result) {
		t.Fatalf("connection should be approved\n%s", stderr.String())
	}
	select {
	case got := <-commands:
		t.Fatalf("the answer to the prompt should not be handled as a command: %q", got)
	case <-time.After(300 * time.Millisecond):
	}

	// the console continues to read commands after the prompt
	if _, err := stdinWriter.Write([]byte("  i  \n")); err != nil {
		t.Fatal(err)
	}
	expectCommand("i")

	cancel()
	select {
	case err := <-consoleDone:
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("console did not stop")
	}
}
