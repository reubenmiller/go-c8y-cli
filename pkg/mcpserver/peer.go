package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// maxProcessChain number of processes shown for a new connection (the peer and its parents)
const maxProcessChain = 4

// maxCommandLength commands are truncated to this length when shown to the user
const maxCommandLength = 120

// ProcessInfo details of a local process
type ProcessInfo struct {
	PID        int
	PPID       int
	Name       string
	Executable string
	Args       []string

	// StartTime when the process started, in a platform specific unit. Together with
	// the pid, it identifies the process, as pids are reused
	StartTime int64
}

// Command returns the command line of the process for display
func (p ProcessInfo) Command() string {
	var cmd string
	if len(p.Args) > 0 {
		args := append([]string{filepath.Base(p.Args[0])}, p.Args[1:]...)
		cmd = strings.Join(args, " ")
	} else if p.Name != "" {
		cmd = p.Name
	} else {
		cmd = "unknown"
	}
	if len(cmd) > maxCommandLength {
		cmd = cmd[:maxCommandLength-3] + "..."
	}
	return cmd
}

// PeerInfo the process connected to the other end of a unix socket
type PeerInfo struct {
	PID int
	UID int
}

// PeerCredentials returns the process and user of the other end of a unix
// socket connection, as reported by the operating system
func PeerCredentials(conn net.Conn) (PeerInfo, error) {
	uc, ok := conn.(*net.UnixConn)
	if !ok {
		return PeerInfo{}, errors.New("not a unix socket connection")
	}
	pid, uid, err := peerCredentials(uc)
	if err != nil {
		return PeerInfo{}, err
	}
	return PeerInfo{PID: pid, UID: uid}, nil
}

// CheckPeerUser returns an error if the other end of the connection belongs to
// another user. If the platform can't report the peer, the check is skipped and
// the socket file permissions are relied upon.
func CheckPeerUser(conn net.Conn) error {
	peer, err := PeerCredentials(conn)
	if err != nil {
		return nil
	}
	if uid := os.Getuid(); uid >= 0 && peer.UID != uid {
		return fmt.Errorf("the other end of the socket belongs to another user (uid %d)", peer.UID)
	}
	return nil
}

// ProcessChain returns the process and its parents, starting with the process itself
func ProcessChain(pid int) []ProcessInfo {
	chain := make([]ProcessInfo, 0, maxProcessChain)
	for len(chain) < maxProcessChain && pid > 0 {
		info, err := processInfo(pid)
		if err != nil {
			break
		}
		chain = append(chain, info)
		if info.PPID <= 1 || info.PPID == pid {
			break
		}
		pid = info.PPID
	}
	return chain
}

// errWaitUnsupported waiting for input without blocking is not supported
var errWaitUnsupported = errors.New("waiting for input is not supported")

// errPeerDisconnected the client disconnected while waiting for the approval
var errPeerDisconnected = errors.New("client disconnected")

// pollInterval how often to check for input and if the client is still connected
const pollInterval = 200 * time.Millisecond

// clientKey identifies the MCP client which started the relay (the parent of the
// connecting process), e.g. a Claude Code session
type clientKey struct {
	pid        int
	executable string
	startTime  int64
}

// DefaultTrustIdleTimeout how long a client stays trusted ("a") without connecting
const DefaultTrustIdleTimeout = 8 * time.Hour

// ConnectionApprover asks the user in the server terminal whether a new client
// may connect. The approval can't be given by the client itself, unlike the
// approval of write commands via MCP elicitation or the client's tool prompt.
//
// MCP clients can open several connections at once (e.g. Claude Code when
// reconnecting), so one answer applies to all connections from the same client
// which are waiting for approval.
type ConnectionApprover struct {
	// Stdin, Stderr server terminal
	Stdin  io.Reader
	Stderr io.Writer

	// Terminal lock which is held while the terminal is used (shared with the login)
	Terminal sync.Locker

	// RelayExecutable path of the c8y executable. Later connections from a client can
	// only be allowed ("a") if the connecting process runs this executable, i.e. it
	// is the "c8y mcp connect" relay started by the client
	RelayExecutable string

	// TrustIdleTimeout how long a client stays trusted without connecting.
	// Defaults to DefaultTrustIdleTimeout
	TrustIdleTimeout time.Duration

	// Notify optional callback to draw the user's attention to the prompt, e.g. with a
	// desktop notification
	Notify func(message string)

	mu        sync.Mutex
	seq       int
	trusted   map[clientKey]time.Time
	decisions map[clientKey]decision
	waiting   map[clientKey]map[net.Conn]struct{}
}

// decision answer given for a client, which applies to its connections which
// arrived before the answer was given
type decision struct {
	allow bool
	seq   int
}

