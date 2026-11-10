//go:build linux || darwin

package mcpserver

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCheckSocketDir(t *testing.T) {
	dir := shortTempDir(t)
	if err := CheckSocketDir(dir); err != nil {
		t.Errorf("private directory should be accepted. %s", err)
	}

	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := CheckSocketDir(dir); err != nil {
		t.Errorf("directory which is only readable by others should be accepted. %s", err)
	}

	for _, mode := range []os.FileMode{0o775, 0o757, 0o777} {
		if err := os.Chmod(dir, mode); err != nil {
			t.Fatal(err)
		}
		if err := CheckSocketDir(dir); err == nil {
			t.Errorf("directory with mode %04o should be rejected", mode)
		}
	}

	// shared directories such as /tmp are writable by everyone
	if err := CheckSocketDir("/tmp"); err == nil {
		t.Errorf("/tmp should be rejected")
	}

	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(filepath.Join(dir, "s.sock")); err != nil {
		t.Errorf("listen in a private directory should work. %s", err)
	}
	if _, err := Listen("/tmp/c8y-mcp-test-shared.sock"); err == nil {
		t.Errorf("listen in a shared directory should fail")
	}
}

func TestPeerCredentials(t *testing.T) {
	socketPath := filepath.Join(shortTempDir(t), "s.sock")
	listener, err := Listen(socketPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()

	client, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	server := <-accepted
	defer server.Close()

	// both ends are this test process
	for name, conn := range map[string]net.Conn{"server": server, "client": client} {
		peer, err := PeerCredentials(conn)
		if err != nil {
			t.Fatalf("%s: %s", name, err)
		}
		if peer.PID != os.Getpid() || peer.UID != os.Getuid() {
			t.Errorf("%s: unexpected peer. got pid=%d uid=%d", name, peer.PID, peer.UID)
		}
		if err := CheckPeerUser(conn); err != nil {
			t.Errorf("%s: %s", name, err)
		}
	}

	chain := ProcessChain(os.Getpid())
	if len(chain) == 0 {
		t.Fatalf("expected process chain")
	}
	if chain[0].PID != os.Getpid() || chain[0].PPID != os.Getppid() {
		t.Errorf("unexpected process info: %+v", chain[0])
	}
	exe, _ := os.Executable()
	if chain[0].Executable == "" || filepath.Base(chain[0].Executable) != filepath.Base(exe) {
		t.Errorf("unexpected executable. got=%s, want=%s", chain[0].Executable, exe)
	}
	if !strings.Contains(chain[0].Command(), "-test.") {
		t.Errorf("expected command to include the test arguments. got=%s", chain[0].Command())
	}
}

func newConnPair(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()
	socketPath := filepath.Join(shortTempDir(t), "s.sock")
	listener, err := Listen(socketPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()
	client, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	server := <-accepted
	t.Cleanup(func() { client.Close(); server.Close() })
	return server, client
}

func TestConnectionApprover(t *testing.T) {
	cases := map[string]bool{
		"y\n":   true,
		"YES\n": true,
		"n\n":   false,
		"\n":    false,
		"":      false,
	}
	for input, want := range cases {
		server, _ := newConnPair(t)
		stderr := &bytes.Buffer{}
		approver := &ConnectionApprover{
			Stdin:    strings.NewReader(input),
			Stderr:   stderr,
			Terminal: &sync.Mutex{},
		}
		if got := approver.Approve(context.Background(), server); got != want {
			t.Errorf("input=%q: got=%v, want=%v", input, got, want)
		}
		if !strings.Contains(stderr.String(), "pid "+strconv.Itoa(os.Getpid())) {
			t.Errorf("prompt should show the connecting process. got: %s", stderr.String())
		}
	}
}

func TestConnectionApproverDoesNotReadAhead(t *testing.T) {
	server, _ := newConnPair(t)
	stdin := strings.NewReader("y\nnext prompt\n")
	approver := &ConnectionApprover{Stdin: stdin, Stderr: io.Discard}
	if !approver.Approve(context.Background(), server) {
		t.Fatal("expected approval")
	}
	rest, _ := io.ReadAll(stdin)
	if string(rest) != "next prompt\n" {
		t.Errorf("approval should only consume one line. remaining=%q", rest)
	}
}

func TestConnectionApproverCancel(t *testing.T) {
	server, _ := newConnPair(t)
	stdinReader, stdinWriter := io.Pipe()
	defer stdinWriter.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool, 1)
	approver := &ConnectionApprover{Stdin: stdinReader, Stderr: io.Discard}
	go func() { done <- approver.Approve(ctx, server) }()

	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case got := <-done:
		if got {
			t.Errorf("cancelled approval should reject the connection")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("approval did not stop when the context was cancelled")
	}
}

func TestServeRejectsUnauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	socketPath := filepath.Join(shortTempDir(t), "s.sock")
	listener, err := Listen(socketPath)
	if err != nil {
		t.Fatal(err)
	}

	serveCtx, stopServe := context.WithCancel(ctx)
	served := make(chan error, 1)
	rejected := make(chan struct{}, 1)
	connected := make(chan int, 1)
	go func() {
		served <- Serve(serveCtx, newTestServer().New(), listener, ServeOptions{
			Gates:      &ConnectionGates{},
			Authorize:  func(ctx context.Context, conn net.Conn) bool { return false },
			OnRejected: func() { rejected <- struct{}{} },
			OnConnect:  func(n int) { connected <- n },
		})
	}()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	select {
	case <-rejected:
	case <-connected:
		t.Fatal("unauthorized client should not be connected")
	case <-time.After(5 * time.Second):
		t.Fatal("connection was not rejected")
	}

	// the server closes the connection
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Read(make([]byte, 1)); err != io.EOF {
		t.Errorf("expected the connection to be closed. got %v", err)
	}

	stopServe()
	if err := <-served; err != nil {
		t.Errorf("unexpected serve error: %s", err)
	}
}

