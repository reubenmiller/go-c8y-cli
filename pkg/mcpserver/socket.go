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

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/pathresolver"
)

// maxSocketPathLength unix socket paths are limited to 104 bytes on macOS and
// 108 bytes on Linux (including the null terminator)
const maxSocketPathLength = 100

// SocketDir directory containing the MCP server sockets
func SocketDir(sessionHome string) string {
	return filepath.Join(sessionHome, "mcp")
}

// SessionKey name used to identify a session, derived from the session file.
// Only known session file extensions are removed, as session names can contain
// dots, e.g. "example.cumulocity.com-firstname.lastname"
func SessionKey(sessionFile string) string {
	if sessionFile == "" {
		return "default"
	}
	name := filepath.Base(sessionFile)
	if ext := filepath.Ext(name); isSessionFileExtension(ext) {
		name = strings.TrimSuffix(name, ext)
	}
	return name
}

func isSessionFileExtension(ext string) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	for _, v := range config.ConfigExtensions {
		if ext == v {
			return true
		}
	}
	return false
}

// ResolveSessionFile returns the absolute path of the session file. A session can be
// given by name (e.g. --session my-tenant), in which case the matching file is looked
// up in the session home directory. The value is returned as is if no file is found.
func ResolveSessionFile(sessionFile string, sessionHome string) string {
	sessionFile = strings.TrimPrefix(sessionFile, "file://")
	if sessionFile == "" {
		return ""
	}
	if info, err := os.Stat(sessionFile); err == nil && !info.IsDir() {
		if abs, err := filepath.Abs(sessionFile); err == nil {
			return abs
		}
		return sessionFile
	}
	if sessionHome != "" && !strings.ContainsRune(sessionFile, os.PathSeparator) {
		matches, _ := pathresolver.ResolvePaths([]string{sessionHome}, sessionFile+".*", config.ConfigExtensions, "ignore")
		for _, match := range matches {
			if isSessionFileExtension(filepath.Ext(match)) && strings.TrimSuffix(filepath.Base(match), filepath.Ext(match)) == sessionFile {
				return match
			}
		}
	}
	return sessionFile
}

// socketName name of the server socket
const socketName = "server.sock"

// DefaultSocketPath socket path used by the server and the relay. It doesn't depend
// on the session, as the session is selected (and can be switched) in the server's
// terminal. Use a custom socket path to run several servers at the same time.
// Shared directories such as /tmp are never used, so another user can't create the
// socket first.
func DefaultSocketPath(sessionHome string) string {
	candidates := []string{filepath.Join(SocketDir(sessionHome), socketName)}
	if cacheDir, err := os.UserCacheDir(); err == nil {
		candidates = append(candidates, filepath.Join(cacheDir, "c8y-mcp", socketName))
	}
	if homeDir, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(homeDir, ".c8y-mcp", socketName))
	}
	for _, p := range candidates {
		if len(p) <= maxSocketPathLength {
			return p
		}
	}
	// Listen will report that the path is too long
	return candidates[0]
}

// Listen creates the unix socket, which is only accessible by the current user.
// It fails if another server is already listening on the socket
func Listen(socketPath string) (net.Listener, error) {
	if len(socketPath) > maxSocketPathLength {
		return nil, fmt.Errorf("socket path is too long (%d characters, max %d). Use --socket to set a shorter path. %s", len(socketPath), maxSocketPathLength, socketPath)
	}
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o700); err != nil {
		return nil, fmt.Errorf("could not create socket directory. %w", err)
	}
	if err := CheckSocketDir(filepath.Dir(socketPath)); err != nil {
		return nil, err
	}

	if _, err := os.Stat(socketPath); err == nil {
		if conn, dialErr := net.DialTimeout("unix", socketPath, time.Second); dialErr == nil {
			conn.Close()
			return nil, fmt.Errorf("an MCP server is already running on %s", socketPath)
		}
		// stale socket from a previous run
		if err := os.Remove(socketPath); err != nil {
			return nil, fmt.Errorf("could not remove stale socket. %w", err)
		}
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(socketPath, 0o600); err != nil {
		listener.Close()
		return nil, fmt.Errorf("could not set socket permissions. %w", err)
	}
	return listener, nil
}