// register records a connection which is waiting for approval and returns its arrival number
func (a *ConnectionApprover) register(key *clientKey, conn net.Conn) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.seq++
	if key != nil {
		if a.waiting == nil {
			a.waiting = map[clientKey]map[net.Conn]struct{}{}
		}
		if a.waiting[*key] == nil {
			a.waiting[*key] = map[net.Conn]struct{}{}
		}
		a.waiting[*key][conn] = struct{}{}
	}
	return a.seq
}

func (a *ConnectionApprover) unregister(key *clientKey, conn net.Conn) {
	if key == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.waiting[*key], conn)
	if len(a.waiting[*key]) == 0 {
		delete(a.waiting, *key)
	}
}

// waitingConnections returns the connections from the client which are waiting
// for approval and are still connected
func (a *ConnectionApprover) waitingConnections(key *clientKey, conn net.Conn) []net.Conn {
	conns := []net.Conn{conn}
	if key != nil {
		a.mu.Lock()
		conns = conns[:0]
		for c := range a.waiting[*key] {
			conns = append(conns, c)
		}
		a.mu.Unlock()
	}
	alive := conns[:0]
	for _, c := range conns {
		if !isPeerClosed(c) {
			alive = append(alive, c)
		}
	}
	return alive
}

// earlierDecision returns the answer given for the client after the connection arrived
func (a *ConnectionApprover) earlierDecision(key *clientKey, arrival int) (allow bool, ok bool) {
	if key == nil {
		return false, false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if lastUsed, ok := a.trusted[*key]; ok {
		timeout := a.TrustIdleTimeout
		if timeout <= 0 {
			timeout = DefaultTrustIdleTimeout
		}
		if time.Since(lastUsed) < timeout {
			a.trusted[*key] = time.Now()
			return true, true
		}
		delete(a.trusted, *key)
	}
	if d, found := a.decisions[*key]; found && arrival <= d.seq {
		return d.allow, true
	}
	return false, false
}

func (a *ConnectionApprover) decide(key *clientKey, allow bool, always bool) {
	if key == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.decisions == nil {
		a.decisions = map[clientKey]decision{}
	}
	a.decisions[*key] = decision{allow: allow, seq: a.seq}
	if always {
		if a.trusted == nil {
			a.trusted = map[clientKey]time.Time{}
		}
		a.trusted[*key] = time.Now()
	}
}

// Approve returns true if the user allows the connection
func (a *ConnectionApprover) Approve(ctx context.Context, conn net.Conn) bool {
	peer, peerErr := PeerCredentials(conn)
	if peerErr == nil {
		if uid := os.Getuid(); uid >= 0 && peer.UID != uid {
			fmt.Fprintf(a.Stderr, "Rejected a connection from another user (uid %d, pid %d)\n", peer.UID, peer.PID)
			return false
		}
	}

	var chain []ProcessInfo
	var client *clientKey
	if peerErr == nil {
		chain = ProcessChain(peer.PID)
		// The relay is started by the MCP client, so its parent identifies the client.
		// This is only used if the connecting process is the c8y relay; otherwise any
		// program started by the client could connect in its name
		if len(chain) > 1 && chain[1].Executable != "" && a.isRelay(chain[0]) {
			client = &clientKey{pid: chain[1].PID, executable: chain[1].Executable, startTime: chain[1].StartTime}
		}
	}

	arrival := a.register(client, conn)
	defer a.unregister(client, conn)

	useEarlierDecision := func() (bool, bool) {
		allow, ok := a.earlierDecision(client, arrival)
		if ok && allow {
			fmt.Fprintf(a.Stderr, "Allowed another connection from %s (pid %d)\n", chain[1].Command(), client.pid)
		}
		return allow, ok
	}

	if allow, ok := useEarlierDecision(); ok {
		return allow
	}

	if a.Terminal != nil {
		a.Terminal.Lock()
		defer a.Terminal.Unlock()
	}
	if ctx.Err() != nil {
		return false
	}
	// The client may have been answered, or have given up, while waiting for the terminal
	if allow, ok := useEarlierDecision(); ok {
		return allow
	}
	waiting := a.waitingConnections(client, conn)
	if len(waiting) == 0 {
		return false
	}

	fmt.Fprintf(a.Stderr, "\nA new MCP client is connecting:\n")
	if peerErr != nil {
		fmt.Fprintf(a.Stderr, "  unknown process (%s)\n", peerErr)
	} else {
		if len(chain) == 0 {
			fmt.Fprintf(a.Stderr, "  unknown process (pid %d)\n", peer.PID)
		}
		for i, p := range chain {
			prefix := "  "
			if i > 0 {
				prefix = "  started by: "
			}
			fmt.Fprintf(a.Stderr, "%s%s\n", prefix, p.Command())
			details := fmt.Sprintf("pid %d", p.PID)
			if p.Executable != "" {
				details += ", " + p.Executable
			}
			fmt.Fprintf(a.Stderr, "      %s\n", details)
		}
	}
	if len(waiting) > 1 {
		fmt.Fprintf(a.Stderr, "The client has opened %d connections. Your answer applies to all of them.\n", len(waiting))
	}
	if peerErr == nil && client == nil && len(chain) > 0 {
		fmt.Fprintf(a.Stderr, "The connecting process is not the c8y relay (c8y mcp connect), so only this connection can be allowed.\n")
	}
	if client != nil {
		fmt.Fprintf(a.Stderr, "Allow this client to use the MCP server? (a = also allow later connections from %s, pid %d) [y/N/a] ", filepath.Base(client.executable), client.pid)
	} else {
		fmt.Fprintf(a.Stderr, "Allow this client to use the MCP server? [y/N] ")
	}

	if a.Notify != nil {
		// terminal bell
		fmt.Fprint(a.Stderr, "\a")
		name := "An MCP client"
		if len(chain) > 1 {
			name = filepath.Base(chain[1].Command())
		}
		go a.Notify(fmt.Sprintf("%s wants to use the Cumulocity MCP server. Approve it in the c8y mcp serve terminal.", name))
	}

	line, err := a.readAnswer(ctx, func() bool { return len(a.waitingConnections(client, conn)) > 0 })
	if err != nil {
		fmt.Fprintln(a.Stderr)
		if errors.Is(err, errPeerDisconnected) {
			fmt.Fprintf(a.Stderr, "The client disconnected before it was approved\n")
		}
		return false
	}

	allow, always := false, false
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		allow = true
	case "a", "always":
		if client != nil {
			allow, always = true, true
		}
	}
	a.decide(client, allow, always)
	if !allow {
		fmt.Fprintf(a.Stderr, "Connection rejected\n")
	}
	return allow
}