// approveAsync starts an approval and returns a channel with the result
func approveAsync(a *ConnectionApprover, conn net.Conn) chan bool {
	result := make(chan bool, 1)
	go func() { result <- a.Approve(context.Background(), conn) }()
	return result
}

func waitResult(t *testing.T, name string, result chan bool) bool {
	t.Helper()
	select {
	case v := <-result:
		return v
	case <-time.After(5 * time.Second):
		t.Fatalf("%s: approval did not finish", name)
		return false
	}
}

// waitForWaiting waits until the approver has n connections waiting
func waitForWaiting(t *testing.T, a *ConnectionApprover, n int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		a.mu.Lock()
		count := 0
		for _, conns := range a.waiting {
			count += len(conns)
		}
		a.mu.Unlock()
		if count == n {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected %d waiting connections", n)
}

// requireClientKey skips the test if the test process has no parent which can
// identify the client (e.g. when go test runs as pid 1 in a container)
func requireClientKey(t *testing.T) {
	t.Helper()
	if chain := ProcessChain(os.Getpid()); len(chain) < 2 || chain[1].Executable == "" {
		t.Skip("the parent of the test process is unknown")
	}
}

func TestConnectionApproverOneAnswerForAllWaitingConnections(t *testing.T) {
	requireClientKey(t)
	stdinReader, stdinWriter := io.Pipe()
	defer stdinWriter.Close()
	stderr := &syncBuffer{}
	terminal := &sync.Mutex{}
	approver := &ConnectionApprover{Stdin: stdinReader, Stderr: stderr, Terminal: terminal, RelayExecutable: testExecutable(t)}

	// Hold the terminal (e.g. busy with a login) until both connections are waiting
	terminal.Lock()
	server1, client1 := newConnPair(t)
	server2, _ := newConnPair(t)
	result1 := approveAsync(approver, server1)
	result2 := approveAsync(approver, server2)
	waitForWaiting(t, approver, 2)
	terminal.Unlock()

	// One of the connections is closed (e.g. replaced by the client), but the
	// prompt stays open as the other connection is still waiting
	time.Sleep(100 * time.Millisecond)
	client1.Close()
	time.Sleep(3 * pollInterval)

	// a single answer approves all waiting connections
	if _, err := stdinWriter.Write([]byte("y\n")); err != nil {
		t.Fatal(err)
	}
	// The closed connection is approved too, but it ends straight away as the client is gone
	ok1, ok2 := waitResult(t, "first", result1), waitResult(t, "second", result2)
	if !ok1 || !ok2 {
		t.Errorf("expected the waiting connections to be approved. first=%v, second=%v\n%s", ok1, ok2, stderr.String())
	}
	if n := strings.Count(stderr.String(), "[y/N/a]"); n != 1 {
		t.Errorf("expected one prompt, got %d\n%s", n, stderr.String())
	}
	if !strings.Contains(stderr.String(), "2 connections") {
		t.Errorf("prompt should mention the number of connections\n%s", stderr.String())
	}

	// A later connection is asked again, as "y" only applies to the waiting connections
	server3, _ := newConnPair(t)
	result3 := approveAsync(approver, server3)
	if _, err := stdinWriter.Write([]byte("n\n")); err != nil {
		t.Fatal(err)
	}
	if waitResult(t, "third", result3) {
		t.Errorf("later connection should have been asked again and rejected")
	}
}

func TestConnectionApproverAlwaysAllow(t *testing.T) {
	requireClientKey(t)
	stderr := &syncBuffer{}
	approver := &ConnectionApprover{Stdin: strings.NewReader("a\n"), Stderr: stderr, Terminal: &sync.Mutex{}, RelayExecutable: testExecutable(t)}

	server1, _ := newConnPair(t)
	if !approver.Approve(context.Background(), server1) {
		t.Fatalf("expected approval\n%s", stderr.String())
	}
	// no more input is available, so later connections must be allowed without a prompt
	for i := range 2 {
		server, _ := newConnPair(t)
		if !waitResult(t, "later", approveAsync(approver, server)) {
			t.Errorf("connection %d should be allowed without a prompt", i)
		}
	}
	if n := strings.Count(stderr.String(), "[y/N/a]"); n != 1 {
		t.Errorf("expected one prompt, got %d\n%s", n, stderr.String())
	}
}

func TestConnectionApproverClientDisconnects(t *testing.T) {
	stdinReader, stdinWriter := io.Pipe()
	defer stdinWriter.Close()
	stderr := &syncBuffer{}
	approver := &ConnectionApprover{Stdin: stdinReader, Stderr: stderr, Terminal: &sync.Mutex{}}

	server, client := newConnPair(t)
	result := approveAsync(approver, server)
	time.Sleep(300 * time.Millisecond)
	client.Close()

	if waitResult(t, "disconnected", result) {
		t.Errorf("connection should not be approved after the client disconnected")
	}
	if !strings.Contains(stderr.String(), "disconnected before it was approved") {
		t.Errorf("expected a message about the disconnect\n%s", stderr.String())
	}
}

// syncBuffer bytes.Buffer which can be written from several goroutines
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestConnectionApproverFileStdin(t *testing.T) {
	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stdinReader.Close()
	defer stdinWriter.Close()
	stderr := &syncBuffer{}
	approver := &ConnectionApprover{Stdin: stdinReader, Stderr: stderr, Terminal: &sync.Mutex{}}

	// approved via the terminal
	server, _ := newConnPair(t)
	result := approveAsync(approver, server)
	time.Sleep(100 * time.Millisecond)
	if _, err := stdinWriter.Write([]byte("y\n")); err != nil {
		t.Fatal(err)
	}
	if !waitResult(t, "approve", result) {
		t.Fatalf("expected approval\n%s", stderr.String())
	}

	// cancelled when the client disconnects, without reading any input afterwards
	server, client := newConnPair(t)
	result = approveAsync(approver, server)
	time.Sleep(300 * time.Millisecond)
	client.Close()
	if waitResult(t, "disconnect", result) {
		t.Fatalf("connection should not be approved after the client disconnected")
	}
	if _, err := stdinWriter.Write([]byte("for the next prompt\n")); err != nil {
		t.Fatal(err)
	}
	line, err := readLine(stdinReader)
	if err != nil || line != "for the next prompt" {
		t.Errorf("input after a cancelled prompt should not be consumed. got=%q, err=%v", line, err)
	}
}

// testExecutable the executable of the test process, which connects to the test
// sockets, so it takes the role of the c8y relay
func testExecutable(t *testing.T) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return exe
}