// ServeOptions optional callbacks used by Serve
type ServeOptions struct {
	// Authorize is called for each new connection, while the MCP session is already
	// running, so the client doesn't time out while the user is asked. Until it
	// returns, tool calls are held back by Gates. The connection is closed if it
	// returns false. All connections are accepted if nil.
	Authorize func(ctx context.Context, conn net.Conn) bool

	// Gates records the result of Authorize for the tool calls. Required if Authorize is set
	Gates *ConnectionGates

	// OnConnect, OnDisconnect are called when an authorized client connects or disconnects
	OnConnect    func(active int)
	OnDisconnect func(active int)

	// OnRejected is called when a connection is not authorized
	OnRejected func()
}

// Serve accepts connections on the listener and runs a separate MCP session per connection
func Serve(ctx context.Context, server *mcp.Server, listener net.Listener, opts ServeOptions) error {
	if opts.Authorize != nil && opts.Gates == nil {
		return errors.New("gates are required when connections are authorized")
	}
	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	var mu sync.Mutex
	active := 0
	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer conn.Close()

			session, err := server.Connect(ctx, &mcp.IOTransport{Reader: conn, Writer: conn}, nil)
			if err != nil {
				return
			}
			stop := context.AfterFunc(ctx, func() { session.Close() })
			defer stop()

			allowed := true
			if opts.Authorize != nil {
				gt := opts.Gates.add(session)
				defer opts.Gates.remove(session)
				allowed = opts.Authorize(ctx, conn)
				gt.resolve(allowed)
			}
			if !allowed {
				if opts.OnRejected != nil {
					opts.OnRejected()
				}
				session.Close()
				return
			}

			mu.Lock()
			active++
			count := active
			mu.Unlock()
			if opts.OnConnect != nil {
				opts.OnConnect(count)
			}

			_ = session.Wait()
			conn.Close()
			mu.Lock()
			active--
			count = active
			mu.Unlock()
			if opts.OnDisconnect != nil {
				opts.OnDisconnect(count)
			}
		}()
	}
}

// Relay copies data between the stdio of the MCP client and the server socket.
// It returns when the server closes the connection, or when the client closes
// stdin and the server has finished sending its responses.
func Relay(ctx context.Context, conn net.Conn, stdin io.Reader, stdout io.Writer) error {
	fromClient := make(chan error, 1)
	fromServer := make(chan error, 1)

	go func() {
		_, err := io.Copy(conn, stdin)
		// the client closed stdin, so tell the server no more data is coming
		if uc, ok := conn.(*net.UnixConn); ok {
			_ = uc.CloseWrite()
		}
		fromClient <- err
	}()

	go func() {
		_, err := io.Copy(stdout, conn)
		fromServer <- err
	}()

	defer conn.Close()
	select {
	case <-ctx.Done():
		return nil
	case err := <-fromServer:
		// the server closed the connection (e.g. it was stopped)
		return err
	case err := <-fromClient:
		select {
		case serverErr := <-fromServer:
			if err == nil {
				err = serverErr
			}
		case <-ctx.Done():
		case <-time.After(10 * time.Second):
		}
		return err
	}
}

// NewUnavailableServer an MCP server which explains that the c8y MCP server is
// not running. It is used by "c8y mcp connect" so the agent can tell the user
// what to do instead of the connection failing silently.
func NewUnavailableServer(version string, message string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "go-c8y-cli",
		Title:   "Cumulocity (go-c8y-cli)",
		Version: version,
	}, &mcp.ServerOptions{
		Instructions: message,
	})

	handler := func(ctx context.Context, req *mcp.CallToolRequest, _ map[string]any) (*mcp.CallToolResult, any, error) {
		return errorResult("%s", message), nil, nil
	}
	for _, name := range []string{ToolSession, ToolRead, ToolWrite} {
		mcp.AddTool(server, &mcp.Tool{
			Name:        name,
			Description: "Unavailable: " + message,
		}, handler)
	}
	return server
}