// isRelay returns true if the process runs the c8y executable (the relay)
func (a *ConnectionApprover) isRelay(p ProcessInfo) bool {
	if a.RelayExecutable == "" || p.Executable == "" {
		return false
	}
	return sameFile(a.RelayExecutable, p.Executable)
}

func sameFile(a, b string) bool {
	if a == b {
		return true
	}
	if resolved, err := filepath.EvalSymlinks(a); err == nil {
		a = resolved
	}
	if resolved, err := filepath.EvalSymlinks(b); err == nil {
		b = resolved
	}
	if a == b {
		return true
	}
	infoA, errA := os.Stat(a)
	infoB, errB := os.Stat(b)
	return errA == nil && errB == nil && os.SameFile(infoA, infoB)
}

func isPeerClosed(conn net.Conn) bool {
	if uc, ok := conn.(*net.UnixConn); ok {
		return peerClosed(uc)
	}
	return false
}

// readAnswer reads a line from the terminal. It stops if the context is cancelled
// or connected returns false (the client disconnected).
func (a *ConnectionApprover) readAnswer(ctx context.Context, connected func() bool) (string, error) {
	if f, ok := a.Stdin.(*os.File); ok {
		if _, err := waitReadable(f, 0); err == nil {
			for {
				if ctx.Err() != nil {
					return "", ctx.Err()
				}
				if !connected() {
					return "", errPeerDisconnected
				}
				line, ok, err := readLineIfReady(f, pollInterval)
				if err != nil {
					return line, err
				}
				if ok {
					return line, nil
				}
			}
		}
	}

	// Fallback for readers which can't be waited on (e.g. in tests)
	answer := make(chan string, 1)
	go func() {
		line, _ := readLine(a.Stdin)
		answer <- line
	}()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case line := <-answer:
			return line, nil
		case <-ticker.C:
			if !connected() {
				return "", errPeerDisconnected
			}
		}
	}
}

// readLineIfReady reads a line if input is available within the timeout. It only
// reads when input is available, so nothing is left waiting on the terminal which
// could take input meant for a later prompt. A terminal only reports input once a
// whole line has been entered.
func readLineIfReady(f *os.File, timeout time.Duration) (string, bool, error) {
	ready, err := waitReadable(f, timeout)
	if err != nil || !ready {
		return "", false, err
	}
	var sb strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				return sb.String(), true, nil
			}
			sb.WriteByte(buf[0])
		}
		if err != nil {
			return sb.String(), sb.Len() > 0, err
		}
		if n == 0 {
			return sb.String(), sb.Len() > 0, io.EOF
		}
		// continue only while more input is available, e.g. the rest of the line
		if more, err := waitReadable(f, pollInterval); err != nil || !more {
			return sb.String(), true, err
		}
	}
}

// readLine reads a single line one byte at a time, so no input is consumed
// which is meant for a later prompt (e.g. the login)
func readLine(r io.Reader) (string, error) {
	var sb strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				return sb.String(), nil
			}
			sb.WriteByte(buf[0])
		}
		if err != nil {
			return sb.String(), err
		}
	}
}

// splitNul splits a NUL separated list, skipping empty values
func splitNul(b []byte) []string {
	parts := strings.Split(string(b), "\x00")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