func TestConnectionApproverOnlyTrustsTheRelay(t *testing.T) {
	requireClientKey(t)
	stderr := &syncBuffer{}
	// the connecting process (the test binary) is not the c8y executable
	approver := &ConnectionApprover{Stdin: strings.NewReader("a\ny\n"), Stderr: stderr, Terminal: &sync.Mutex{}, RelayExecutable: "/usr/bin/true"}

	server, _ := newConnPair(t)
	if approver.Approve(context.Background(), server) {
		t.Errorf("\"a\" should not be accepted for a process which is not the relay")
	}
	if strings.Contains(stderr.String(), "[y/N/a]") || !strings.Contains(stderr.String(), "not the c8y relay") {
		t.Errorf("\"a\" should not be offered\n%s", stderr.String())
	}
	// and nothing is trusted, so the next connection is asked again (and allowed with y)
	server, _ = newConnPair(t)
	if !approver.Approve(context.Background(), server) {
		t.Errorf("expected approval with y")
	}
	if n := strings.Count(stderr.String(), "[y/N]"); n != 2 {
		t.Errorf("expected two prompts, got %d\n%s", n, stderr.String())
	}
}

func TestConnectionApproverTrustExpires(t *testing.T) {
	requireClientKey(t)
	stderr := &syncBuffer{}
	approver := &ConnectionApprover{
		Stdin:            strings.NewReader("a\nn\n"),
		Stderr:           stderr,
		Terminal:         &sync.Mutex{},
		RelayExecutable:  testExecutable(t),
		TrustIdleTimeout: 200 * time.Millisecond,
	}
	server, _ := newConnPair(t)
	if !approver.Approve(context.Background(), server) {
		t.Fatalf("expected approval")
	}
	// trusted while the client keeps connecting
	for range 3 {
		time.Sleep(100 * time.Millisecond)
		server, _ = newConnPair(t)
		if !approver.Approve(context.Background(), server) {
			t.Fatalf("client should still be trusted\n%s", stderr.String())
		}
	}
	// asked again after being idle
	time.Sleep(300 * time.Millisecond)
	server, _ = newConnPair(t)
	if approver.Approve(context.Background(), server) {
		t.Errorf("trust should have expired, and the user answered n\n%s", stderr.String())
	}
}

func TestConnectionApproverNotifies(t *testing.T) {
	requireClientKey(t)
	stderr := &syncBuffer{}
	notified := make(chan string, 1)
	approver := &ConnectionApprover{
		Stdin:    strings.NewReader("y\n"),
		Stderr:   stderr,
		Terminal: &sync.Mutex{},
		Notify:   func(message string) { notified <- message },
	}
	server, _ := newConnPair(t)
	if !approver.Approve(context.Background(), server) {
		t.Fatalf("expected approval")
	}
	select {
	case msg := <-notified:
		if !strings.Contains(msg, "c8y mcp serve") {
			t.Errorf("unexpected notification: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("expected a notification")
	}
	if !strings.Contains(stderr.String(), "\a") {
		t.Errorf("expected a terminal bell")
	}
}
